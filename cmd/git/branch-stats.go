package git

import (
	"context"
	"fmt"

	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/connordoman/doman/internal/pkg"
	"github.com/connordoman/doman/internal/txt"
	"github.com/spf13/cobra"
)

var BranchStatsCommand = &cobra.Command{
	Use:   "branch-stats",
	Short: "Print statistics about a git branch",
	RunE:  runBranchStatsCommand,
	Args:  cobra.MaximumNArgs(1),
}

func init() {
}

func runBranchStatsCommand(cmd *cobra.Command, args []string) error {
	branch := ""
	if len(args) > 0 {
		branch = args[0]
	} else {
		currentBranch, err := pkg.GetGitBranch()
		if err != nil {
			return fmt.Errorf("failed to get current branch: %w", err)
		}
		branch = currentBranch
	}

	parentCommit, err := pkg.GetGitBranchParentCommit(branch)
	if err != nil {
		return fmt.Errorf("failed to get parent commit: %w", err)
	}

	branchTimestamp, err := pkg.GetCommitTimestamp(branch)
	if err != nil {
		return fmt.Errorf("failed to get branch timestamp: %w", err)
	}

	fmt.Println(txt.Greyf("Branch:"), pkg.RenderGitBranch(branch))
	fmt.Println(txt.Greyf("Parent:"), pkg.RenderGitCommit("#"+parentCommit))
	fmt.Println(txt.Greyf("Last Commit:"), branchTimestamp.Format("Jan 02, 2006 15:04:05 MST"))

	var prs []pkg.GitHubPRSimple

	if pkg.CheckGitHubCLIInstalled() {
		if spinner.New().Title("Checking for PRs").Style(lipgloss.NewStyle().Foreground(lipgloss.Color("#2563eb"))).ActionWithErr(func(ctx context.Context) error {
			prs, err = pkg.GetPRListForBranch(branch)
			return err
		}).Run(); err != nil {
			return fmt.Errorf("failed to get PR list for branch: %w", err)
		}
	}

	if err != nil {
		if pkg.IsGitHubCLIUnavailableError(err) {
			fmt.Println("GitHub CLI is not installed")
		} else {
			return fmt.Errorf("failed to get PR list for branch: %w", err)
		}
	}

	if len(prs) > 0 {
		fmt.Println(txt.Greyf("PRs:"))
		for _, pr := range prs {
			fmt.Println("  -", pr.ColorizedString())
		}
	} else {
		fmt.Println(txt.Greyf("No PRs"))
	}
	return nil
}
