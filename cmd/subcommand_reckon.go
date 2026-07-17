package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.lichturm.de/lich/internal/lich"
)

var (
	reckonAuthor        string
	reckonRemote        string
	reckonPrimaryBranch string
)

var reckonCmd = &cobra.Command{
	Use:     "reckon [component]",
	Aliases: []string{"update"},
	Short:   "List open renovate MRs",
	Long: `reckon lists unmerged branches created by renovate (or another author).

Branches are discovered via plain git remote inspection — no hosting API or
credentials beyond normal git access are required.

If a component name is provided, only MRs whose branch name contains that
string are shown. This is useful when a repository contains many components
and you want to focus on one.

Examples:
  lich reckon
  lich reckon specific-component
  lich reckon --author dependabot
  lich reckon --remote upstream --primary-branch main`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

	err := lich.Reckon()
	if err != nil {
		return fmt.Errorf(
			"could not execute reckon subcommand: %w",
			err,
		)

	}


		return nil
	},
}

func init() {
	reckonCmd.Flags().StringVar(&reckonAuthor, "author", "renovate",
		"filter MRs by author prefix in branch name (e.g. 'renovate', 'dependabot')")
	reckonCmd.Flags().StringVar(&reckonRemote, "remote", "origin",
		"git remote to inspect")
	reckonCmd.Flags().StringVar(&reckonPrimaryBranch, "primary-branch", "master",
		"primary branch that MRs target")
	rootCmd.AddCommand(reckonCmd)
}
