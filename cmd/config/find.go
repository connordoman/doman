package config

import (
	"fmt"
	"strings"

	"github.com/connordoman/doman/internal/config"
	"github.com/connordoman/doman/internal/txt"
	"github.com/spf13/cobra"
)

var FindConfigCommand = &cobra.Command{
	Use:   "find",
	Short: "Print the path to the configuration file",
	RunE:  runConfigFindCommand,
}

func init() {

}

func runConfigFindCommand(cmd *cobra.Command, args []string) error {
	configPath, err := config.GetConfigPath()
	if err != nil {
		return err
	}

	if strings.Contains(configPath, " ") {
		configPath = fmt.Sprintf("\"%s\"", strings.TrimSpace(configPath))
	}

	cmd.Println(txt.Bluef("%s", configPath))
	return nil
}
