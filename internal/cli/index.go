package cli

import (
	"fmt"

	"github.com/frostyard/plow/internal/repo"
	"github.com/spf13/cobra"
)

var (
	indexDist string
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Regenerate repository index files",
	Long:  `Regenerates the Packages and Release files for a distribution, and generates HTML index pages for browser-friendly navigation.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := repo.DefaultConfig()
		r := repo.New(repoRoot, cfg)

		if err := r.GeneratePackagesIndex(indexDist); err != nil {
			return fmt.Errorf("generate packages index: %w", err)
		}
		fmt.Printf("Generated Packages index for %s\n", indexDist)

		if err := r.GenerateRelease(indexDist); err != nil {
			return fmt.Errorf("generate release: %w", err)
		}
		fmt.Printf("Generated Release for %s\n", indexDist)

		if err := r.GenerateHTMLIndexes(); err != nil {
			return fmt.Errorf("generate HTML indexes: %w", err)
		}
		fmt.Println("Generated HTML index pages")

		return nil
	},
}

func init() {
	indexCmd.Flags().StringVarP(&indexDist, "dist", "d", "stable", "Distribution to regenerate index for")
	rootCmd.AddCommand(indexCmd)
}
