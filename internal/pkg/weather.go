package pkg

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

const (
	weatherApiUrl = "https://api.openweathermap.org/data/2.5/weather"
)

// WeatherResponse represents the complete OpenWeather API response
type WeatherResponse struct {
	Base       string    `json:"base"`
	Clouds     Clouds    `json:"clouds"`
	Cod        int       `json:"cod"`
	Coord      Coord     `json:"coord"`
	Dt         int64     `json:"dt"`
	ID         int       `json:"id"`
	Main       Main      `json:"main"`
	Name       string    `json:"name"`
	Sys        Sys       `json:"sys"`
	Timezone   int       `json:"timezone"`
	Visibility int       `json:"visibility"`
	Weather    []Weather `json:"weather"`
	Wind       Wind      `json:"wind"`
}

// Clouds represents cloud coverage data
type Clouds struct {
	All int `json:"all"`
}

// Coord represents geographical coordinates
type Coord struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// Main represents main weather data
type Main struct {
	FeelsLike float64 `json:"feels_like"`
	GrndLevel int     `json:"grnd_level"`
	Humidity  int     `json:"humidity"`
	Pressure  int     `json:"pressure"`
	SeaLevel  int     `json:"sea_level"`
	Temp      float64 `json:"temp"`
	TempMax   float64 `json:"temp_max"`
	TempMin   float64 `json:"temp_min"`
}

// Sys represents system data
type Sys struct {
	Country string `json:"country"`
	ID      int    `json:"id"`
	Sunrise int64  `json:"sunrise"`
	Sunset  int64  `json:"sunset"`
	Type    int    `json:"type"`
}

// Weather represents weather condition data
type Weather struct {
	Description string `json:"description"`
	Icon        string `json:"icon"`
	ID          int    `json:"id"`
	Main        string `json:"main"`
}

// Wind represents wind data
type Wind struct {
	Deg   int     `json:"deg"`
	Speed float64 `json:"speed"`
}

// Simple latitude-longitude struct
type Location struct {
	Lat float64 `json:"latitude"`
	Lon float64 `json:"longitude"`
}

func getWeatherApiUrl(lat, lon float64) string {
	apiKey := viper.GetString("weather.openweathermap.api_key")
	if apiKey == "" {
		log.Fatal("API Key for OpenWeatherMap is not set. Please set it in your configuration.")
	}
	return fmt.Sprintf("%s?lat=%f&lon=%f&units=metric&appid=%s", weatherApiUrl, lat, lon, apiKey)
}

func getCurrentLocationLatLon() (*Location, error) {
	ipInfo, err := GetIPInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get current location: %w", err)
	}

	if ipInfo.Loc == "" {
		return nil, fmt.Errorf("location data not found in IP info response")
	}

	locParts := strings.Split(ipInfo.Loc, ",")
	if len(locParts) != 2 {
		return nil, fmt.Errorf("invalid location format in IP info response: %s", ipInfo.Loc)
	}

	lat, err := strconv.ParseFloat(locParts[0], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse latitude: %w", err)
	}

	lon, err := strconv.ParseFloat(locParts[1], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse longitude: %w", err)
	}

	return &Location{Lat: lat, Lon: lon}, nil
}

func GetWeatherCurrentLocation() (*WeatherResponse, error) {
	location, err := getCurrentLocationLatLon()
	if err != nil {
		return nil, fmt.Errorf("failed to get current location: %w", err)
	}

	return GetWeather(location.Lat, location.Lon)
}

func GetWeather(lat, lon float64) (*WeatherResponse, error) {
	apiUrl := getWeatherApiUrl(lat, lon)
	resp, err := http.Get(apiUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching weather data: %s", resp.Status)
	}

	var weatherResponse WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResponse); err != nil {
		return nil, fmt.Errorf("failed to decode weather response: %w", err)
	}

	return &weatherResponse, nil
}
