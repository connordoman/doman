package git

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"doman.sh/doman/internal/pkg"
	"doman.sh/doman/internal/txt"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

type AuthorCommandJSON struct {
	User   pkg.GitUser `json:"user"`
	GitHub string      `json:"github,omitempty"`
}

var AuthorCommand = &cobra.Command{
	Use:   "author",
	Short: "Print the current git.user & git.email",
	Run:   executeAuthor,
}

func init() {
	AuthorCommand.Flags().BoolP("global", "g", false, "Check the global git config")
	AuthorCommand.Flags().BoolP("echo", "e", false, "Print the underlying commands being executed")
	AuthorCommand.Flags().BoolP("json", "j", false, "Print the author information in JSON format")
}

// bash:   echo "$(git config --get user.name) <$(git config --get user.email)>"

func executeAuthor(cmd *cobra.Command, args []string) {
	jsonFlag, _ := cmd.Flags().GetBool("json")

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

	fmt.Printf("%s %s\n", txt.Boldf("%s", strings.TrimSpace(userName)), txt.Greyf("%s", strings.TrimSpace(email)))

	ghCLIInstalled := pkg.CheckGitHubCLIInstalled()
	ghUser := ""
	if ghCLIInstalled {
		ghUser, err = pkg.GetGitHubCLIUser()
		if err != nil {
			pkg.PrintInfo("could not get gh user (are you logged in?)")
			return
		}
	} else {
		return
	}

	if jsonFlag {
		authorJSON := AuthorCommandJSON{
			User: pkg.GitUser{
				Name:  strings.TrimSpace(userName),
				Email: strings.TrimSpace(email),
			},
		}
		if ghCLIInstalled {
			authorJSON.GitHub = ghUser
		}
		jsonOutput, err := json.MarshalIndent(authorJSON, "", "  ")
		if err != nil {
			pkg.FailAndExit("Failed to marshal author information to JSON: %v", err)
		}
		fmt.Println(string(jsonOutput))
		return
	}

	ghUserStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))

	if ghCLIInstalled {
		fmt.Printf("%s %s\n", ghUserStyle.Bold(true).Render("gh"), ghUserStyle.Render("→ "+strings.TrimSpace(ghUser)))
	} else {
		fmt.Printf("%s %s %s\n", ghUserStyle.Render("\uea6c"), ghUserStyle.Bold(true).Render("gh"), ghUserStyle.Render("not installed"))
	}
}
