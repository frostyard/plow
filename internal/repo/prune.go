package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/frostyard/plow/internal/deb"
)

// PruneOptions configures the prune operation.
type PruneOptions struct {
	KeepVersions int  // Number of versions to keep per package
	DryRun       bool // If true, only report what would be deleted
}

// PruneResult contains the result of a prune operation.
type PruneResult struct {
	Deleted []string // Paths of deleted files
	Kept    []string // Paths of kept files
}

// Prune removes old package versions, keeping only the newest N versions.
func (r *Repository) Prune(opts PruneOptions) (*PruneResult, error) {
	if opts.KeepVersions < 1 {
		opts.KeepVersions = 5
	}

	poolDir := filepath.Join(r.Root, "pool", "main")
	result := &PruneResult{}

	// Group packages by name and architecture
	packages := make(map[string][]*packageFile)

	err := filepath.Walk(poolDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".deb") {
			return nil
		}

		pkg, err := deb.Parse(path)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}

		key := pkg.Name + "_" + pkg.Architecture
		packages[key] = append(packages[key], &packageFile{
			Path:    path,
			Version: pkg.Version,
		})
		return nil
	})

	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	// For each package, sort by version and prune old ones
	for _, pkgs := range packages {
		// Sort by version, newest first
		sortPackageFiles(pkgs)

		for i, pf := range pkgs {
			if i < opts.KeepVersions {
				result.Kept = append(result.Kept, pf.Path)
			} else {
				result.Deleted = append(result.Deleted, pf.Path)
				if !opts.DryRun {
					if err := os.Remove(pf.Path); err != nil {
						return nil, fmt.Errorf("delete %s: %w", pf.Path, err)
					}
				}
			}
		}
	}

	// Clean up empty directories
	if !opts.DryRun {
		if err := cleanEmptyDirs(poolDir); err != nil {
			return nil, fmt.Errorf("clean empty directories: %w", err)
		}
	}

	return result, nil
}

type packageFile struct {
	Path    string
	Version string
}

func sortPackageFiles(files []*packageFile) {
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if deb.Compare(files[i].Version, files[j].Version) < 0 {
				files[i], files[j] = files[j], files[i]
			}
		}
	}
}

func cleanEmptyDirs(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}

		if len(entries) == 0 && path != root {
			return os.Remove(path)
		}
		return nil
	})
}
