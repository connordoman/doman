package weather

const (
	TemperatureMinCelsius = -20
	TemperatureMaxCelsius = 40
)

func TemperaturePercent(tempCelsius float32) float64 {
	if tempCelsius < TemperatureMinCelsius {
		return 0.0
	}
	if tempCelsius > TemperatureMaxCelsius {
		return 1.0
	}
	return float64(tempCelsius-TemperatureMinCelsius) / float64(TemperatureMaxCelsius-TemperatureMinCelsius)
}
