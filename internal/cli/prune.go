package cli

import (
	"fmt"

	"github.com/frostyard/plow/internal/repo"
	"github.com/spf13/cobra"
)

var (
	pruneDryRun bool
)

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove old package versions",
	Long:  `Removes old package versions from the pool, keeping only the newest N versions per package.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := repo.DefaultConfig()
		r := repo.New(repoRoot, cfg)

		result, err := r.Prune(repo.PruneOptions{
			KeepVersions: keepVersions,
			DryRun:       pruneDryRun,
		})
		if err != nil {
			return fmt.Errorf("prune: %w", err)
		}

		if pruneDryRun {
			fmt.Println("Dry run - no files deleted")
		}

		fmt.Printf("Kept: %d packages\n", len(result.Kept))
		fmt.Printf("Deleted: %d packages\n", len(result.Deleted))

		if len(result.Deleted) > 0 {
			fmt.Println("\nDeleted packages:")
			for _, p := range result.Deleted {
				fmt.Printf("  - %s\n", p)
			}
		}

		return nil
	},
}

func init() {
	pruneCmd.Flags().BoolVarP(&pruneDryRun, "dry-run", "n", false, "Show what would be deleted without deleting")
	rootCmd.AddCommand(pruneCmd)
}
