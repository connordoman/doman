package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"time"

	"github.com/connordoman/doman/internal/pkg/weather"
	"github.com/connordoman/doman/internal/txt"
	"github.com/spf13/cobra"
)

var WeatherCommand = &cobra.Command{
	Use:   "weather",
	Short: "Get the weather for a given location",
	RunE:  runWeatherCommand,
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

func runWeatherCommand(cmd *cobra.Command, args []string) error {

	client := weather.NewWeatherClient(weather.BaseURL)
	hourlyParams := []string{"temperature_2m", "precipitation", "precipitation_probability", "weather_code", "cloud_cover", "is_day"}
	params := &weather.ForecastParams{
		Latitude:     defaultLatitude,
		Longitude:    defaultLongitude,
		Hourly:       hourlyParams,
		Timezone:     "auto",
		ForecastDays: 1,
	}
	responses, err := client.FetchForecast(params)
	if err != nil {
		return err
	}

	fmt.Println(time.Now().Format("Monday, 2 January 2006"))

	for _, wr := range responses {
		fmt.Println(wr.Summary())

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

				emojiCol := txt.PadRightWidth(cond.Emoji, 2)
				tempCol := txt.PadLeftWidth(fmt.Sprintf("%0.0f°C", cond.TemperatureC), 4)
				prob := ""
				if cond.PrecipitationProbability >= 0 {
					// Colorize first, then pad using ANSI-aware width
					raw := fmt.Sprintf("(%0.0f%%)", cond.PrecipitationProbability)
					colored := txt.ColorizePercent(raw, float64(cond.PrecipitationProbability), txt.GreenYellowRedGradient())
					prob = txt.PadRightWidth(colored, 5)
				} else {
					prob = txt.PadRightWidth("", 5)
				}
				fmt.Printf("%02d:00 %s %s %s %s\n", time.Hour(), emojiCol, tempCol, prob, cond.Summary)
			}
		}
	}

	return nil
}
