package kustomize

import (
	"fmt"
	"os"
	"path/filepath"

	"go.lichturm.de/lich/internal/lich"
)


func FindKustomization(unit lich.UnitOfWork) (lich.UnitOfWork, error) {
	candidates := []string{"kustomization.yaml", "kustomization.yml", "Kustomization"}
	for _, filename := range candidates {
		path := filepath.Join(unit.Origin, filename)

		_, err := os.Stat(path)
		if err != nil {
			// if the issue is not the file's absence, we might have permission problems
			if !os.IsNotExist(err) {
				return lich.UnitOfWork{}, err
				// if the issue is the file's absence, that might be okay
			} else {
				continue
			}
		}

		// We just assume that there is only one relevant kustomization
		// if there are any more, we skip them
		// TODO: Fix
		unit.Kustomization = path
		return unit, nil
	}
	return lich.UnitOfWork{}, fmt.Errorf("No kustomization file was found in %s", unit.Origin)
}
