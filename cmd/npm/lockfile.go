package npm

import (
	"path"

	"github.com/connordoman/doman/internal/pkg"
	"github.com/spf13/cobra"
)

var LockfileCommand = &cobra.Command{
	Use:   "npm.lockfile",
	Short: "Generate a lockfile for the current npm project without installing node_modules",
	Run:   execute,
}

func init() {
	LockfileCommand.Flags().StringP("dir", "d", "", "Directory to run the command in (defaults to current directory)")
	LockfileCommand.Flags().BoolP("echo", "e", false, "Print the underlying commands being executed")
}

func execute(cmd *cobra.Command, args []string) {
	echoOn, _ := cmd.Flags().GetBool("echo")
	if echoOn {
		pkg.SetEchoOn()
	} else {
		pkg.SetEchoOff()
	}

	dir, _ := cmd.Flags().GetString("dir")
	if dir != "" {
		if err := pkg.Chdir(dir); err != nil {
			pkg.FailAndExit("failed to change directory to %s: %w", dir, err)
		}
	}

	cwd, err := pkg.Cwd()
	if err != nil {
		pkg.FailAndExit("failed to get current working directory: %w", err)
	}

	packageJson := path.Join(cwd, "package.json")
	if !pkg.FileExists(packageJson) {
		pkg.FailAndExit("no package.json found in current directory: %s", cwd)
	}

	_, err = pkg.RunCommand("npm", "install", "--lockfile-only")
	if err != nil {
		pkg.FailAndExit("failed to generate npm lockfile: %w", err)
	}

	pkg.PrintSuccess("Generated lockfile")
}
