package weather

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	openmeteo "github.com/connordoman/doman/internal/pkg/weather/openmeteo"
	flatbuffers "github.com/google/flatbuffers/go"
)

const BaseURL = "https://api.open-meteo.com"

// ParseSizePrefixedWeatherApiResponses parses a concatenated sequence of
// size-prefixed FlatBuffers WeatherApiResponse messages.
// Each message is prefixed with a 32-bit little-endian length.
func ParseSizePrefixedWeatherApiResponses(data []byte) ([]*openmeteo.WeatherApiResponse, error) {
	if len(data) == 0 {
		return nil, errors.New("empty response body")
	}

	var results []*openmeteo.WeatherApiResponse
	pos := 0
	for pos < len(data) {
		if pos+4 > len(data) {
			return nil, errors.New("truncated size prefix in FlatBuffers stream")
		}
		segLen := int(binary.LittleEndian.Uint32(data[pos : pos+4]))
		pos += 4
		if segLen < 0 || pos+segLen > len(data) {
			return nil, errors.New("invalid segment length in FlatBuffers stream")
		}

		// For size-prefixed messages, the FlatBuffers table starts immediately after the 4-byte prefix.
		resp := openmeteo.GetRootAsWeatherApiResponse(data, flatbuffers.UOffsetT(pos))
		results = append(results, resp)
		pos += segLen
	}
	return results, nil
}

// HourlyData is a convenience structure for hourly time series aligned
// to requested variable ordering.
type HourlyData struct {
	Time      []time.Time
	Variables map[string][]float32
	TimeMap   map[time.Time]map[string]float32
}

func (h *HourlyData) String() string {
	var sb strings.Builder
	sb.WriteString("HourlyData:\n")
	for v, values := range h.Variables {
		sb.WriteString(fmt.Sprintf("  %s: %v\n", v, values))
	}
	return sb.String()
}

func (h *HourlyData) BuildTimeMap() {
	h.TimeMap = make(map[time.Time]map[string]float32)
	for i, time := range h.Time {
		h.TimeMap[time] = make(map[string]float32)
		for v, values := range h.Variables {
			h.TimeMap[time][v] = values[i]
		}
	}
}

func (h *HourlyData) StringTimeMap() string {
	var sb strings.Builder
	// Build lazily for better DX if not yet constructed
	if h.TimeMap == nil {
		h.BuildTimeMap()
	}
	sb.WriteString("TimeMap:\n")
	for time, values := range h.TimeMap {
		sb.WriteString(fmt.Sprintf("  %s: %v\n", time, values))
	}

	return sb.String()
}

func BuildHourlyData(resp *openmeteo.WeatherApiResponse, hourlyVarNames []string) (*HourlyData, error) {
	if resp == nil {
		return nil, errors.New("nil WeatherApiResponse")
	}
	var hourly openmeteo.VariablesWithTime
	if resp.Hourly(&hourly) == nil {
		return nil, errors.New("missing hourly data")
	}
	t0 := hourly.Time()
	tEnd := hourly.TimeEnd()
	interval := hourly.Interval()
	if interval <= 0 {
		return nil, errors.New("invalid hourly interval")
	}
	n := int((tEnd - t0) / int64(interval))
	times := make([]time.Time, n)
	// Prefer IANA location if provided for correct DST handling; fallback to fixed offset
	var loc *time.Location
	tzName := string(resp.Timezone())
	if tzName != "" {
		if l, err := time.LoadLocation(tzName); err == nil {
			loc = l
		}
	}
	if loc == nil {
		name := string(resp.TimezoneAbbreviation())
		if name == "" {
			if tzName != "" {
				name = tzName
			} else {
				name = "UTC"
			}
		}
		loc = time.FixedZone(name, int(resp.UtcOffsetSeconds()))
	}
	for i := 0; i < n; i++ {
		epoch := t0 + int64(i)*int64(interval)
		times[i] = time.Unix(epoch, 0).In(loc)
	}

	vars := make(map[string][]float32, len(hourlyVarNames))
	for idx, name := range hourlyVarNames {
		var vw openmeteo.VariableWithValues
		if hourly.Variables(&vw, idx) == nil {
			return nil, errors.New("variable index out of range for hourly data")
		}
		l := vw.ValuesLength()
		arr := make([]float32, l)
		for i := 0; i < l; i++ {
			arr[i] = vw.Values(i)
		}
		vars[name] = arr
	}

	return &HourlyData{
		Time:      times,
		Variables: vars,
	}, nil
}

// TimeSeriesData is a generic time-indexed series used for hourly, daily, current and minutely15.
type TimeSeriesData struct {
	Time      []time.Time
	Variables map[string][]float32
}

