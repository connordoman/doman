package cmd

import (
	"fmt"

	"github.com/connordoman/doman/internal/config"
	"github.com/connordoman/doman/internal/pkg/alias"
	"github.com/connordoman/doman/internal/txt"
	"github.com/spf13/cobra"
)

var AliasCmd = &cobra.Command{
	Use:   "alias <name> <command>",
	Short: "Create an an alias for a command",
	RunE:  runAliasCmd,
	Args:  cobra.ExactArgs(2),
}

func init() {
	AliasCmd.Flags().StringP("description", "d", "", "Description of the alias")
}

func runAliasCmd(cmd *cobra.Command, args []string) error {
	name := args[0]
	command := args[1]

	verbose, _ := cmd.Flags().GetBool("verbose")
	description, _ := cmd.Flags().GetString("description")

	usingZsh := config.IsUsingZsh()
	if verbose {
		fmt.Println(txt.Bluef("using zsh: %v", usingZsh))
	}

	if !usingZsh {
		return fmt.Errorf("sorry, aliases are currently only supported in zsh")
	}

	if err := alias.Setup(); err != nil {
		return err
	}

	a, err := alias.NewAlias(name, command)
	if err != nil {
		return err
	}

	if description != "" {
		a.Describe(description)
		if verbose {
			fmt.Printf("%q\n", a.Description)
		}
	}

	aliasPath, err := a.Save()
	if err != nil {
		return err
	}

	fmt.Println(txt.Successf("created alias file at \"%s\"", aliasPath))
	fmt.Println(a)

	return nil
}
