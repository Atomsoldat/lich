package lich_git

import (
	"fmt"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
)

// this function should result in the same output as
// git branch --no-merged master
// ACTUALLY, this one seems better for our intended use case
// git branch --remote --no-merged master
// with that, we get unmerged remote branches
// now how do we get go-git to do the same?

// TODO: what about the linux-kernel(TM)?
func CollectUnmergedBranches(
	headCommit *object.Commit,
	branchRef *plumbing.Reference,
	// we are diverging from the iterator pattern used by go-git
	// TODO: it might be necessary to write this more efficiently for large repos
	repo *git.Repository,
	unmergedBranches *[]plumbing.Reference,
) error {

	// We can probably not slog through ten thousands of commits for each dumb branch
	// we should limit ourselves to the point of divergence
	// there is a function for that
	// https://pkg.go.dev/github.com/go-git/go-git/v6@v6.0.0-alpha.4/plumbing/object#Commit.MergeBase
	// question is how that is determined
	// TODO: figure out whether invoking this is more or less efficient than
	// determining the commit log once and performing the check manually against
	// each commit hash in it for each branch hash
	//commitLog.ForEach(func(commit *object.Commit) error { ... })

	branchCommit, err := repo.CommitObject((*branchRef).Hash())
	if err != nil {
		return fmt.Errorf(
			"failed to determine commit belonging to hash %s, name %s: %w",
			(*branchRef).Hash(),
			(*branchRef).Name(),
			err,
		)
	}

	merged, err := (*branchCommit).IsAncestor(headCommit)
	if err != nil {
		return fmt.Errorf(
			"failed to determine merge status for commit %s, message %s: %w",
			(*branchCommit).Hash,
			(*branchCommit).Message,
			err,
		)
	}

	if !merged {
		*unmergedBranches = append(*unmergedBranches, *branchRef)
	}
	return nil

}
