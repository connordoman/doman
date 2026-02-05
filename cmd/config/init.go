package config

import (
	"fmt"
	"os"

	"github.com/connordoman/doman/internal/config"
	"github.com/connordoman/doman/internal/pkg"
	"github.com/spf13/cobra"
)

var InitConfigCommand = &cobra.Command{
	Use:   "init",
	Short: "Create the configuration file if missing",
	RunE:  runInitConfigCommand,
}

func runInitConfigCommand(cmd *cobra.Command, args []string) error {
	configDir, err := config.GetConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", configDir, err)
	}

	configPath := config.ConfigPath("config.yaml")
	if pkg.FileExists(configPath) {
		pkg.PrintInfo("Configuration file already exists: %s", configPath)
		return nil
	}

	if err := config.SaveConfig(); err != nil {
		return fmt.Errorf("failed to create configuration file: %w", err)
	}

	pkg.PrintSuccess("Configuration file created: %s", configPath)
	return nil
}