func buildTimeSeriesFromVariablesWithTime(vwt *openmeteo.VariablesWithTime, utcOffsetSeconds int32, varNames []string) (*TimeSeriesData, error) {
	if vwt == nil {
		return nil, errors.New("nil VariablesWithTime")
	}
	t0 := vwt.Time()
	tEnd := vwt.TimeEnd()
	interval := vwt.Interval()
	if interval <= 0 {
		return nil, errors.New("invalid interval for time series")
	}
	n := int((tEnd - t0) / int64(interval))
	offset := int64(utcOffsetSeconds)
	times := make([]time.Time, n)
	for i := 0; i < n; i++ {
		epoch := t0 + int64(i)*int64(interval) + offset
		times[i] = time.Unix(epoch, 0).UTC()
	}
	vars := make(map[string][]float32, len(varNames))
	for idx, name := range varNames {
		var vw openmeteo.VariableWithValues
		if vwt.Variables(&vw, idx) == nil {
			return nil, errors.New("variable index out of range for time series")
		}
		l := vw.ValuesLength()
		arr := make([]float32, l)
		for i := 0; i < l; i++ {
			arr[i] = vw.Values(i)
		}
		vars[name] = arr
	}
	return &TimeSeriesData{Time: times, Variables: vars}, nil
}

// MonthlyData represents aggregated per-month variable arrays with metadata.
type MonthlyData struct {
	Year      int
	Month     int
	Count     int
	Variables map[string][]float32
}

func BuildMonthlyData(resp *openmeteo.WeatherApiResponse, monthlyVarNames []string) (*MonthlyData, error) {
	if resp == nil {
		return nil, errors.New("nil WeatherApiResponse")
	}
	var m openmeteo.VariablesWithMonth
	if resp.Monthly(&m) == nil {
		return nil, errors.New("missing monthly data")
	}
	vars := make(map[string][]float32, len(monthlyVarNames))
	for idx, name := range monthlyVarNames {
		var vw openmeteo.VariableWithValues
		if m.Variables(&vw, idx) == nil {
			return nil, errors.New("variable index out of range for monthly data")
		}
		l := vw.ValuesLength()
		arr := make([]float32, l)
		for i := 0; i < l; i++ {
			arr[i] = vw.Values(i)
		}
		vars[name] = arr
	}
	return &MonthlyData{
		Year:      int(m.Year()),
		Month:     int(m.Month()),
		Count:     int(m.Count()),
		Variables: vars,
	}, nil
}

// WeatherResponse is a convenience wrapper around openmeteo.WeatherApiResponse
// providing typed helpers for common fields and derived values.
type WeatherResponse struct {
	raw        *openmeteo.WeatherApiResponse
	Hourly     *HourlyData
	Daily      *TimeSeriesData
	Current    *TimeSeriesData
	Minutely15 *TimeSeriesData
	Monthly    *MonthlyData
}

func NewWeatherResponse(resp *openmeteo.WeatherApiResponse) *WeatherResponse {
	return &WeatherResponse{raw: resp}
}

func (w *WeatherResponse) Latitude() float64            { return float64(w.raw.Latitude()) }
func (w *WeatherResponse) Longitude() float64           { return float64(w.raw.Longitude()) }
func (w *WeatherResponse) Elevation() float64           { return float64(w.raw.Elevation()) }
func (w *WeatherResponse) UtcOffsetSeconds() int        { return int(w.raw.UtcOffsetSeconds()) }
func (w *WeatherResponse) Timezone() string             { return string(w.raw.Timezone()) }
func (w *WeatherResponse) TimezoneAbbreviation() string { return string(w.raw.TimezoneAbbreviation()) }

// BuildHourly populates the Hourly field using the provided variable names in the
// same order as requested from the API (must match indices).
func (w *WeatherResponse) BuildHourly(hourlyVarNames []string) error {
	h, err := BuildHourlyData(w.raw, hourlyVarNames)
	if err != nil {
		return err
	}
	w.Hourly = h
	// Ensure convenient accessors (like StringTimeMap) work out of the box
	w.Hourly.BuildTimeMap()
	return nil
}

func (w *WeatherResponse) BuildDaily(dailyVarNames []string) error {
	var v openmeteo.VariablesWithTime
	if w.raw.Daily(&v) == nil {
		return errors.New("missing daily data")
	}
	ts, err := buildTimeSeriesFromVariablesWithTime(&v, w.raw.UtcOffsetSeconds(), dailyVarNames)
	if err != nil {
		return err
	}
	w.Daily = ts
	return nil
}

func (w *WeatherResponse) BuildCurrent(currentVarNames []string) error {
	var v openmeteo.VariablesWithTime
	if w.raw.Current(&v) == nil {
		return errors.New("missing current data")
	}
	ts, err := buildTimeSeriesFromVariablesWithTime(&v, w.raw.UtcOffsetSeconds(), currentVarNames)
	if err != nil {
		return err
	}
	w.Current = ts
	return nil
}

