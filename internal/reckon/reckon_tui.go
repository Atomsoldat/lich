package reckon

import (
	"fmt"

	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6"
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

	fmt.Print(selectedBranches)

	repo, err := lich_git.OpenRepo()
	if err != nil {
		return fmt.Errorf(
				"Failed opening repo: %w",
				err,
			)
	}
	defer repo.Close()

	for i := range selectedBranches {
	    checkoutBranch(repo, selectedBranches[i])
        }
	return nil
}


func checkoutBranch(repo *git.Repository, branch plumbing.Reference) (error){

    //we will need to pass or open a Repository instance
    // https://pkg.go.dev/github.com/go-git/go-git/v6#Repository.Worktree
    worktree, err := repo.Worktree()
    if err != nil {
        return fmt.Errorf("Failed to fetch repo worktree: %v", err)
    }


    // https://pkg.go.dev/github.com/go-git/go-git/v6@v6.0.0-alpha.5/plumbing#Reference.Name
    branchName := branch.Name()


    checkoutOptions :=  new(git.CheckoutOptions)
    checkoutOptions.Keep = true
    // https://pkg.go.dev/github.com/go-git/go-git/v6@v6.0.0-alpha.5/plumbing#ReferenceName
    checkoutOptions.Branch = branchName // we need a plumbing.ReferenceName here , alternatively a plumbing.Hash


    // for this to work, we need a Worktree struct
    // the worktree is also the thingy that can git add, git commit, git pull, git grep
    // https://pkg.go.dev/github.com/go-git/go-git/v6#Worktree
    // https://pkg.go.dev/github.com/go-git/go-git/v6#Worktree.Checkout
    worktree.Checkout(checkoutOptions)
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
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice.Strings())
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return tea.NewView(s)
}
