package cmd

import (
	"github.com/connordoman/doman/cmd/ask"
	"github.com/connordoman/doman/cmd/completions"
	"github.com/connordoman/doman/cmd/example"
	"github.com/connordoman/doman/cmd/git"
	go_self "github.com/connordoman/doman/cmd/go"
	"github.com/connordoman/doman/cmd/npm"
	"github.com/connordoman/doman/cmd/sys"
	"github.com/connordoman/doman/internal/config"
	"github.com/connordoman/doman/internal/pkg"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "doman",
	Short: "A simple CLI application for domain-management tasks",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		pkg.SetEcho(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// Initialize configuration
	if err := config.InitConfig(); err != nil {
		pkg.FailAndExit("Error initializing configuration: %v", err)
	}

	// Set up Viper
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Flags
	rootCmd.PersistentFlags().BoolP("echo", "e", false, "Print the underlying commands being executed")

	// Commands
	rootCmd.AddCommand(ask.AskCommand)
	rootCmd.AddCommand(completions.CompletionsCommand)
	rootCmd.AddCommand(git.AuthorCommand)
	rootCmd.AddCommand(git.RemotesCommand)
	rootCmd.AddCommand(npm.LockfileCommand)
	rootCmd.AddCommand(sys.IPCommand)

	// Subject to removal
	rootCmd.AddCommand(go_self.RunCommand)
	rootCmd.AddCommand(example.ExampleCommand)
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}
