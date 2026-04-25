package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// TODO: make configurable
const inputDir string = "components"
const outputDir string = "manifests"

var workBatch []UnitOfWork

// TODO: if we add fields for "processing", "completed", "success" and so on
// we can parallelise the templating
type UnitOfWork struct {
	name          string
	origin        string
	kustomization string
	destination   string
}

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
					UnitOfWork{
						name:        statResult.Name(),
						origin:      filepath.Join(inputDir, statResult.Name()),
						destination: filepath.Join(outputDir, statResult.Name()),
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
	Run: func(cmd *cobra.Command, args []string) {
		println("I am the run block")
		// for each unit of work, run kustomize build
	},
}

var (
	animateComponents []string
)

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

func findKustomization(unit UnitOfWork) (UnitOfWork, error) {
	candidates := []string{"kustomization.yaml", "kustomization.yml", "Kustomization"}
	for _, filename := range candidates {
		path := filepath.Join(unit.origin, filename)

		_, err := os.Stat(path)
		if err != nil {
			// if the issue is not the file's absence, we might have permission problems
			if !os.IsNotExist(err) {
				return UnitOfWork{}, err
				// if the issue is the file's absence, that might be okay
			} else {
				continue
			}
		}

		// We just assume that there is only one relevant kustomization
		// if there are any more, we skip them
		// TODO: Fix
		unit.kustomization = path
		return unit, nil
	}
	return UnitOfWork{}, fmt.Errorf("No kustomization file was found in %s", unit.origin)
}
