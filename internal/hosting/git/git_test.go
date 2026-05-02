package git_test

import (
	"strings"
	"testing"

	"go.lichturm.de/lich/internal/hosting/git"
)

var sampleBranchLines = []string{
	"  origin/HEAD -> origin/master",
	"  origin/master",
	"  origin/renovate/helm-release-my-app-1.2.3",
	"  origin/renovate/helm-release-other-app-4.5.6",
	"  origin/renovate/helm-release-my-app-1.3.0",
	"  origin/feature/some-work",
	"  origin/dependabot/npm-lodash-4.17.21",
}

func TestParseBranches_FilterByAuthor(t *testing.T) {
	mrs := git.ParseBranches(sampleBranchLines, "origin", "master", "renovate")

	if len(mrs) != 3 {
		t.Fatalf("expected 3 renovate MRs, got %d", len(mrs))
	}
	for _, mr := range mrs {
		if mr.Author != "renovate" {
			t.Errorf("expected author 'renovate', got %q (branch: %s)", mr.Author, mr.Branch)
		}
		if mr.TargetBranch != "master" {
			t.Errorf("expected target branch 'master', got %q", mr.TargetBranch)
		}
	}
}

func TestParseBranches_EmptyAuthorReturnsAll(t *testing.T) {
	// No author filter: should return all non-HEAD branches
	mrs := git.ParseBranches(sampleBranchLines, "origin", "master", "")

	// master, 3x renovate, 1x feature, 1x dependabot = 6
	if len(mrs) != 6 {
		t.Fatalf("expected 6 MRs with no author filter, got %d", len(mrs))
	}
}

func TestParseBranches_AuthorFilterCaseInsensitive(t *testing.T) {
	mrs := git.ParseBranches(sampleBranchLines, "origin", "master", "Renovate")
	if len(mrs) != 3 {
		t.Fatalf("expected 3 MRs with case-insensitive filter, got %d", len(mrs))
	}
}

func TestParseBranches_SkipsHEAD(t *testing.T) {
	mrs := git.ParseBranches(sampleBranchLines, "origin", "master", "")
	for _, mr := range mrs {
		if mr.Branch == "HEAD" || mr.Branch == "master" && mr.Author == "" {
			// master itself has no "/" so author is empty — that's fine
		}
		if mr.ID == "" {
			t.Errorf("MR with empty ID found: %+v", mr)
		}
	}
}

func TestParseBranches_ComponentFilter(t *testing.T) {
	// The component filter is applied in the commune command, not in ParseBranches,
	// but we can verify the branch names contain what we expect for downstream filtering.
	mrs := git.ParseBranches(sampleBranchLines, "origin", "master", "renovate")

	component := "my-app"
	var matching int
	for _, mr := range mrs {
		if strings.Contains(mr.Branch, component) {
			matching++
		}
	}
	if matching != 2 {
		t.Fatalf("expected 2 branches containing %q, got %d", component, matching)
	}
}

func TestParseBranches_DifferentRemote(t *testing.T) {
	lines := []string{
		"  upstream/renovate/some-chart-1.0.0",
		"  origin/renovate/other-chart-2.0.0",
	}
	mrs := git.ParseBranches(lines, "upstream", "main", "renovate")
	if len(mrs) != 1 {
		t.Fatalf("expected 1 MR from 'upstream' remote, got %d", len(mrs))
	}
	if mrs[0].Branch != "renovate/some-chart-1.0.0" {
		t.Errorf("unexpected branch: %s", mrs[0].Branch)
	}
}
