package git

import (
	"fmt"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
)

func displayBranches(repo *git.Repository) (error) {
	branches, _ := repo.Branches()
	branches.ForEach(func(branch *plumbing.Reference) error {
		fmt.Println(branch.Hash().String(), branch.Name())
		// ForEach expects a callback returning an error
		return nil
	})
	return nil
}

func fetchRemote(repo *git.Repository) (error) {
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
