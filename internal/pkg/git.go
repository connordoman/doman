package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"doman.sh/doman/internal/txt"
	"github.com/charmbracelet/lipgloss"
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

var ErrGitHubCLIUnavailable = fmt.Errorf("GitHub CLI is not installed")

var (
	openPRStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#00a63e"))
	closedPRStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#e7000b"))
	mergedPRStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#8957e5"))

	gitBranchStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#4493f8")).Background(lipgloss.Color("#111d2e"))
	gitCommitStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#fe9a00")).Background(lipgloss.Color("#461901"))
)

func CheckGitHubCLIInstalled() bool {
	_, err := RunCommand(GitHubCLICmd, "version")

	return err == nil
}

func IsGitHubCLIUnavailableError(err error) bool {
	return errors.Is(err, ErrGitHubCLIUnavailable)
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

func GetGitBranch() (string, error) {
	branch, err := RunCommand("git", "branch", "--show-current")
	if err != nil {
		return "", fmt.Errorf("failed to get git branch: %w", err)
	}

	return strings.TrimSpace(branch), nil
}

func GetGitBranches() ([]string, error) {
	branchesRaw, err := RunCommand("git", "branch", "--format=%(refname:short)")
	if err != nil {
		return nil, fmt.Errorf("failed to get git branches: %w", err)
	}
	lines := strings.Split(branchesRaw, "\n")

	var branches []string
	for _, line := range lines {
		branch := strings.TrimSpace(line)
		if branch == "" {
			continue
		}

		branches = append(branches, branch)
	}

	return branches, nil
}

func GetGitRemoteName() (string, error) {
	remote, err := RunCommand("git", "remote", "show")
	if err != nil {
		return "", fmt.Errorf("failed to get git remote: %w", err)
	}

	return strings.TrimSpace(remote), nil
}

func ValidateGitCommitHash(commitHash string) bool {
	_, err := RunCommand("git", "rev-parse", "--verify", commitHash)
	return err == nil
}

func GetAllGitBranchesShortNames() ([]string, error) {
	branches, err := RunCommand("git", "branch", "--list", "--no-color", "--format=%(refname:short)")
	if err != nil {
		return nil, fmt.Errorf("failed to get all git branches short names: %w", err)
	}

	return strings.Split(strings.TrimSpace(branches), "\n"), nil
}

func GetGitMergeBase(branch, otherBranch string) (string, error) {
	mergeBase, err := RunCommand("git", "merge-base", branch, otherBranch)
	if err != nil {
		return "", fmt.Errorf("failed to get git merge base: %w", err)
	}

	return strings.TrimSpace(mergeBase), nil
}

func GetCommitTimestamp(commitHash string) (*time.Time, error) {
	timestampString, err := RunCommand("git", "show", "-s", "--format=%ct", commitHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit timestamp: %w", err)
	}

	timestampUnix, err := strconv.ParseInt(strings.TrimSpace(timestampString), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse commit timestamp: %w", err)
	}

	timestamp := time.Unix(timestampUnix, 0)
	return &timestamp, nil
}

func GetGitFirstCommit(branch string) (string, error) {
	firstCommit, err := RunCommand("git", "rev-list", "--max-count=1", branch)
	if err != nil {
		return "", fmt.Errorf("failed to get git first commit: %w", err)
	}

	return strings.TrimSpace(firstCommit), nil
}

func GetGitBranchParentCommit(branch string) (string, error) {
	// use reflog (may not be available)
	reflog, err := RunCommand("git", "reflog", "show", branch)
	if err == nil && strings.TrimSpace(reflog) != "" {
		lines := strings.Split(strings.TrimSpace(reflog), "\n")
		if len(lines) > 0 {
			lastLine := lines[len(lines)-1]
			fields := strings.Fields(lastLine)
			if len(fields) > 0 {
				commitHash := fields[0]
				if ValidateGitCommitHash(commitHash) {
					return commitHash, nil
				}
			}
		}
	}

	// analyze commit graph
	allBranches, err := GetAllGitBranchesShortNames()
	if err != nil {
		return "", fmt.Errorf("failed to get all git branches short names: %w", err)
	}

	var bestMergeBase string
	var bestMergeBaseTimestamp *time.Time

	for _, otherBranch := range allBranches {
		otherBranch = strings.TrimSpace(otherBranch)
		if otherBranch == "" || otherBranch == branch {
			continue
		}

		mergeBase, err := GetGitMergeBase(branch, otherBranch)
		if err != nil {
			continue
		}

		commitTimestamp, err := GetCommitTimestamp(mergeBase)
		if err != nil {
			continue
		}

		if bestMergeBaseTimestamp == nil || commitTimestamp.After(*bestMergeBaseTimestamp) {
			bestMergeBase = mergeBase
			bestMergeBaseTimestamp = commitTimestamp
		}
	}

	if bestMergeBase != "" {
		return bestMergeBase, nil
	}

	// fallback if no merge base found
	// maybe initial commit or branch w/o history
	firstCommit, err := GetGitFirstCommit(branch)
	if err != nil {
		if strings.HasSuffix(err.Error(), "exit status 128") {
			return "unknown", nil
		}
		return "", fmt.Errorf("failed to get git first commit: %w", err)
	}

	return firstCommit, nil
}

type GitHubPRSimple struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	Base   string `json:"baseRefName"`
}

const (
	GitHubPRStateOpen   = "OPEN"
	GitHubPRStateClosed = "CLOSED"
	GitHubPRStateMerged = "MERGED"
)

func (p GitHubPRSimple) String() string {
	return fmt.Sprintf("#%d %s %s %s", p.Number, p.Title, txt.Greyf("→"), p.Base)
}

func (p GitHubPRSimple) ColorizedString() string {
	var style lipgloss.Style
	var titleStyle = lipgloss.NewStyle()
	switch p.State {
	case GitHubPRStateOpen:
		style = openPRStyle
	case GitHubPRStateClosed:
		style = closedPRStyle
		titleStyle = titleStyle.Strikethrough(true)
	case GitHubPRStateMerged:
		style = mergedPRStyle
	}

	number := style.Render(fmt.Sprintf("#%d", p.Number))
	base := RenderGitBranch(p.Base)
	title := titleStyle.Render(p.Title)

	return fmt.Sprintf("%s %s %s %s", number, title, txt.Greyf("→"), base)
}

func GetPRListForBranch(branch string) ([]GitHubPRSimple, error) {
	prs, err := RunCommand(GitHubCLICmd, "pr", "list", "--state=all", "--json=number,title,state,baseRefName", "--head", branch)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR list for branch: %w", err)
	}

	var prsResponse []GitHubPRSimple
	if err := json.Unmarshal([]byte(prs), &prsResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal PR list: %w", err)
	}

	return prsResponse, nil
}

func RenderGitBranch(branch string) string {
	return gitBranchStyle.Render(" " + branch + " ")
}

func RenderGitCommit(commit string) string {
	return gitCommitStyle.Render(" " + commit + " ")
}

func CheckIsGitRepo() bool {
	_, err := RunCommand("git", "rev-parse", "--is-inside-work-tree")
	return err == nil
}
