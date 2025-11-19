package weather

import "fmt"

// ConditionType represents a coarse weather condition classification.
type ConditionType string

const (
	ConditionClear            ConditionType = "clear"
	ConditionClearNight       ConditionType = "clear_night"
	ConditionSunny            ConditionType = "sunny"
	ConditionPartlyCloudy     ConditionType = "partly_cloudy"
	ConditionCloudy           ConditionType = "cloudy"
	ConditionOvercast         ConditionType = "overcast"
	ConditionFog              ConditionType = "fog"
	ConditionDrizzle          ConditionType = "drizzle"
	ConditionRain             ConditionType = "rain"
	ConditionRainShowers      ConditionType = "rain_showers"
	ConditionFreezingRain     ConditionType = "freezing_rain"
	ConditionSnow             ConditionType = "snow"
	ConditionSnowShowers      ConditionType = "snow_showers"
	ConditionSnowGrains       ConditionType = "snow_grains"
	ConditionThunderstorm     ConditionType = "thunderstorm"
	ConditionThunderstormHail ConditionType = "thunderstorm_hail"
	ConditionMixed            ConditionType = "mixed"
	ConditionPossibleWet      ConditionType = "possible_wet"
	ConditionLikelyWet        ConditionType = "likely_wet"
)

// Condition describes detected conditions for a single timestep.
type Condition struct {
	Type                     ConditionType
	Emoji                    string
	Summary                  string
	TemperatureC             float32
	PrecipitationMmPerHour   float32
	PrecipitationProbability float32 // 0-100; -1 if unavailable
	CloudCoverPercent        float32 // 0-100; -1 if unavailable
	WeatherCode              int     // -1 if unavailable
	IsDay                    bool
}

// DetectConditionFromValues uses WMO weather_code primarily and refines with precipitation,
// precipitation_probability, cloud_cover, and is_day.
// temperature_2m is used to infer phase when overriding based on precipitation.
func DetectConditionFromValues(temperatureC, precipitationMmPerHour, precipitationProbability float32, weatherCode int, cloudCoverPercent float32, isDay bool) Condition {
	// Establish a baseline condition from WMO weather code
	baseType, baseEmoji, baseSummary := baselineFromWmo(weatherCode, isDay)

	cond := Condition{
		Type:                     baseType,
		Emoji:                    baseEmoji,
		Summary:                  baseSummary,
		TemperatureC:             temperatureC,
		PrecipitationMmPerHour:   precipitationMmPerHour,
		PrecipitationProbability: precipitationProbability,
		CloudCoverPercent:        cloudCoverPercent,
		WeatherCode:              weatherCode,
		IsDay:                    isDay,
	}

	// If baseline is one of the cloudiness/clear states, refine using cloud cover thresholds
	if baseType == ConditionSunny || baseType == ConditionClear || baseType == ConditionClearNight || baseType == ConditionPartlyCloudy || baseType == ConditionCloudy || baseType == ConditionOvercast {
		if cloudCoverPercent >= 0 { // available
			if !isDay && (baseType == ConditionSunny || baseType == ConditionClear) {
				cond.Type = ConditionClearNight
				cond.Emoji = "🌙"
				cond.Summary = "clear night"
			} else if cloudCoverPercent < 20 {
				cond.Type = ternary(isDay, ConditionSunny, ConditionClear).(ConditionType)
				cond.Emoji = ternary(isDay, "☀️", "🌙").(string)
				cond.Summary = ternary(isDay, "sunny", "clear").(string)
			} else if cloudCoverPercent < 70 {
				cond.Type = ConditionPartlyCloudy
				cond.Emoji = ternary(isDay, "🌤️", "☁️").(string)
				cond.Summary = "partly cloudy"
			} else {
				cond.Type = ConditionOvercast
				cond.Emoji = "☁️"
				cond.Summary = "overcast"
			}
		}
	}

	// If precipitation occurs now and baseline is ambiguous cloudiness, escalate
	if precipitationMmPerHour > 0.1 {
		// Decide phase by temperature
		if temperatureC <= -1.0 {
			cond.Type = ConditionSnow
			cond.Emoji = "❄️"
			cond.Summary = intensityLabel(precipitationMmPerHour) + " snow"
			return cond
		}
		if temperatureC > -1.0 && temperatureC < 2.0 {
			cond.Type = ConditionMixed
			cond.Emoji = "🌨️"
			cond.Summary = intensityLabel(precipitationMmPerHour) + " rain/snow"
			return cond
		}
		// Otherwise, rain or showers depending on baseline code groups
		if isShowerCode(weatherCode) {
			cond.Type = ConditionRainShowers
			cond.Emoji = emojiForRainIntensity(precipitationMmPerHour)
			cond.Summary = intensityLabel(precipitationMmPerHour) + " showers"
			return cond
		}
		cond.Type = ConditionRain
		cond.Emoji = emojiForRainIntensity(precipitationMmPerHour)
		cond.Summary = intensityLabel(precipitationMmPerHour) + " rain"
		return cond
	}

	// No precipitation measured: if probability is high, nudge the summary
	if precipitationProbability >= 70 {
		if cond.Type == ConditionSunny || cond.Type == ConditionClear || cond.Type == ConditionPartlyCloudy || cond.Type == ConditionCloudy {
			cond.Type = ConditionLikelyWet
			cond.Emoji = "🌦️"
			cond.Summary = "likely precipitation"
		}
	} else if precipitationProbability >= 40 {
		if cond.Type == ConditionSunny || cond.Type == ConditionClear || cond.Type == ConditionPartlyCloudy {
			cond.Type = ConditionPossibleWet
			cond.Emoji = "🌦️"
			cond.Summary = "chance of precipitation"
		}
	}

	return cond
}

