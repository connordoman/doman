package completions

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/connordoman/doman/internal/pkg"
	"github.com/spf13/cobra"
)

var setupCommand = &cobra.Command{
	Use:   "setup [bash|zsh|fish]",
	Short: "Setup shell completions for doman",
	Long: `Set up shell completions for doman CLI. This will generate the
completion script, install it to the appropriate location, and add the
necessary configuration to your shell profile.

Examples:
  # Set up completions for zsh
  doman completions setup zsh
  
  # Set up completions for bash
  doman completions setup bash

Supported shells: bash, zsh, fish`,
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs: []string{"bash", "zsh", "fish"},
	Run: func(cmd *cobra.Command, args []string) {
		shell := args[0]
		setupCompletions(cmd, shell)
	},
}

func setupCompletions(cmd *cobra.Command, shell string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		pkg.PrintError("Error getting home directory: %v\n", err)
		return
	}

	completionsDir := filepath.Join(homeDir, ".config", "doman", "completions")
	err = os.MkdirAll(completionsDir, 0755)
	if err != nil {
		pkg.PrintError("Failed to create completions directory: %v", err)
		return
	}

	switch shell {
	case "zsh":
		setupZshCompletions(cmd, homeDir, completionsDir)
	case "bash":
		setupBashCompletions(cmd, homeDir, completionsDir)
	case "fish":
		setupFishCompletions(cmd, homeDir, completionsDir)
	default:
		pkg.PrintError("Unsupported shell: %s. Supported shells are: bash, zsh, fish", shell)
		cmd.Help()
		return
	}
}

func setupZshCompletions(cmd *cobra.Command, homeDir, completionsDir string) {
	completionFile := filepath.Join(completionsDir, "_doman")
	tempFile, err := os.Create(completionFile)
	if err != nil {
		pkg.PrintError("Error creating completion file: %v\n", err)
		return
	}
	defer tempFile.Close()

	cmd.Root().GenZshCompletion(tempFile)

	zshrcPath := filepath.Join(homeDir, ".zshrc")

	content, err := os.ReadFile(zshrcPath)
	if err != nil && !os.IsNotExist(err) {
		pkg.PrintError("Error reading .zshrc file: %v\n", err)
		return
	}

	zshConfig := fmt.Sprintf("# doman completions\nfpath=(%s $fpath)\nautoload -Uz compinit && compinit\n", completionsDir)
	if strings.Contains(string(content), completionsDir) {
		fmt.Println("Zsh completions already configured in .zshrc")
	} else {
		f, err := os.OpenFile(zshrcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			pkg.PrintError("Error opening .zshrc: %v\n", err)
			return
		}

		if _, err := f.WriteString("\n" + zshConfig); err != nil {
			f.Close()
			pkg.PrintError("Error writing to .zshrc: %v\n", err)
			return
		}
		f.Close()
		pkg.PrintInfo("Added zsh completions to .zshrc")
	}

	pkg.PrintSuccess("Zsh completions generated successfully")
	pkg.PrintInfo("Please restart your terminal or run `source ~/.zshrc` to apply changes")
}

func setupBashCompletions(cmd *cobra.Command, homeDir, completionsDir string) {
	// Generate completion script
	completionFile := filepath.Join(completionsDir, "doman.bash")
	tempFile, err := os.Create(completionFile)
	if err != nil {
		fmt.Printf("Error creating completion file: %v\n", err)
		return
	}
	defer tempFile.Close()

	cmd.Root().GenBashCompletion(tempFile)

	// Determine bash config file
	bashrcPath := filepath.Join(homeDir, ".bashrc")
	if _, err := os.Stat(bashrcPath); os.IsNotExist(err) {
		// Try .bash_profile on macOS
		bashrcPath = filepath.Join(homeDir, ".bash_profile")
	}

	// Read current file
	content, err := os.ReadFile(bashrcPath)
	if err != nil && !os.IsNotExist(err) {
		fmt.Printf("Error reading bash config: %v\n", err)
		return
	}

	// Check if completions already configured
	bashConfig := fmt.Sprintf("# doman completions\n[[ -f %s ]] && source %s\n", completionFile, completionFile)
	if strings.Contains(string(content), completionFile) {
		fmt.Println("Bash completions already configured")
	} else {
		// Append configuration
		f, err := os.OpenFile(bashrcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Error opening bash config: %v\n", err)
			return
		}

		if _, err := f.WriteString("\n" + bashConfig); err != nil {
			f.Close()
			fmt.Printf("Error writing to bash config: %v\n", err)
			return
		}
		f.Close()
		fmt.Printf("Added Bash completion configuration to %s\n", bashrcPath)
	}

	fmt.Println("Bash completions set up successfully!")
	fmt.Printf("Restart your shell or run 'source %s' to enable completions\n", bashrcPath)
}

func setupFishCompletions(cmd *cobra.Command, homeDir, completionsDir string) {
	// Create fish completions directory if it doesn't exist
	fishDir := filepath.Join(homeDir, ".config", "fish", "completions")
	err := os.MkdirAll(fishDir, 0755)
	if err != nil {
		fmt.Printf("Error creating fish completions directory: %v\n", err)
		return
	}

	// Generate completion script directly to fish directory
	completionFile := filepath.Join(fishDir, "doman.fish")
	tempFile, err := os.Create(completionFile)
	if err != nil {
		fmt.Printf("Error creating completion file: %v\n", err)
		return
	}
	defer tempFile.Close()

	cmd.Root().GenFishCompletion(tempFile, true)

	fmt.Println("Fish completions set up successfully!")
	fmt.Println("Fish will automatically load the completions on next shell start")
}
