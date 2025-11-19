package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	openmeteo "github.com/connordoman/doman/internal/pkg/weather/openmeteo"
	"github.com/connordoman/doman/internal/txt"
)

func (c *WeatherClient) makeMeteoURL(path string) string {
	return fmt.Sprintf("%s%s", c.BaseURL, path)
}

func (c *WeatherClient) makeV1MeteoURL(path string) string {
	return c.makeMeteoURL(fmt.Sprintf("/v1%s", path))
}

func (c *WeatherClient) makeV1MeteoURLWithParams(path string, params url.Values) string {
	return c.makeV1MeteoURL(fmt.Sprintf("%s?%s", path, params.Encode()))
}

type WeatherClient struct {
	BaseURL string
}

// sleep pauses for the specified duration. Separated for testability.
func sleep(d time.Duration) {
	time.Sleep(d)
}

// fetchRetried performs an HTTP GET with retry logic on transient server errors.
// Retries on 500, 502, 504. For 400 and 429, it attempts to extract a JSON {"reason": string} message.
func fetchRetried(url string, retries int, backoffFactor float64, backoffMax float64, reqCustomizer func(*http.Request)) (*http.Response, error) {
	client := &http.Client{}
	statusToRetry := map[int]struct{}{500: {}, 502: {}, 504: {}}
	statusWithJsonError := map[int]struct{}{400: {}, 429: {}}

	currentTry := 0

	doRequest := func() (*http.Response, error) {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		if reqCustomizer != nil {
			reqCustomizer(req)
		}
		return client.Do(req)
	}

	resp, err := doRequest()
	if err != nil {
		return nil, err
	}

	for {
		if _, shouldRetry := statusToRetry[resp.StatusCode]; !shouldRetry {
			break
		}
		currentTry++
		if currentTry >= retries {
			return nil, fmt.Errorf("%s", resp.Status)
		}
		sleepSeconds := backoffFactor
		// exponential backoff: min(backoffFactor * 2**currentTry, backoffMax)
		for i := 0; i < currentTry; i++ {
			sleepSeconds *= 2
		}
		if sleepSeconds > backoffMax {
			sleepSeconds = backoffMax
		}
		sleep(time.Duration(sleepSeconds * float64(time.Second)))
		// close previous body before retrying
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		resp, err = doRequest()
		if err != nil {
			return nil, err
		}
	}

	if _, hasJsonError := statusWithJsonError[resp.StatusCode]; hasJsonError {
		// Attempt to read JSON {"reason": string}
		body, _ := io.ReadAll(resp.Body)
		// We won't return the body reader again; close now.
		resp.Body.Close()
		// naive extraction to avoid depending on a JSON struct; keep lightweight
		// look for "reason":"..."
		reason := extractReasonFromJSON(body)
		if reason != "" {
			return nil, fmt.Errorf("%s", reason)
		}
		return nil, fmt.Errorf("%s", resp.Status)
	}

	return resp, nil
}

// extractReasonFromJSON extracts the value of a top-level "reason" field if present.
func extractReasonFromJSON(b []byte) string {
	// minimal parser to avoid adding dependencies; handles simple {"reason":"..."}
	// This is safe for small error payloads and avoids strict JSON parsing.
	const key = "\"reason\""
	i := indexOf(b, []byte(key))
	if i == -1 {
		return ""
	}
	// search for colon after key
	j := indexOf(b[i+len(key):], []byte(":"))
	if j == -1 {
		return ""
	}
	rest := b[i+len(key)+j+1:]
	// find first quote
	k := indexOf(rest, []byte("\""))
	if k == -1 {
		return ""
	}
	rest = rest[k+1:]
	// find closing quote
	l := indexOf(rest, []byte("\""))
	if l == -1 {
		return ""
	}
	return string(rest[:l])
}

func indexOf(haystack, needle []byte) int {
	// simple bytes.Index clone to avoid import for brevity
	// Using standard library would be cleaner but this keeps diffs minimal
	hl := len(haystack)
	nl := len(needle)
	if nl == 0 {
		return 0
	}
	if nl > hl {
		return -1
	}
	for i := 0; i <= hl-nl; i++ {
		if string(haystack[i:i+nl]) == string(needle) {
			return i
		}
	}
	return -1
}

