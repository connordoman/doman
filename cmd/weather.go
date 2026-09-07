package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"time"

	weatherCmd "doman.sh/cmd/weather"
	"doman.sh/internal/pkg/weather"
	"doman.sh/internal/txt"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var WeatherCommand = &cobra.Command{
	Use:   "weather [location]",
	Short: "Get the weather for a given location",
	Long: `Get the weather for a given location.

You can specify a location by:
  - Name: weather "Lake Louise"
  - Coordinates: weather --lat 51.4254 --lon -116.1773

If both a location name and coordinates are provided, coordinates take precedence.`,
	RunE: runWeatherCommand,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if verbose, err := cmd.Flags().GetBool("verbose"); err == nil && verbose {
			log.SetOutput(os.Stdout)
		} else {
			log.SetOutput(io.Discard)
		}

		return nil
	},
}

const (
	defaultLatitude  = 49.2997
	defaultLongitude = -121.7859
)

func init() {
	WeatherCommand.Flags().Float64("lat", defaultLatitude, "Latitude of the location")
	WeatherCommand.Flags().Float64("lon", defaultLongitude, "Longitude of the location")

	WeatherCommand.AddCommand(weatherCmd.TempCommand)
}

func runWeatherCommand(cmd *cobra.Command, args []string) error {
	latFlag, _ := cmd.Flags().GetFloat64("lat")
	lonFlag, _ := cmd.Flags().GetFloat64("lon")

	localTime := time.Now()

	client := weather.NewWeatherClient(weather.BaseURL)

	// Determine if user provided custom coordinates
	latChanged := cmd.Flags().Changed("lat")
	lonChanged := cmd.Flags().Changed("lon")
	useCoordinates := latChanged || lonChanged

	var locationResult *weather.GeocodingResult

	// If no coordinates provided and there's a location argument, geocode it
	if !useCoordinates && len(args) > 0 {
		locationName := args[0]
		log.Printf("Geocoding location: %s\n", locationName)

		result, err := client.Geocode(locationName)
		if err != nil {
			return fmt.Errorf("failed to geocode location '%s': %w", locationName, err)
		}

		latFlag = result.Latitude
		lonFlag = result.Longitude
		locationResult = result
	} else if useCoordinates {
		// If coordinates were provided, try reverse geocoding
		log.Printf("Reverse geocoding coordinates: %.4f, %.4f\n", latFlag, lonFlag)
		result, err := client.ReverseGeocode(latFlag, lonFlag)
		if err == nil && result != nil {
			locationResult = result
		}
		// If reverse geocoding fails, we'll just show coordinates
	}

	// Display the location if we have it
	if locationResult != nil {
		locationText := fmt.Sprintf("%s, %s (%.4f°N, %.4f°E)", locationResult.Name, locationResult.Country, locationResult.Latitude, locationResult.Longitude)
		fmt.Println(txt.Underlinef("%s", locationText))
	}

	hourlyParams := []string{"temperature_2m", "precipitation", "precipitation_probability", "weather_code", "cloud_cover", "is_day"}
	params := &weather.ForecastParams{
		Latitude:     latFlag,
		Longitude:    lonFlag,
		Hourly:       hourlyParams,
		Timezone:     "auto",
		ForecastDays: 1,
	}
	responses, err := client.FetchForecast(params)
	if err != nil {
		return err
	}

	fmt.Println(lipgloss.NewStyle().Underline(true).Render(localTime.Format("Monday, 2 January 2006")))
	fmt.Println()

	for _, wr := range responses {
		fmt.Println(wr.Summary())
		fmt.Println()

		if wr.Hourly != nil {
			// sort the time map by time
			sortedTimeMap := make(map[time.Time]map[string]float32)
			for time, values := range wr.Hourly.TimeMap {
				sortedTimeMap[time] = values
			}
			sortedTimes := make([]time.Time, 0, len(sortedTimeMap))
			for time := range sortedTimeMap {
				sortedTimes = append(sortedTimes, time)
			}
			sort.Slice(sortedTimes, func(i, j int) bool { return sortedTimes[i].Before(sortedTimes[j]) })
			for _, time := range sortedTimes {
				values := sortedTimeMap[time]
				cond := weather.DetectConditionFromSeries(values)

				// emojiCol := txt.PadRightWidth(cond.Emoji, 3)
				emojiCol := cond.Emoji
				// tempCol := txt.PadLeftWidth(fmt.Sprintf("%0.0f°C", cond.TemperatureC), 4)
				// tempCol := fmt.Sprintf("%0.0f°C", cond.TemperatureC)
				tempCol := txt.ColorizePercent(fmt.Sprintf("%0.0f°C", cond.TemperatureC), weather.TemperaturePercent(cond.TemperatureC)*100, txt.TemperatureGradient())

				prob := ""
				if cond.PrecipitationProbability >= 0 {
					prob = lipgloss.NewStyle().Foreground(lipgloss.Color("#57534e")).Render(fmt.Sprintf("(%0.0f%%)", cond.PrecipitationProbability))

					// 	raw := fmt.Sprintf("(%0.0f%%)", cond.PrecipitationProbability)
					// 	colored := txt.ColorizePercent(raw, float64(cond.PrecipitationProbability), txt.GreenYellowRedGradient())
					// 	prob = colored
				}

				// Use lipgloss styles with fixed widths to spread values evenly
				emojiStyle := lipgloss.NewStyle().Width(4).Align(lipgloss.Left)
				tempStyle := lipgloss.NewStyle().Width(6).Align(lipgloss.Center)
				probStyle := lipgloss.NewStyle().Width(6).Align(lipgloss.Left)

				entry := lipgloss.JoinHorizontal(lipgloss.Top,
					tempStyle.Render(tempCol),
					probStyle.Render(prob),
					emojiStyle.Render(emojiCol),
				)

				if time.Hour() == localTime.Hour() {
					hourHighlighted := lipgloss.NewStyle().Foreground(lipgloss.Color("#2563eb")).Render(fmt.Sprintf("%05s", localTime.Format("15:04")))
					fmt.Printf("%s %s\n", hourHighlighted, entry)
				} else {
					fmt.Printf("%02d:00 %s\n", time.Hour(), entry)
				}
			}
		}
	}

	return nil
}
