// Package repo manages Debian repository structure and metadata.
package repo

import (
	"compress/gzip"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/frostyard/plow/internal/deb"
)

// Config holds repository configuration.
type Config struct {
	Origin        string
	Label         string
	Description   string
	Architectures []string
	Components    []string
	Distributions []string
}

// DefaultConfig returns the default repository configuration.
func DefaultConfig() Config {
	return Config{
		Origin:        "Frostyard",
		Label:         "Frostyard",
		Description:   "Frostyard Debian Repository",
		Architectures: []string{"amd64"},
		Components:    []string{"main"},
		Distributions: []string{"stable", "testing"},
	}
}

// Repository represents a Debian repository on disk.
type Repository struct {
	Root   string
	Config Config
}

// New creates a new Repository instance.
func New(root string, cfg Config) *Repository {
	return &Repository{
		Root:   root,
		Config: cfg,
	}
}

// Init creates the initial directory structure for the repository.
func (r *Repository) Init() error {
	for _, dist := range r.Config.Distributions {
		for _, comp := range r.Config.Components {
			for _, arch := range r.Config.Architectures {
				dir := filepath.Join(r.Root, "dists", dist, comp, "binary-"+arch)
				if err := os.MkdirAll(dir, 0755); err != nil {
					return fmt.Errorf("create directory %s: %w", dir, err)
				}

				// Create empty Packages file
				packagesPath := filepath.Join(dir, "Packages")
				if _, err := os.Stat(packagesPath); os.IsNotExist(err) {
					if err := os.WriteFile(packagesPath, []byte{}, 0644); err != nil {
						return fmt.Errorf("create Packages file: %w", err)
					}
				}
			}
		}
	}

	// Create pool directory
	poolDir := filepath.Join(r.Root, "pool", "main")
	if err := os.MkdirAll(poolDir, 0755); err != nil {
		return fmt.Errorf("create pool directory: %w", err)
	}

	return nil
}

// AddPackage adds a .deb file to the repository.
// It copies the file to the pool and updates the package index.
func (r *Repository) AddPackage(debPath, dist string) (*deb.Package, error) {
	pkg, err := deb.Parse(debPath)
	if err != nil {
		return nil, fmt.Errorf("parse deb: %w", err)
	}

	// Determine destination path in pool
	filename := filepath.Base(debPath)
	poolPath := pkg.PoolPath(filename)
	fullPoolPath := filepath.Join(r.Root, poolPath)

	// Create directory
	if err := os.MkdirAll(filepath.Dir(fullPoolPath), 0755); err != nil {
		return nil, fmt.Errorf("create pool directory: %w", err)
	}

	// Copy file
	if err := copyFile(debPath, fullPoolPath); err != nil {
		return nil, fmt.Errorf("copy deb to pool: %w", err)
	}

	// Set the filename for the package index
	pkg.Filename = poolPath

	return pkg, nil
}

// GeneratePackagesIndex generates the Packages, Packages.gz, and Packages.xz files
// for a given distribution.
func (r *Repository) GeneratePackagesIndex(dist string) error {
	for _, comp := range r.Config.Components {
		for _, arch := range r.Config.Architectures {
			if err := r.generatePackagesForArch(dist, comp, arch); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Repository) generatePackagesForArch(dist, comp, arch string) error {
	// Scan pool for packages of this architecture
	poolDir := filepath.Join(r.Root, "pool", comp)
	packages, err := r.scanPool(poolDir, arch)
	if err != nil {
		return fmt.Errorf("scan pool: %w", err)
	}

	// Build Packages content
	var content strings.Builder
	for _, pkg := range packages {
		content.WriteString(pkg.ControlString())
		content.WriteString("\n")
	}

	// Write Packages file
	distDir := filepath.Join(r.Root, "dists", dist, comp, "binary-"+arch)
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return fmt.Errorf("create dist directory: %w", err)
	}

	packagesPath := filepath.Join(distDir, "Packages")
	if err := os.WriteFile(packagesPath, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("write Packages: %w", err)
	}

	// Generate Packages.gz
	gzPath := packagesPath + ".gz"
	gzFile, err := os.Create(gzPath)
	if err != nil {
		return fmt.Errorf("create Packages.gz: %w", err)
	}
	gzWriter := gzip.NewWriter(gzFile)
	if _, err := gzWriter.Write([]byte(content.String())); err != nil {
		gzWriter.Close()
		gzFile.Close()
		return fmt.Errorf("write Packages.gz: %w", err)
	}
	gzWriter.Close()
	gzFile.Close()

	// Generate Packages.xz (using xz command)
	xzPath := packagesPath + ".xz"
	cmd := exec.Command("xz", "-k", "-f", packagesPath)
	if err := cmd.Run(); err != nil {
		// xz might not be installed, that's okay
		fmt.Fprintf(os.Stderr, "Warning: could not create %s: %v\n", xzPath, err)
	}

	return nil
}

func (r *Repository) scanPool(poolDir, arch string) ([]*deb.Package, error) {
	var packages []*deb.Package

	err := filepath.Walk(poolDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".deb") {
			return nil
		}

		pkg, err := deb.Parse(path)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}

		// Filter by architecture
		if pkg.Architecture != arch && pkg.Architecture != "all" {
			return nil
		}

		// Set relative filename
		relPath, err := filepath.Rel(r.Root, path)
		if err != nil {
			return err
		}
		pkg.Filename = relPath

		packages = append(packages, pkg)
		return nil
	})

	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	// Sort packages by name, then version (newest first)
	sort.Slice(packages, func(i, j int) bool {
		if packages[i].Name != packages[j].Name {
			return packages[i].Name < packages[j].Name
		}
		return deb.Compare(packages[i].Version, packages[j].Version) > 0
	})

	return packages, nil
}

