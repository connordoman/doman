package cmd

import "github.com/spf13/cobra"

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:
	
Bash:
    $ source <(doman completion bash)
    # To load completions for each session, execute once:
    # Linux:
    $ doman completion bash > /etc/bash_completion.d/doman
    # macOS:
    $ doman completion bash > /usr/local/etc/bash_completion.d/doman

Zsh:
    # If shell completion is not already enabled in your environment,
    # you will need to enable it.  You can execute the following once:
    $ echo "autoload -U compinit; compinit" >> ~/.zshrc
    # To load completions for each session, execute once:
    $ doman completion zsh > "${fpath[1]}/_doman"
    # You will need to start a new shell for this setup to take effect.

Fish:
    $ doman completion fish | source
    # To load completions for each session, execute once:
    $ doman completion fish > ~/.config/fish/completions/doman.fish

PowerShell:
    PS> doman completion powershell | Out-String | Invoke-Expression
    # To load completions for every new session, run:
    PS> doman completion powershell > doman.ps1
    # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(cmd.OutOrStdout())
		case "zsh":
			cmd.Root().GenZshCompletion(cmd.OutOrStdout())
		case "fish":
			cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
		case "powershell":
			cmd.Root().GenPowerShellCompletion(cmd.OutOrStdout())
		default:
			cmd.Help() // Show help if an invalid argument is provided
			return
		}
	},
}