func (w *WeatherResponse) BuildMinutely15(minutelyVarNames []string) error {
	var v openmeteo.VariablesWithTime
	if w.raw.Minutely15(&v) == nil {
		return errors.New("missing minutely15 data")
	}
	ts, err := buildTimeSeriesFromVariablesWithTime(&v, w.raw.UtcOffsetSeconds(), minutelyVarNames)
	if err != nil {
		return err
	}
	w.Minutely15 = ts
	return nil
}

func (w *WeatherResponse) BuildMonthly(monthlyVarNames []string) error {
	m, err := BuildMonthlyData(w.raw, monthlyVarNames)
	if err != nil {
		return err
	}
	w.Monthly = m
	return nil
}

// Summary produces a short single-line human-readable description of the location context.
func (w *WeatherResponse) Summary() string {
	// Format UTC offset as ±HH:MM instead of raw seconds
	offsetSeconds := w.UtcOffsetSeconds()
	offsetHours := offsetSeconds / 3600
	offsetMinutes := (offsetSeconds % 3600) / 60
	if offsetMinutes < 0 {
		offsetMinutes = -offsetMinutes
	}

	var offsetStr string
	if offsetSeconds >= 0 {
		offsetStr = fmt.Sprintf("+%02d:%02d", offsetHours, offsetMinutes)
	} else {
		offsetStr = fmt.Sprintf("-%02d:%02d", -offsetHours, offsetMinutes)
	}

	// Include timezone name if available
	tzName := w.Timezone()
	var tzDisplay string
	if tzName != "" {
		tzDisplay = fmt.Sprintf("%s (UTC%s)", tzName, offsetStr)
	} else {
		tzDisplay = fmt.Sprintf("UTC%s", offsetStr)
	}

	return fmt.Sprintf("Coordinates: %.4f°N %.4f°E | Elevation: %.0fm | %s", w.Latitude(), w.Longitude(), w.Elevation(), tzDisplay)
}

// ForecastParams captures supported API parameters and can be transformed into URL values.
type ForecastParams struct {
	Latitude          float64
	Longitude         float64
	Timezone          string
	StartDate         string
	EndDate           string
	ForecastDays      int
	PastDays          int
	ForecastHours     int
	PastHours         int
	Models            string
	TemperatureUnit   string
	WindSpeedUnit     string
	PrecipitationUnit string
	Timeformat        string
	CellSelection     string
	CurrentWeather    bool
	Current           []string
	Hourly            []string
	Daily             []string
	Minutely15        []string
	Additional        map[string]string
}

func (p *ForecastParams) ToURLValues() url.Values {
	v := url.Values{}
	if p.Latitude != 0 {
		v.Set("latitude", fmt.Sprintf("%f", p.Latitude))
	}
	if p.Longitude != 0 {
		v.Set("longitude", fmt.Sprintf("%f", p.Longitude))
	}
	if p.Timezone != "" {
		v.Set("timezone", p.Timezone)
	}
	if p.StartDate != "" {
		v.Set("start_date", p.StartDate)
	}
	if p.EndDate != "" {
		v.Set("end_date", p.EndDate)
	}
	if p.ForecastDays > 0 {
		v.Set("forecast_days", strconv.Itoa(p.ForecastDays))
	}
	if p.PastDays > 0 {
		v.Set("past_days", strconv.Itoa(p.PastDays))
	}
	if p.ForecastHours > 0 {
		v.Set("forecast_hours", strconv.Itoa(p.ForecastHours))
	}
	if p.PastHours > 0 {
		v.Set("past_hours", strconv.Itoa(p.PastHours))
	}
	if p.Models != "" {
		v.Set("models", p.Models)
	}
	if p.TemperatureUnit != "" {
		v.Set("temperature_unit", p.TemperatureUnit)
	}
	if p.WindSpeedUnit != "" {
		v.Set("wind_speed_unit", p.WindSpeedUnit)
	}
	if p.PrecipitationUnit != "" {
		v.Set("precipitation_unit", p.PrecipitationUnit)
	}
	if p.Timeformat != "" {
		v.Set("timeformat", p.Timeformat)
	}
	if p.CellSelection != "" {
		v.Set("cell_selection", p.CellSelection)
	}
	if len(p.Current) > 0 {
		v.Set("current", strings.Join(p.Current, ","))
	}
	if len(p.Hourly) > 0 {
		v.Set("hourly", strings.Join(p.Hourly, ","))
	}
	if len(p.Daily) > 0 {
		v.Set("daily", strings.Join(p.Daily, ","))
	}
	if len(p.Minutely15) > 0 {
		v.Set("minutely_15", strings.Join(p.Minutely15, ","))
	}
	if p.CurrentWeather {
		v.Set("current_weather", "true")
	}
	for k, val := range p.Additional {
		if k == "" || val == "" {
			continue
		}
		v.Set(k, val)
	}
	return v
}
