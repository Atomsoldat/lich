package kustomize

import "os/exec"
import "context"
import 	"go.lichturm.de/lich/internal/lich"
import "fmt"
import "path/filepath"
import "os"

func Build(ctx context.Context, uow lich.UnitOfWork) error {
	// TODO: allow passing other flags
	cmd := exec.CommandContext(
		ctx,
		"kustomize",
		"build",
		"--enable-helm",
		"--output", filepath.Join(uow.Destination, "manifests.yaml"),
        uow.Origin,
	)

	err := os.MkdirAll(uow.Destination, 0750)
	if err != nil {
		return fmt.Errorf(
			"Could not create destination directory for unit of work %s: %w",
			uow.Name,
			err,
		)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"kustomize build for unit of work %s failed: %w\n%s",
			uow.Name,
			err,
			out,
		)
	}
	return nil
}
