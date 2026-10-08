package reckon

import (
	"fmt"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"go.lichturm.de/lich/internal/lich_git"

	tea "charm.land/bubbletea/v2"
)

// bubbletea basically thinks of an application in a client server way
// the "server" has a model containing all the data
// the client sends messages that possibly modify the model
// the server returns updated visual data
type model struct {
	choices  []plumbing.Reference // items on todo list
	cursor   int                  // which todo list item the cursor is on
	selected map[int]struct{}     // which todo list items are selected
}

var unmergedBranches []plumbing.Reference

// boiler plate so we can return errors in our function
// go does not allow that in the main() funtion
func TuiWorkflow() error {
	// use long form so we don't shadow the global variable
	var err error
	err, unmergedBranches = FindUnmergedRemoteBranches()
	if err != nil {
		return fmt.Errorf("Could not determine unmerged branches: %v", err)
	}

	program := tea.NewProgram(initialModel())

	var returnModel tea.Model
	returnModel, err = program.Run()
	if err != nil {
		fmt.Printf("Error occured: %v", err)
		return fmt.Errorf("Error Occured during TUI execution: %v", err)
	}

	// Type assertion allows us to access the values in our implementation of the tea.Model interface
	selections := returnModel.(model).selected

	var selectedBranches []plumbing.Reference

	// build an array of selected branches using the index of selections to retrieve them
	// from unmergedBranches
	for i := range selections {
		selectedBranches = append(selectedBranches, unmergedBranches[i])
	}

	repo, err := lich_git.OpenRepo()
	if err != nil {
		return fmt.Errorf(
			"Failed opening repo: %w",
			err,
		)
	}
	defer repo.Close()

	for i := range selectedBranches {
		err = checkoutBranch(repo, selectedBranches[i])
		if err != nil {
			return fmt.Errorf("Failed to check out branch %v: %v", selectedBranches[i].Name(), err)
		}
	}
	return nil
}

// in english, what this does is:
// get all remotes from .git/config
// from there, get each remote's fetch refspecs
// for each one
//
//	flip that refspec
//	if the provided remote ReferenceName matches the source side of our flipped refspec
//	that means we have the correct remote out of our possibly big bunch of remotes
//	we can then  grab the correct local ref name from the destination side of the flipped refspec
//
// we end up returning either a local ref name (that may or may not exist yet) , or an error, if remoteRef belongs to no configured remote
func localBranchName(repo *git.Repository, remoteRef plumbing.ReferenceName) (plumbing.ReferenceName, error) {
	remotes, err := repo.Remotes()
	if err != nil {
		return "", fmt.Errorf("failed to list remotes: %w", err)
	}
	for _, remote := range remotes {
		for _, refspec := range remote.Config().Fetch {
			// https://git-scm.com/book/en/v2/Git-Internals-The-Refspec
			// https://pkg.go.dev/github.com/go-git/go-git/v6@v6.0.0-beta.1/config#RefSpec
			// a refspec is a renaming rule in .git/config or git commands
			// its purpose is to map refs from a source to a destination
			// it mainly consists out of two patterns separated by a colon
			// `src:dst`
			//
			// in fetch refspecs
			// src is the ref pattern matching a branch in the remote repo;
			// it's how the remote thinks about a branch
			// dst is the ref pattern matching a local, remote-tracking ref, it's how we will think about that branch
			// that's how git knows how to call different remote server's branches locally
			//
			// in push refspecs
			// src is a ref pattern to a local ref (e.g. a branch)
			// dst is a ref pattern matching a branch in the remote
			//
			// example (fetch refspec):
			//
			// [remote "origin"]
			//     url = git@github.com:MausoleumManagement/private-cloud.git
			//     fetch = +refs/heads/*:refs/remotes/origin/*
			//
			// `+` is related to updates
			// basically, it means "it's okay to  update to a non fast-forward change,
			// just jump to whatever the remote is doing, because that is a remote branch,
			// and we have no local stuff on it"

			// refs/heads/* represents all branches in the remote, the way the remote sees them
			// (refs/heads is where local branches live in a git repo)
			// and refs/remotes/origin/* represents all those branches from the remote repo `origin` in our local repo
			reverse := refspec.Reverse()
			if reverse.Match(remoteRef) {
				// Dst returns the "to" name for a given "from" name
				return reverse.Dst(remoteRef), nil
			}
		}
	}
	return "", fmt.Errorf("no remote fetch refspec matches %v", remoteRef)
}