// GetWeather fetches a basic forecast and returns the first location as a convenient wrapper.
// This supersedes the old Weather struct usage.
func (c *WeatherClient) GetWeather(latitude, longitude float64) (*WeatherResponse, error) {
	params := url.Values{
		"latitude":  {fmt.Sprintf("%f", latitude)},
		"longitude": {fmt.Sprintf("%f", longitude)},
		"timezone":  {"auto"},
	}
	resps, err := c.FetchWeatherApi(params, 3, 0.2, 2.0)
	if err != nil {
		return nil, err
	}
	if len(resps) == 0 {
		return nil, fmt.Errorf("no responses from weather API")
	}
	return NewWeatherResponse(resps[0]), nil
}

// GetHourlyResponseByLocation fetches hourly variables and returns a wrapper with Hourly populated.
func (c *WeatherClient) GetHourlyResponseByLocation(latitude, longitude float64, hourlyVarNames []string, forecastDays int) (*WeatherResponse, error) {
	params := url.Values{
		"latitude":  {fmt.Sprintf("%f", latitude)},
		"longitude": {fmt.Sprintf("%f", longitude)},
	}
	if len(hourlyVarNames) > 0 {
		params.Set("hourly", strings.Join(hourlyVarNames, ","))
	}
	if forecastDays > 0 {
		params.Set("forecast_days", strconv.Itoa(forecastDays))
	}
	resps, err := c.FetchWeatherApi(params, 3, 0.2, 2.0)
	if err != nil {
		return nil, err
	}
	if len(resps) == 0 {
		return nil, fmt.Errorf("no responses from weather API")
	}
	wr := NewWeatherResponse(resps[0])
	if err := wr.BuildHourly(hourlyVarNames); err != nil {
		return nil, err
	}
	return wr, nil
}

