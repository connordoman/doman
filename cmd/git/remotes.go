package git

import (
	"fmt"
	"strings"

	"github.com/connordoman/doman/internal/pkg"
	"github.com/connordoman/doman/internal/txt"
	"github.com/spf13/cobra"
)

var RemotesCommand = &cobra.Command{
	Use:   "remotes",
	Short: "Print a list of git remotes",
	Run:   executeRemotes,
}

func init() {
}

func executeRemotes(cmd *cobra.Command, args []string) {
	remotesRaw, err := pkg.RunCommand("git", "remote", "-v")
	if err != nil {
		pkg.FailAndExit("failed to get git remotes: %v", err)
	}

	if remotesRaw == "" {
		pkg.PrintError("No remotes found")
		return
	}

	remotes := strings.Split(remotesRaw, "\n")

	for _, remote := range remotes {
		if remote == "" {
			continue
		}
		formattedRemote, err := formatRemote(remote)
		if err != nil {
			pkg.PrintError("Error formatting remote '%s': %v", remote, err)
			continue
		}
		fmt.Println(formattedRemote)
	}
}

func parseRemote(remote string) (string, string, string) {
	modifiedRemote := strings.ReplaceAll(remote, "\t", " ")
	modifiedRemote = strings.TrimSpace(modifiedRemote)
	if !strings.Contains(modifiedRemote, " ") {
		return "", "", ""
	}

	parts := strings.Split(modifiedRemote, " ")
	if len(parts) < 2 {
		return "", "", ""
	}

	name := parts[0]
	url := parts[1]
	fetchType := "fetch"
	if len(parts) > 2 && parts[2] == "(push)" {
		fetchType = "push"
	}

	return name, url, fetchType
}

func formatRemote(remote string) (string, error) {
	name, url, fetchType := parseRemote(remote)
	if name == "" || url == "" {
		return "", fmt.Errorf("invalid remote format")
	}

	fetchIcon := "↑"
	if fetchType == "fetch" {
		fetchIcon = "↓"
	}

	return fmt.Sprintf("%s %s %s", txt.Boldf("%s", name), fetchIcon, txt.Greyf("%s", url)), nil
}