func checkoutBranch(repo *git.Repository, branch plumbing.Reference) error {

	//we will need to pass or open a Repository instance
	// https://pkg.go.dev/github.com/go-git/go-git/v6#Repository.Worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("Failed to fetch repo worktree: %v", err)
	}

	// TODO: symbolic refs like origin/HEAD  should be resolved for super correctness
	// Use case?
	// https://pkg.go.dev/github.com/go-git/go-git/v6@v6.0.0-alpha.5/plumbing#Reference.Name
	branchName := branch.Name()

	localName := branchName

	if branchName.IsRemote() {
		// resolve proper local name for provided remote branch
		localName, err = localBranchName(repo, branchName)
		if err != nil {
			return fmt.Errorf("Failed to determine local branch name: %v", err)
		}
	} else {
		localName = branchName
	}
	fmt.Printf("DEBUG: localName is: %v\n", localName)

	checkoutOptions := new(git.CheckoutOptions)
	// since we do not set Force = true, we can not clobber local changes
	// in this case, this simply means "don't carry over changed, committed stuff
	// to the working tree when checking out the other branch"
	// we also fail on uncommited changes
	checkoutOptions.Keep = false
	checkoutOptions.Force = false

	fmt.Printf("DEBUG: passed branchName: %v\n", localName)

	// try getting a reference for our referenceName; if that succeeds, the branch already exists locally
	_, err = repo.Reference(localName, false)
	if err == nil {
		fmt.Printf("DEBUG: Branch %v exists locally already\n", localName)
		checkoutOptions.Branch = localName
		checkoutOptions.Create = false

	} else {
		// https://pkg.go.dev/github.com/go-git/go-git/v6@v6.0.0-alpha.5/plumbing#ReferenceName
		// we need a plumbing.ReferenceName here , alternatively a plumbing.Hash
		fmt.Printf("DEBUG: Branch %v does not exist locally\n", localName)
		checkoutOptions.Branch = localName

		// TODO: for this to work, we need to first figure out whether a given reference is symbolic
		// resolve the provided references, in case it's a symbolic reference like HEAD
		// those have no Hash attribute by themselves
		//resolvedReference, err := repo.Reference(localName, true)
		//if err != nil {
		//	return fmt.Errorf("Failed to resolve Reference for branch %v: %v", localName, err)
		//}

		branchHash := branch.Hash()
		checkoutOptions.Hash = branchHash
		checkoutOptions.Create = true
	}

	// for this to work, we need a Worktree struct
	// the worktree is also the thingy that can git add, git commit, git pull, git grep
	// https://pkg.go.dev/github.com/go-git/go-git/v6#Worktree
	// https://pkg.go.dev/github.com/go-git/go-git/v6#Worktree.Checkout
	err = worktree.Checkout(checkoutOptions)
	if err != nil {
		return fmt.Errorf("Failed to check out desired branch %v: %v", localName, err)
	}
	fmt.Println("DEBUG: Checked out branch")

	return nil
}

// this is bubbletea's initial state
func initialModel() model {
	return model{
		choices: unmergedBranches,
		// map with ints as keys and structs as values
		selected: make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

// This gets called, when the "client" does stuff
// based on what the client did, we update the model
// perhaps do some other stuff
// and then return the updated model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyPressMsg:

		// Neat, what was the actual key pressed?
		switch msg.String() {
		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// The "enter" key and the space bar toggle the selected state
		// for the item that the cursor is pointing at.
		case "enter", "space":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

// based on the current state of our model, this function
func (m model) View() tea.View {
	// The header
	s := "Which branches shall we operate on, master?\n\n"

	// Iterate over our choices
	for i, choice := range m.choices {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Is this choice selected?
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = "x" // selected!
		}

		// Render the row
		//s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice.Strings())
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice.Name().Short())
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return tea.NewView(s)
}
