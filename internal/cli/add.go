package cli

import (
	"fmt"

	"github.com/frostyard/plow/internal/repo"
	"github.com/spf13/cobra"
)

var (
	addDist string
)

var addCmd = &cobra.Command{
	Use:   "add <deb-file>",
	Short: "Add a .deb package to the repository",
	Long: `Adds a .deb package to the repository pool, updates the package index,
and optionally prunes old versions.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		debPath := args[0]

		cfg := repo.DefaultConfig()
		r := repo.New(repoRoot, cfg)

		// Add the package
		pkg, err := r.AddPackage(debPath, addDist)
		if err != nil {
			return fmt.Errorf("add package: %w", err)
		}

		fmt.Printf("Added package: %s %s (%s)\n", pkg.Name, pkg.Version, pkg.Architecture)
		fmt.Printf("  Pool path: %s\n", pkg.Filename)

		// Prune old versions
		if keepVersions > 0 {
			result, err := r.Prune(repo.PruneOptions{
				KeepVersions: keepVersions,
			})
			if err != nil {
				return fmt.Errorf("prune: %w", err)
			}
			if len(result.Deleted) > 0 {
				fmt.Printf("  Pruned %d old version(s)\n", len(result.Deleted))
			}
		}

		// Regenerate index
		if err := r.GeneratePackagesIndex(addDist); err != nil {
			return fmt.Errorf("generate packages index: %w", err)
		}
		fmt.Printf("  Updated Packages index for %s\n", addDist)

		// Generate Release
		if err := r.GenerateRelease(addDist); err != nil {
			return fmt.Errorf("generate release: %w", err)
		}
		fmt.Printf("  Updated Release for %s\n", addDist)

		return nil
	},
}

func init() {
	addCmd.Flags().StringVarP(&addDist, "dist", "d", "stable", "Distribution to add the package to (stable, testing)")
	rootCmd.AddCommand(addCmd)
}
