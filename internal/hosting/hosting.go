package hosting

import "context"

// MergeRequest represents a pending MR discovered from any hosting backend.
// Fields that require a hosting API (Title, Labels) may be empty when using the
// generic git backend.
type MergeRequest struct {
	// ID is an opaque identifier for the MR: a branch name for the git backend,
	// or a numeric ID / URL for hosting-API backends.
	ID           string   `json:"id"`
	Branch       string   `json:"branch"`
	TargetBranch string   `json:"target_branch"`
	Author       string   `json:"author"`
	Title        string   `json:"title"`
	Labels       []string `json:"labels,omitempty"`
}

// HostingService is the interface every backend must implement.
// See internal/hosting/git for the only current implementation.
// TODO: authorPattern is a confusing name, it should be less specific like "substring"
// or maybe a generic "searchParameters" struct that we pass and that gets handled
// depending on what is set
type HostingService interface {
	// ListMergeRequests returns open MRs whose inferred author matches
	// authorPattern (case-insensitive substring). An empty authorPattern
	// returns all discovered MRs.
	ListMergeRequests(ctx context.Context, authorPattern string) ([]MergeRequest, error)
}
