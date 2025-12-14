package pkg

import (
	"encoding/json"
	"fmt"
)

type GitUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type GitHubCLIUser struct {
	State       string `json:"state"`
	Active      bool   `json:"active"`
	Host        string `json:"host"`
	Login       string `json:"login"`
	TokenSource string `json:"tokenSource"`
	Scopes      string `json:"scopes"`
	GitProtocol string `json:"gitProtocol"`
}

type GitHubCLIAuthStatusResponse struct {
	Hosts map[string][]GitHubCLIUser `json:"hosts"`
}

const (
	GitHubCLICmd = "gh"
)

func CheckGitHubCLIInstalled() bool {
	_, err := RunCommand(GitHubCLICmd, "version")

	return err == nil
}

func GetGitHubCLIUser() (string, error) {
	authStatus, err := RunCommand(GitHubCLICmd, "auth", "status", "-a", "--json", "hosts")
	if err != nil {
		return "", fmt.Errorf("failed to get GitHub CLI auth status: %w", err)
	}

	var authStatusResponse GitHubCLIAuthStatusResponse
	if err := json.Unmarshal([]byte(authStatus), &authStatusResponse); err != nil {
		return "", fmt.Errorf("failed to unmarshal GitHub CLI auth status: %w", err)
	}

	for _, users := range authStatusResponse.Hosts["github.com"] {
		if users.Active {
			return users.Login, nil
		}
	}

	return "", fmt.Errorf("no active GitHub CLI user found")
}