// DetectConditionFromSeries looks up variables from an HourlyData's TimeMap for a given time key.
// The map keys must include "temperature_2m" and "precipitation"; "precipitation_probability" is optional.
func DetectConditionFromSeries(values map[string]float32) Condition {
	temp := values["temperature_2m"]
	precip := values["precipitation"]
	prob, hasProb := values["precipitation_probability"]
	if !hasProb {
		prob = -1
	}
	cc, hasCC := values["cloud_cover"]
	if !hasCC {
		cc = -1
	}
	wcFloat, hasWC := values["weather_code"]
	wc := -1
	if hasWC {
		wc = int(wcFloat + 0.5)
	}
	isDayF, hasIsDay := values["is_day"]
	isDay := false
	if hasIsDay && isDayF >= 1 {
		isDay = true
	}
	return DetectConditionFromValues(temp, precip, prob, wc, cc, isDay)
}

func intensityLabel(mmPerHour float32) string {
	// Basic intensity thresholds per common convention
	if mmPerHour < 0.5 {
		return "light"
	}
	if mmPerHour < 2.0 {
		return "moderate"
	}
	return "heavy"
}

func emojiForRainIntensity(mmPerHour float32) string {
	if mmPerHour < 0.5 {
		return "🌦️"
	}
	if mmPerHour < 2.0 {
		return "🌧️"
	}
	return "⛈️"
}

// FormatConditionShort returns a compact string for CLI output, e.g. "🌧️ 12°C 1.2mm (80%) - rain".
func FormatConditionShort(c Condition) string {
	t := fmt.Sprintf("%0.0f°C", c.TemperatureC)
	precip := ""
	if c.PrecipitationMmPerHour > 0 {
		precip = fmt.Sprintf(" %0.2fmm", c.PrecipitationMmPerHour)
	}
	prob := ""
	if c.PrecipitationProbability >= 0 {
		prob = fmt.Sprintf(" (%0.0f%%)", c.PrecipitationProbability)
	}
	summary := ""
	if c.Summary != "" {
		summary = fmt.Sprintf(" - %s %s", c.Summary, precip)
	}
	return fmt.Sprintf("%s %4s%6s%s", c.Emoji, t, prob, summary)
}

// baselineFromWmo maps WMO weather codes to a baseline condition.
func baselineFromWmo(code int, isDay bool) (ConditionType, string, string) {
	switch code {
	case 0:
		if isDay {
			return ConditionSunny, "☀️", "sunny"
		}
		return ConditionClearNight, "🌙", "clear night"
	case 1:
		if isDay {
			return ConditionPartlyCloudy, "🌤️", "mainly clear"
		}
		return ConditionPartlyCloudy, "☁️", "mainly clear"
	case 2:
		return ConditionPartlyCloudy, "⛅", "partly cloudy"
	case 3:
		return ConditionOvercast, "☁️", "overcast"
	case 45, 48:
		return ConditionFog, "🌫️", "fog"
	case 51, 53, 55:
		return ConditionDrizzle, "🌦️", "drizzle"
	case 61, 63, 65:
		return ConditionRain, "🌧️", "rain"
	case 66, 67:
		return ConditionFreezingRain, "🧊", "freezing rain"
	case 71, 73, 75:
		return ConditionSnow, "❄️", "snow"
	case 77:
		return ConditionSnowGrains, "❄️", "snow grains"
	case 80, 81, 82:
		return ConditionRainShowers, "🌦️", "rain showers"
	case 85, 86:
		return ConditionSnowShowers, "🌨️", "snow showers"
	case 95:
		return ConditionThunderstorm, "⛈️", "thunderstorm"
	case 96, 99:
		return ConditionThunderstormHail, "⛈️", "thunderstorm with hail"
	default:
		// Unknown or unavailable code: fall back to clear/cloudy heuristic by day/night
		if isDay {
			return ConditionSunny, "☀️", "clear"
		}
		return ConditionClearNight, "🌙", "clear night"
	}
}

func isShowerCode(code int) bool {
	return code == 80 || code == 81 || code == 82
}

// tiny helper for inline conditional
func ternary(cond bool, a, b interface{}) interface{} {
	if cond {
		return a
	}
	return b
}
