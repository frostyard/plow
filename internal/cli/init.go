package cli

import (
	"fmt"

	"github.com/frostyard/plow/internal/repo"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize repository directory structure",
	Long:  `Creates the initial directory structure for a Debian repository.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := repo.DefaultConfig()
		r := repo.New(repoRoot, cfg)

		if err := r.Init(); err != nil {
			return fmt.Errorf("initialize repository: %w", err)
		}

		if err := r.GenerateHTMLIndexes(); err != nil {
			return fmt.Errorf("generate HTML indexes: %w", err)
		}

		fmt.Println("Repository initialized successfully")
		fmt.Printf("  Root: %s\n", repoRoot)
		fmt.Printf("  Distributions: %v\n", cfg.Distributions)
		fmt.Printf("  Components: %v\n", cfg.Components)
		fmt.Printf("  Architectures: %v\n", cfg.Architectures)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
