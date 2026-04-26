package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile       string
	nonInteractive bool
)

var rootCmd = &cobra.Command{
	Use:   "lich",
	Short: "Lich: kubernetes GitOps workflow assistant",
	Long: `Lich automates the GitOps workflow for kustomize-based Kubernetes manifests.

It detects renovate MRs/PRs, rerenders manifests via kustomize, and manages
the git workflow (commit, rebase, push) so the rendered output stays in sync
with upstream chart updates.

External dependencies: git, kustomize`,
}

// Execute is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default: $HOME/.config/lich/config.yaml or ./lichconfig.yaml)")
	rootCmd.PersistentFlags().BoolVar(&nonInteractive, "no-interactive", false,
		"disable interactive prompts; auto-enabled when stdin is not a TTY")
}

// initConfig sets up viper for config file and environment variable support.
// Config loading is best-effort: a missing config file is not an error.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		if home, err := os.UserHomeDir(); err == nil {
			viper.AddConfigPath(home + "/.config/lich")
		}
		viper.AddConfigPath(".")
		viper.SetConfigName("lichconfig")
		viper.SetConfigType("yaml")
	}

	// LICH_* environment variables override config file values.
	viper.SetEnvPrefix("LICH")
	viper.AutomaticEnv()

	// read config based on previously set parameters
	// Intentionally ignore errors, since config is optional.
	_ = viper.ReadInConfig()
}

// IsInteractive reports whether the tool is running in an interactive terminal.
// The --no-interactive flag overrides the auto-detection.
func IsInteractive() bool {
	if nonInteractive {
		return false
	}
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
