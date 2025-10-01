package git

import (
	"fmt"
	"log"
	"strings"

	"github.com/connordoman/doman/internal/pkg"
	"github.com/connordoman/doman/internal/txt"
	"github.com/spf13/cobra"
)

var AuthorCommand = &cobra.Command{
	Use:   "author",
	Short: "Print the current git.user & git.email",
	Run:   executeAuthor,
}

func init() {
	AuthorCommand.Flags().BoolP("global", "g", false, "Check the global git config")
	AuthorCommand.Flags().BoolP("echo", "e", false, "Print the underlying commands being executed")
}

// bash:   echo "$(git config --get user.name) <$(git config --get user.email)>"

func executeAuthor(cmd *cobra.Command, args []string) {

	global, _ := cmd.Flags().GetBool("global")

	gitArgs := []string{"config", "--get"}
	if global {
		gitArgs = append(gitArgs, "--global")
	}

	userName, err := pkg.RunCommand("git", append(gitArgs, "user.name")...)
	if err != nil {
		log.Fatalf("Error getting user.name: %v", err)
	}

	email, err := pkg.RunCommand("git", append(gitArgs, "user.email")...)
	if err != nil {
		log.Fatalf("Error getting user.email: %v", err)
	}

	fmt.Printf("%s\t%s\n", txt.Boldf("\ueb99 %s", strings.TrimSpace(userName)), txt.Greyf("\ueb1c %s", strings.TrimSpace(email)))
}
