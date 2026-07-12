package lich

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"

	"go.lichturm.de/lich/internal/lich_git"
)

// Reckon fetches all references from the remote repository
// It then compares all remote branch references to the master branch
// Unmerged remote branch references will be printed to stdout
func Reckon() error {
	
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf(
			"could not determine working directory: %w",
			err,
		)
	}
	
	path, err := lich_git.FindGitDirParent(pwd)
	if err != nil {
		return fmt.Errorf(
			"could not determine repo directory: %w",
			err,
		)
	}

	// TODO: hardcoded path needs to be properly determined
	//path := "/datengruft/programming/MausoleumManagement/private-cloud"

	// there's a more fancy function for when you store the
	// repo apart from the worktree
	// seemed a bit overkill to start with
	// TODO: if we feel like it, try figuring this out later
	// https://pkg.go.dev/github.com/go-git/go-git/v6#Open
	repo, err := git.PlainOpen(path)
	if err != nil {
		return fmt.Errorf(
			"failed opening repo at path %s: %w",
			path,
			err,
		)
	}
	defer repo.Close()

	// update references locally
	lich_git.FetchRemote(repo)

	// TODO: it would be nice to extract the repo name
	fmt.Println("INFO: Fetched Remote branches for repo")

	// this yields an iterator over all local branches
	// possibly useful, but for now we want to work with remote branches
	//branches, err := repo.Branches()
	//if err != nil {
	//	return fmt.Errorf(
	//		"failed getting repo branches: %w",
	//		err,
	//	)
	//}

	references, err := repo.References()
	if err != nil {
		return fmt.Errorf(
			"failed getting repo references: %w",
			err,
		)
	}

	// TODO:  we probably want to make this configurable
	// i.e. it should not matter which branch is currently
	// checked out
	headRef, err := repo.Head()
	if err != nil {
		return fmt.Errorf(
			"failed getting HEAD reference: %w",
			err,
		)
	}

	headCommit, err := repo.CommitObject((*headRef).Hash())
	if err != nil {
		return fmt.Errorf(
			"failed to determine commit belonging to hash %s, name %s: %w",
			(*headRef).Hash(),
			(*headRef).Name(),
			err,
		)
	}

	// TODO: the size of the slice affects our print statement below (empty lines)
	// we can probably do this smarter by omitting empty lines
	unmergedBranches := make([]plumbing.Reference, 0)

	// TODO: parallelise?
	// TODO: what about the linux-kernel(TM)?
	// anonymous callback wrapping our proper function
	// because the signature of ForEach() demands it

	// This is the local branch variant
	//branches.ForEach(func(branchRef *plumbing.Reference) error {
	// this uses all references that are branches, both remote and local ones
	err = references.ForEach(func(ref *plumbing.Reference) error {

		// Skip symbolic refs like "origin/HEAD -> origin/master".
		if ref.Type() != plumbing.HashReference {
			return nil
		}

		// This is the local branch variant, in case we need it later
		//if ref.Name().IsBranch() { ... }

		if ref.Name().IsRemote() {
			err = lich_git.CollectUnmergedBranches(headCommit, ref, repo, &unmergedBranches)
		}
		
		return nil
	})
	if err != nil {
		return err
	}

	// do stuff with branches
	// TODO: we should probably skip empty branch entries or just prevent them from
	// even showing up in the slice
	fmt.Println("INFO: Determined unmerged branches:")
	for _, branch := range unmergedBranches {
		fmt.Printf("%s\n", branch.Name())
	}

	// TODO: in the future, we might want to clean up merged branches locally and in the remote
	// we could make that an optional configuration

	return nil
}
