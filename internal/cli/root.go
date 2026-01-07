package cli

import (
	"github.com/spf13/cobra"
)

var (
	repoRoot     string
	keepVersions int
)

func Execute() error {
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "plow",
	Short: "Plow - Debian repository manager",
	Long: `Plow is a tool for managing Debian package repositories.

It handles adding packages, generating repository metadata (Packages, Release),
signing with GPG, and pruning old package versions.`,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&repoRoot, "repo-root", "r", ".", "Path to repository root")
	rootCmd.PersistentFlags().IntVar(&keepVersions, "keep-versions", 5, "Number of versions to keep per package when pruning")
}
