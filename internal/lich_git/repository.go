package lich_git

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
)

func DisplayBranches(repo *git.Repository) (error) {
	branches, err := repo.Branches()
	if err != nil {
		return fmt.Errorf(
			"failed fetching repo branches: %w",
			err,
		)
	}
	branches.ForEach(func(branch *plumbing.Reference) error {
		fmt.Println(branch.Hash().String(), branch.Name())
		// ForEach expects a callback returning an error
		return nil
	})
	return nil
}

func FetchRemote(repo *git.Repository) (error) {
	// https://pkg.go.dev/github.com/go-git/go-git/v6#FetchOptions
	fetchopts := git.FetchOptions{
		// default to origin
		//RemoteName: "",
		//RemoteURL: "",
		Force: false,
		Prune: false,
		// TODO: this sounds interesting
		//Filter: packp.Filter,
	
	}
	
	err := repo.Fetch(&fetchopts)
	return err

}

// TODO: what about working trees checked out separately?
func FindGitDirParent(dir string) (string, error) {
	candidate := filepath.Join(dir, ".git")

	result, err := os.Stat(candidate)
	if err != nil {
		return "",
			fmt.Errorf(
			"failed accessing %s: %w",
			candidate,
			err,
		)
	}

	if result.IsDir() {
	    return dir, nil
	// the "parent" of the root dir is the root dir across platforms
	} else if dir == filepath.Dir(dir) {
	    return "", fmt.Errorf(
			"could not discover git dir, walked upwards until %s",
			dir,
		)
	} else {
	return	FindGitDirParent(filepath.Dir(dir))
	}
}