// GenerateRelease generates the Release file for a distribution.
func (r *Repository) GenerateRelease(dist string) error {
	distDir := filepath.Join(r.Root, "dists", dist)

	// Collect all files that need checksums
	var files []releaseFile
	err := filepath.Walk(distDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		name := filepath.Base(path)
		// Only include index files
		if !strings.HasPrefix(name, "Packages") && !strings.HasPrefix(name, "Release") {
			return nil
		}
		// Skip the Release file itself
		if name == "Release" || name == "Release.gpg" || name == "InRelease" {
			return nil
		}

		relPath, _ := filepath.Rel(distDir, path)
		rf, err := newReleaseFile(path, relPath)
		if err != nil {
			return err
		}
		files = append(files, rf)
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk dist directory: %w", err)
	}

	// Build Release content
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Origin: %s\n", r.Config.Origin))
	b.WriteString(fmt.Sprintf("Label: %s\n", r.Config.Label))
	b.WriteString(fmt.Sprintf("Suite: %s\n", dist))
	b.WriteString(fmt.Sprintf("Codename: %s\n", dist))
	b.WriteString(fmt.Sprintf("Architectures: %s\n", strings.Join(r.Config.Architectures, " ")))
	b.WriteString(fmt.Sprintf("Components: %s\n", strings.Join(r.Config.Components, " ")))
	b.WriteString(fmt.Sprintf("Description: %s\n", r.Config.Description))
	b.WriteString(fmt.Sprintf("Date: %s\n", time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 UTC")))

	// MD5Sum
	b.WriteString("MD5Sum:\n")
	for _, f := range files {
		b.WriteString(fmt.Sprintf(" %s %16d %s\n", f.MD5, f.Size, f.Path))
	}

	// SHA1
	b.WriteString("SHA1:\n")
	for _, f := range files {
		b.WriteString(fmt.Sprintf(" %s %16d %s\n", f.SHA1, f.Size, f.Path))
	}

	// SHA256
	b.WriteString("SHA256:\n")
	for _, f := range files {
		b.WriteString(fmt.Sprintf(" %s %16d %s\n", f.SHA256, f.Size, f.Path))
	}

	releasePath := filepath.Join(distDir, "Release")
	if err := os.WriteFile(releasePath, []byte(b.String()), 0644); err != nil {
		return fmt.Errorf("write Release: %w", err)
	}

	return nil
}

type releaseFile struct {
	Path   string
	Size   int64
	MD5    string
	SHA1   string
	SHA256 string
}

func newReleaseFile(fullPath, relPath string) (releaseFile, error) {
	f, err := os.Open(fullPath)
	if err != nil {
		return releaseFile{}, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return releaseFile{}, err
	}

	md5h := md5.New()
	sha1h := sha1.New()
	sha256h := sha256.New()
	multi := io.MultiWriter(md5h, sha1h, sha256h)

	if _, err := io.Copy(multi, f); err != nil {
		return releaseFile{}, err
	}

	return releaseFile{
		Path:   relPath,
		Size:   stat.Size(),
		MD5:    fmt.Sprintf("%x", md5h.Sum(nil)),
		SHA1:   fmt.Sprintf("%x", sha1h.Sum(nil)),
		SHA256: fmt.Sprintf("%x", sha256h.Sum(nil)),
	}, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return out.Sync()
}
