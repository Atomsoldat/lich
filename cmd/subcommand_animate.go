package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.lichturm.de/lich/internal/kustomize"
	"go.lichturm.de/lich/internal/lich"
)

// TODO: make configurable
const inputDir string = "components"
const outputDir string = "manifests"

var (
	animateComponents []string
)

var workBatch []lich.UnitOfWork

var animateCmd = &cobra.Command{
	Use:     "animate [component]",
	Aliases: []string{"update"},
	Short:   "Render component manifests",
	Long: `animate executes "kustomize build" on the component subdirectory provided, or all of them, if none is specified.

Examples:
  lich animate
  lich animate my-application`,

	PreRunE: func(cmd *cobra.Command, args []string) error {

		// just a slice of basenames (i.e. last bit of the filename)
		inputDirContent := []string{}

		if len(args) == 0 {
			var err error
			// get all files in the input dir
			dirEntries, err := os.ReadDir(inputDir)
			if err != nil {
				return fmt.Errorf("Reading input dir failed: %w", err)
			}

			for _, item := range dirEntries {
				inputDirContent = append(inputDirContent, item.Name())
			}
		} else {
			// just look at the subdirs passed as arguments
			for _, item := range args {
				inputDirContent = append(inputDirContent, item)
			}

		}

		// Define units of work based on the input dir content we have discovered
		// TODO: we want to be able to ignore certain directories here
		// or maybe just in general
		for _, item := range inputDirContent {
			dirname := filepath.Join(inputDir, item)
			statResult, err := os.Stat(dirname)
			if err != nil {
				return fmt.Errorf("Reading dir %s failed: %w", dirname, err)
			}
			if statResult.IsDir() {
				workBatch = append(
					workBatch,
					lich.UnitOfWork{
						Name:        statResult.Name(),
						Origin:      filepath.Join(inputDir, statResult.Name()),
						Destination: filepath.Join(outputDir, statResult.Name()),
					},
				)
			} else {
				return fmt.Errorf("%s is not a directory", dirname)
			}
		}

		// check whether each unitOfWork contains a kustomization.yaml
		// TODO we should be able to pass a different filename
		for i, unit := range workBatch {
			var err error
			workBatch[i], err = findKustomization(unit)
			if err != nil {
				return err
			}
		}

		slog.Info("Verified subdirectories to process", "workBatch", workBatch)

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		println("I am the run block")
		// for each unit of work, run kustomize build in a separate context
		// subordinate to the main context
		// not sure what kind of context is best here yet
		// TODO: think about that
		for _, uow := range workBatch {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			err := kustomize.Build(ctx, uow)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	animateCmd.Flags().StringSliceVar(
		&animateComponents,
		"animate-components",
		// an empty slice
		[]string{},
		"component(s) to render",
	)

	rootCmd.AddCommand(animateCmd)
}

func findKustomization(unit lich.UnitOfWork) (lich.UnitOfWork, error) {
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
