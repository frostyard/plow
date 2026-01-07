package cli

import (
	"fmt"
	"path/filepath"

	"github.com/frostyard/plow/internal/gpg"
	"github.com/spf13/cobra"
)

var (
	signDist  string
	signKeyID string
)

var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Sign the repository Release file",
	Long:  `Signs the Release file, creating Release.gpg (detached) and InRelease (inline).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		distDir := filepath.Join(repoRoot, "dists", signDist)

		signer := gpg.NewSigner(signKeyID)
		if err := signer.SignRelease(distDir); err != nil {
			return fmt.Errorf("sign release: %w", err)
		}

		fmt.Printf("Signed Release for %s\n", signDist)
		fmt.Printf("  Created: %s/Release.gpg\n", distDir)
		fmt.Printf("  Created: %s/InRelease\n", distDir)

		return nil
	},
}

func init() {
	signCmd.Flags().StringVarP(&signDist, "dist", "d", "stable", "Distribution to sign")
	signCmd.Flags().StringVarP(&signKeyID, "key", "k", "", "GPG key ID to use for signing")
	rootCmd.AddCommand(signCmd)
}
