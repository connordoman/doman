package weather

import (
	"fmt"

	"github.com/connordoman/doman/internal/pkg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var WeatherCommand = &cobra.Command{
	Use:   "weather",
	Short: "Get the weather for your current location",
	RunE:  runWeatherCommand,
}

func init() {
	WeatherCommand.Flags().StringP("api-key", "A", "", "OpenWeatherMap API key")
	WeatherCommand.Flags().BoolP("short", "s", false, "Display a short version of the weather")
}

func runWeatherCommand(cmd *cobra.Command, args []string) error {
	apiKey, _ := cmd.Flags().GetString("api-key")
	if apiKey != "" {
		viper.Set("weather.openweathermap.api_key", apiKey)
	}

	weather, err := pkg.GetWeatherCurrentLocation()
	if err != nil {
		return err
	}
	if weather == nil {
		pkg.FailAndExit("Failed to retrieve weather data for your current location.")
	}

	// temp := weather.Main.Temp
	// desc := weather.Weather[0].Description
	// placeName := weather.Name

	// fmt.Printf("%.1f°C \u2013 %s \u2013 %s\n", temp, txt.Capitalize(desc), placeName)

	if short, _ := cmd.Flags().GetBool("short"); short {
		fmt.Printf("%s\n", weather.ShortString())
		return nil
	}

	fmt.Printf("%s\n", weather.String())
	return nil
}
