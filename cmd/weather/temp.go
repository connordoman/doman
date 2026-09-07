package weather

import (
	"fmt"
	"strings"

	"doman.sh/internal/pkg/weather"
	"doman.sh/internal/txt"
	"github.com/spf13/cobra"
)

var TempCommand = &cobra.Command{
	Use:   "temp",
	Short: "Display the temperature gradient",
	RunE:  runTempCommand,
}

func init() {
	TempCommand.Flags().Float64("min", weather.TemperatureMinCelsius, "Minimum temperature in Celsius")
	TempCommand.Flags().Float64("max", weather.TemperatureMaxCelsius, "Maximum temperature in Celsius")
}

func runTempCommand(cmd *cobra.Command, args []string) error {
	minFlag, _ := cmd.Flags().GetFloat64("min")
	maxFlag, _ := cmd.Flags().GetFloat64("max")

	minCelsius := float32(minFlag)
	maxCelsius := float32(maxFlag)

	temps := []string{}
	bars := []string{}
	for tempCelsius := minCelsius; tempCelsius <= maxCelsius; tempCelsius++ {
		percent := weather.TemperaturePercent(tempCelsius)
		color := txt.ColorizePercent(fmt.Sprintf("%0.0f°C", tempCelsius), percent*100, txt.TemperatureGradient())
		temps = append(temps, color)

		bar := txt.ColorizePercent(strings.Repeat("█", int(percent*10)), percent*100, txt.TemperatureGradient())
		bars = append(bars, bar)
	}
	fmt.Println(strings.Join(temps, "\n"))
	fmt.Println(strings.Join(bars, ""))

	return nil
}
