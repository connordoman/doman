package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	ConfigDir    = ".config"
	AppConfigDir = "doman"
)

func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ConfigDir, AppConfigDir)
	return configDir, nil
}

func ConfigPath(segments ...string) string {
	configDir, err := GetConfigPath()
	if err != nil {
		return ""
	}
	return filepath.Join(configDir, filepath.Join(segments...))
}

func SaveConfig() error {
	configDir, err := GetConfigPath()
	if err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.yaml")
	return viper.WriteConfigAs(configPath)
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
	viper.SetDefault("ask.openai.default_model", "gpt-4o-mini")
	viper.SetDefault("ask.openai.quick_model", "gpt-5-mini")
	viper.SetDefault("ask.preferred_languages", []string{})

	viper.BindEnv("ask.openai.api_key", "OPENAI_API_KEY")

	_ = viper.ReadInConfig()

	return nil
}
