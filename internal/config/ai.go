package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	ConfigDir = "doman"
)

func GetConfigPath() (string, error) {
	configBase, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}

	configDir := filepath.Join(configBase, ConfigDir)
	return configDir, nil
}

func InitConfig() error {
	configDir, err := GetConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", configDir, err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)

	viper.AutomaticEnv()
	viper.SetEnvPrefix("DOMAN")

	viper.SetDefault("ask.default_service", "openai")
	viper.SetDefault("ask.openai.model", "gpt-4o-mini")

	viper.BindEnv("ask.openai.api_key", "OPENAI_API_KEY")

	_ = viper.ReadInConfig()

	return nil
}

func SaveConfig() error {
	configDir, err := GetConfigPath()
	if err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.yaml")
	return viper.WriteConfigAs(configPath)
}

func IsAskConfigured() bool {
	service := viper.GetString("ask.default_service")

	switch service {
	case "openai":
		return viper.GetString("ask.openai.api_key") != ""
	default:
		return false
	}
}
