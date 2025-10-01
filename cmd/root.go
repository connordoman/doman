package cmd

import (
	"github.com/connordoman/doman/cmd/ask"
	"github.com/connordoman/doman/cmd/completions"
	configCmd "github.com/connordoman/doman/cmd/config"
	"github.com/connordoman/doman/cmd/fun"
	"github.com/connordoman/doman/cmd/git"
	go_self "github.com/connordoman/doman/cmd/go"
	"github.com/connordoman/doman/cmd/npm"
	"github.com/connordoman/doman/cmd/rand"
	"github.com/connordoman/doman/cmd/version"
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
	SilenceErrors: true,
	SilenceUsage:  true,
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
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")

	// Commands
	rootCmd.AddCommand(
		ask.AskCommand,
		completions.CompletionsCommand,
		git.AuthorCommand,
		git.RemotesCommand,
		npm.LockfileCommand,
		IPCommand,
		fun.ShrugCommand,
		rand.RandCommand,
		version.VersionCommand,
		configCmd.ConfigCommand,
		SqrtCmd,
		AliasCommand,
	)

	// Subject to removal
	rootCmd.AddCommand(go_self.RunCommand)
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}
