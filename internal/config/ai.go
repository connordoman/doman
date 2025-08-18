package config

import (
	"github.com/spf13/viper"
)

func IsAskConfigured() bool {
	service := viper.GetString("ask.default_service")

	switch service {
	case "openai":
		return viper.GetString("ask.openai.api_key") != ""
	default:
		return false
	}
}
