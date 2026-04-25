package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	gitbackend "go.lichturm.de/lich/internal/hosting/git"
)

var (
	communeAuthor        string
	communeRemote        string
	communePrimaryBranch string
)

var communeCmd = &cobra.Command{
	Use:     "commune [component]",
	Aliases: []string{"update"},
	Short:   "List open renovate MRs",
	Long: `commune lists open merge requests created by renovate (or another author).

Branches are discovered via plain git remote inspection — no hosting API or
credentials beyond normal git access are required.

If a component name is provided, only MRs whose branch name contains that
string are shown. This is useful when a repository contains many components
and you want to focus on one.

Examples:
  lich commune
  lich commune my-app
  lich commune --author dependabot
  lich commune --remote upstream --primary-branch main`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var component string
		if len(args) == 1 {
			component = args[0]
		}

		backend := gitbackend.New(communeRemote, communePrimaryBranch, ".")

		mrs, err := backend.ListMergeRequests(context.Background(), communeAuthor)
		if err != nil {
			return err
		}

		// Filter by component (substring match on branch name)
		if component != "" {
			filtered := mrs[:0]
			for _, mr := range mrs {
				if strings.Contains(mr.Branch, component) {
					filtered = append(filtered, mr)
				}
			}
			mrs = filtered
		}

		if len(mrs) == 0 {
			msg := fmt.Sprintf("No open MRs found (author: %q, remote: %q", communeAuthor, communeRemote)
			if component != "" {
				msg += fmt.Sprintf(", component: %q", component)
			}
			msg += ")"
			fmt.Fprintln(os.Stderr, msg)
			return nil
		}

		if !IsInteractive() {
			return outputJSON(mrs)
		}

		fmt.Printf("Found %d MR(s):\n\n", len(mrs))
		for i, mr := range mrs {
			fmt.Printf("  [%d] %s\n", i+1, mr.Branch)
			fmt.Printf("      author: %-20s target: %s\n", mr.Author, mr.TargetBranch)
		}
		return nil
	},
}

func outputJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func init() {
	communeCmd.Flags().StringVar(&communeAuthor, "author", "renovate",
		"filter MRs by author prefix in branch name (e.g. 'renovate', 'dependabot')")
	communeCmd.Flags().StringVar(&communeRemote, "remote", "origin",
		"git remote to inspect")
	communeCmd.Flags().StringVar(&communePrimaryBranch, "primary-branch", "master",
		"primary branch that MRs target")
	rootCmd.AddCommand(communeCmd)
}
