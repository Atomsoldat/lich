// The git package implements hosting.HostingService using plain git branch analysis
// It requires only git to be installed and works with any remote hosting service
// No hosting API credentials are needed
//
// Limitations:
//   - Branch existence is used as a proxy for MR open state; already-merged
//     branches whose remote tracking ref has not been pruned will still appear.
//   - Title and Labels are not available without a hosting API; both are left
//     as the branch name and nil respectively.
package git

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"go.lichturm.de/lich/internal/hosting"
)

// Backend implements hosting.HostingService via git fetch + git branch -r.
type Backend struct {
	remote        string
	primaryBranch string
	repoPath      string
}

// New creates a Backend. repoPath is the path to the git repository root;
// pass "." to use the current working directory.
func New(remote, primaryBranch, repoPath string) *Backend {
	return &Backend{
		remote:        remote,
		primaryBranch: primaryBranch,
		repoPath:      repoPath,
	}
}

// ListMergeRequests fetches the remote and returns branches whose inferred
// author matches authorPattern. See hosting.HostingService for semantics.
func (b *Backend) ListMergeRequests(ctx context.Context, authorPattern string) ([]hosting.MergeRequest, error) {
	if err := b.fetch(ctx); err != nil {
		return nil, err
	}

	lines, err := b.remoteBranchLines(ctx)
	if err != nil {
		return nil, err
	}

	return ParseBranches(lines, b.remote, b.primaryBranch, authorPattern), nil
}

func (b *Backend) fetch(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "git", "-C", b.repoPath, "fetch", b.remote, "-q")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"git fetch %s failed (is the remote reachable and are credentials configured?): %w\n%s",
			b.remote, err, strings.TrimSpace(string(out)),
		)
	}
	return nil
}

func (b *Backend) remoteBranchLines(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", b.repoPath, "branch", "-r")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git branch -r failed: %w", err)
	}

	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// ParseBranches parses the output lines of "git branch -r" and returns
// MergeRequests matching authorPattern. Exported so it can be unit-tested
// without invoking git.
func ParseBranches(lines []string, remote, primaryBranch, authorPattern string) []hosting.MergeRequest {
	prefix := remote + "/"
	var mrs []hosting.MergeRequest

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip HEAD symbolic-ref lines, e.g. "origin/HEAD -> origin/master"
		if strings.Contains(line, "->") {
			continue
		}
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		branch := strings.TrimPrefix(line, prefix)
		author := inferAuthor(branch)

		if authorPattern != "" && !strings.Contains(strings.ToLower(author), strings.ToLower(authorPattern)) {
			continue
		}

		mrs = append(mrs, hosting.MergeRequest{
			ID:           branch,
			Branch:       branch,
			TargetBranch: primaryBranch,
			Author:       author,
			Title:        branch,
		})
	}
	return mrs
}

// TODO: This is really inflexible. We should just use a regex by default
// and give the user an option to override it, perhaps with some predefined ones as well
// inferAuthor derives an author name from the branch name prefix convention.
// For example, "renovate/helm-release-foo-1.2.3" yields "renovate".
// Returns an empty string if the branch has no "/" prefix.
func inferAuthor(branch string) string {
	if idx := strings.Index(branch, "/"); idx != -1 {
		return branch[:idx]
	}
	return ""
}