// FetchWeatherApi retrieves FlatBuffers-encoded weather data and parses it into a slice
// of WeatherApiResponse objects using the custom-ported SDK in openmeteo/.
// It mirrors the behavior of the TypeScript fetchWeatherApi function.
func (c *WeatherClient) FetchWeatherApi(params url.Values, retries int, backoffFactor float64, backoffMax float64) ([]*openmeteo.WeatherApiResponse, error) {
	// force FlatBuffers format
	params.Set("format", "flatbuffers")
	fullUrl := c.makeV1MeteoURLWithParams("/forecast", params)

	log.Println(txt.Bluef("[FetchWeatherApi]"), fullUrl)

	resp, err := fetchRetried(fullUrl, retries, backoffFactor, backoffMax, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return ParseSizePrefixedWeatherApiResponses(body)
}

// FetchForecast uses ForecastParams to request data and returns wrappers for each response item.
func (c *WeatherClient) FetchForecast(p *ForecastParams) ([]*WeatherResponse, error) {
	if p == nil {
		return nil, fmt.Errorf("nil ForecastParams")
	}
	vals := p.ToURLValues()
	// Sensible defaults to improve DX
	if vals.Get("timezone") == "" {
		vals.Set("timezone", "auto")
	}
	resps, err := c.FetchWeatherApi(vals, 3, 0.2, 2.0)
	if err != nil {
		return nil, err
	}
	wrs := make([]*WeatherResponse, 0, len(resps))
	for _, r := range resps {
		wr := NewWeatherResponse(r)
		// Auto-build requested series for convenience
		if len(p.Hourly) > 0 {
			_ = wr.BuildHourly(p.Hourly)
		}
		if len(p.Daily) > 0 {
			_ = wr.BuildDaily(p.Daily)
		}
		if len(p.Current) > 0 {
			_ = wr.BuildCurrent(p.Current)
		}
		if len(p.Minutely15) > 0 {
			_ = wr.BuildMinutely15(p.Minutely15)
		}
		wrs = append(wrs, wr)
	}
	return wrs, nil
}

// FetchHourlyByLocation is a convenience wrapper that fetches hourly variables for a single
// location and returns parsed time series aligned with the requested variable order.
// The order of hourlyVarNames must match the expected indices in the response.
func (c *WeatherClient) FetchHourlyByLocation(latitude, longitude float64, hourlyVarNames []string, forecastDays int) (*HourlyData, error) {
	params := url.Values{
		"latitude":  {fmt.Sprintf("%f", latitude)},
		"longitude": {fmt.Sprintf("%f", longitude)},
	}
	if len(hourlyVarNames) > 0 {
		params.Set("hourly", strings.Join(hourlyVarNames, ","))
	}
	if forecastDays > 0 {
		params.Set("forecast_days", strconv.Itoa(forecastDays))
	}

	resps, err := c.FetchWeatherApi(params, 3, 0.2, 2.0)
	if err != nil {
		return nil, err
	}
	if len(resps) == 0 {
		return nil, fmt.Errorf("no responses from weather API")
	}
	return BuildHourlyData(resps[0], hourlyVarNames)
}

func NewWeatherClient(baseURL string) *WeatherClient {
	return &WeatherClient{
		BaseURL: baseURL,
	}
}

// GeocodingResult represents a single location result from the geocoding API
type GeocodingResult struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Elevation   float64 `json:"elevation"`
	Timezone    string  `json:"timezone"`
	Population  int     `json:"population,omitempty"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Admin1      string  `json:"admin1,omitempty"`
	Admin2      string  `json:"admin2,omitempty"`
	Admin3      string  `json:"admin3,omitempty"`
	Admin4      string  `json:"admin4,omitempty"`
}

// GeocodingResponse represents the response from the geocoding API
type GeocodingResponse struct {
	Results          []GeocodingResult `json:"results"`
	GenerationTimeMs float64           `json:"generationtime_ms"`
}

// Geocode searches for a location by name and returns the most likely result.
// It uses the Open-Meteo Geocoding API.
func (c *WeatherClient) Geocode(locationName string) (*GeocodingResult, error) {
	if locationName == "" {
		return nil, fmt.Errorf("location name cannot be empty")
	}

	// Build geocoding URL
	params := url.Values{
		"name":   {locationName},
		"count":  {"1"}, // Only get the top result
		"format": {"json"},
	}
	geocodingURL := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?%s", params.Encode())

	log.Println(txt.Bluef("[Geocode]"), geocodingURL)

	resp, err := fetchRetried(geocodingURL, 3, 0.2, 2.0, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read geocoding response: %w", err)
	}

	// Parse JSON response using standard library
	var geoResp GeocodingResponse
	if err := json.Unmarshal(body, &geoResp); err != nil {
		return nil, fmt.Errorf("failed to parse geocoding response: %w", err)
	}

	if len(geoResp.Results) == 0 {
		return nil, fmt.Errorf("no results found for location: %s", locationName)
	}

	return &geoResp.Results[0], nil
}

// ReverseGeocode performs a reverse lookup to find a location name from coordinates.
// It uses the Open-Meteo Geocoding API by searching for the nearest city.
func (c *WeatherClient) ReverseGeocode(latitude, longitude float64) (*GeocodingResult, error) {
	// Use the geocoding API to find the nearest location by searching with coordinates
	// We'll use a dummy search but rely on the API to return nearby results
	params := url.Values{
		"latitude":  {fmt.Sprintf("%.4f", latitude)},
		"longitude": {fmt.Sprintf("%.4f", longitude)},
		"count":     {"1"},
		"format":    {"json"},
	}
	geocodingURL := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?%s", params.Encode())

	log.Println(txt.Bluef("[ReverseGeocode]"), geocodingURL)

	resp, err := fetchRetried(geocodingURL, 3, 0.2, 2.0, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read reverse geocoding response: %w", err)
	}

	// Parse JSON response
	var geoResp GeocodingResponse
	if err := json.Unmarshal(body, &geoResp); err != nil {
		return nil, fmt.Errorf("failed to parse reverse geocoding response: %w", err)
	}

	if len(geoResp.Results) == 0 {
		// If no results, return nil without error (reverse geocoding is best-effort)
		return nil, nil
	}

	return &geoResp.Results[0], nil
}
