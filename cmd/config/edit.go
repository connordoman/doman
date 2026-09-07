package config

import (
	"fmt"
	"os"
	"os/exec"

	"doman.sh/doman/internal/config"
	"doman.sh/doman/internal/pkg"
	"github.com/spf13/cobra"
)

var EditConfigCommand = &cobra.Command{
	Use:   "edit",
	Short: "Edit the configuration file",
	RunE:  runEditConfigCommand,
}

func init() {

}

func runEditConfigCommand(cmd *cobra.Command, args []string) error {
	configPath := config.ConfigPath("config.yaml")

	if !pkg.FileExists(configPath) {
		return fmt.Errorf("configuration file does not exist: %s", configPath)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		// detect nvim
		if _, err := exec.LookPath("nvim"); err == nil {
			editor = "nvim"
		} else {
			editor = "vi"
		}
	}

	command := exec.Command(editor, configPath)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	return command.Run()
}
