// Package deb provides utilities for parsing Debian package files.
package deb

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/blakesmith/ar"
)

// Package represents metadata extracted from a .deb file.
type Package struct {
	Name          string
	Version       string
	Architecture  string
	Maintainer    string
	Description   string
	Depends       string
	PreDepends    string
	Recommends    string
	Suggests      string
	Conflicts     string
	Provides      string
	Replaces      string
	Section       string
	Priority      string
	Homepage      string
	Size          int64  // File size in bytes
	InstalledSize int64  // Installed size in KB
	Filename      string // Relative path in pool
	MD5sum        string
	SHA1          string
	SHA256        string
}

// Parse reads a .deb file and extracts its metadata.
func Parse(path string) (*Package, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open deb: %w", err)
	}
	defer f.Close() //nolint:errcheck // Read-only file, close error is not critical

	// Get file size and checksums
	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat deb: %w", err)
	}

	// Calculate checksums
	md5h := md5.New()
	sha1h := sha1.New()
	sha256h := sha256.New()
	multiWriter := io.MultiWriter(md5h, sha1h, sha256h)

	if _, err := io.Copy(multiWriter, f); err != nil {
		return nil, fmt.Errorf("calculate checksums: %w", err)
	}

	// Seek back to beginning
	if _, err := f.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("seek: %w", err)
	}

	// Parse ar archive
	arReader := ar.NewReader(f)
	var controlData []byte

	for {
		header, err := arReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read ar: %w", err)
		}

		name := strings.TrimSuffix(header.Name, "/")

		// Look for control.tar.gz, control.tar.xz, or control.tar.zst
		if strings.HasPrefix(name, "control.tar") {
			controlData, err = extractControl(arReader, name)
			if err != nil {
				return nil, fmt.Errorf("extract control: %w", err)
			}
			break
		}
	}

	if controlData == nil {
		return nil, fmt.Errorf("control file not found in deb")
	}

	pkg, err := parseControl(controlData)
	if err != nil {
		return nil, fmt.Errorf("parse control: %w", err)
	}

	pkg.Size = stat.Size()
	pkg.MD5sum = hex.EncodeToString(md5h.Sum(nil))
	pkg.SHA1 = hex.EncodeToString(sha1h.Sum(nil))
	pkg.SHA256 = hex.EncodeToString(sha256h.Sum(nil))

	return pkg, nil
}

func extractControl(r io.Reader, archiveName string) ([]byte, error) {
	var tarReader *tar.Reader

	switch {
	case strings.HasSuffix(archiveName, ".gz"):
		gzr, err := gzip.NewReader(r)
		if err != nil {
			return nil, fmt.Errorf("open gzip: %w", err)
		}
		defer gzr.Close() //nolint:errcheck // Decompression complete, close error is not critical
		tarReader = tar.NewReader(gzr)
	case strings.HasSuffix(archiveName, ".xz"):
		// For xz, we'll shell out since Go doesn't have native xz support
		// Read all data first
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, fmt.Errorf("read xz data: %w", err)
		}
		return extractControlFromXz(data)
	case strings.HasSuffix(archiveName, ".zst"):
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, fmt.Errorf("read zst data: %w", err)
		}
		return extractControlFromZstd(data)
	default:
		// Assume uncompressed tar
		tarReader = tar.NewReader(r)
	}

	return findControlInTar(tarReader)
}

func findControlInTar(tarReader *tar.Reader) ([]byte, error) {
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read tar: %w", err)
		}

		name := strings.TrimPrefix(header.Name, "./")
		if name == "control" {
			return io.ReadAll(tarReader)
		}
	}
	return nil, fmt.Errorf("control file not found in tar")
}

func extractControlFromXz(data []byte) ([]byte, error) {
	// Write to temp file and use xz command
	tmpFile, err := os.CreateTemp("", "control.tar.xz")
	if err != nil {
		return nil, err
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName) //nolint:errcheck // Best effort cleanup

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return nil, err
	}
	if err := tmpFile.Close(); err != nil {
		return nil, err
	}

	return extractControlWithCmd("xz", []string{"-dk", "-c", tmpName})
}

func extractControlFromZstd(data []byte) ([]byte, error) {
	tmpFile, err := os.CreateTemp("", "control.tar.zst")
	if err != nil {
		return nil, err
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName) //nolint:errcheck // Best effort cleanup

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return nil, err
	}
	if err := tmpFile.Close(); err != nil {
		return nil, err
	}

	return extractControlWithCmd("zstd", []string{"-d", "-c", tmpName})
}

func extractControlWithCmd(cmd string, args []string) ([]byte, error) {
	// Import os/exec at runtime equivalent - use shell
	// This is a simplified version; in practice we'd use os/exec
	// For now, focus on gzip which is most common
	return nil, fmt.Errorf("xz/zstd decompression not implemented - use gzip")
}

func parseControl(data []byte) (*Package, error) {
	pkg := &Package{}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))

	var currentField string
	var currentValue strings.Builder

	saveField := func() {
		if currentField == "" {
			return
		}
		value := strings.TrimSpace(currentValue.String())
		switch currentField {
		case "Package":
			pkg.Name = value
		case "Version":
			pkg.Version = value
		case "Architecture":
			pkg.Architecture = value
		case "Maintainer":
			pkg.Maintainer = value
		case "Description":
			pkg.Description = value
		case "Depends":
			pkg.Depends = value
		case "Pre-Depends":
			pkg.PreDepends = value
		case "Recommends":
			pkg.Recommends = value
		case "Suggests":
			pkg.Suggests = value
		case "Conflicts":
			pkg.Conflicts = value
		case "Provides":
			pkg.Provides = value
		case "Replaces":
			pkg.Replaces = value
		case "Section":
			pkg.Section = value
		case "Priority":
			pkg.Priority = value
		case "Homepage":
			pkg.Homepage = value
		case "Installed-Size":
			if size, err := strconv.ParseInt(value, 10, 64); err == nil {
				pkg.InstalledSize = size
			}
		}
	}

	for scanner.Scan() {
		line := scanner.Text()

		// Continuation line (starts with space or tab)
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			currentValue.WriteString("\n")
			currentValue.WriteString(line)
			continue
		}

		// New field
		saveField()

		if idx := strings.Index(line, ":"); idx > 0 {
			currentField = line[:idx]
			currentValue.Reset()
			if idx+1 < len(line) {
				currentValue.WriteString(strings.TrimSpace(line[idx+1:]))
			}
		}
	}

	// Save last field
	saveField()

	if pkg.Name == "" {
		return nil, fmt.Errorf("missing Package field")
	}
	if pkg.Version == "" {
		return nil, fmt.Errorf("missing Version field")
	}
	if pkg.Architecture == "" {
		return nil, fmt.Errorf("missing Architecture field")
	}

	return pkg, nil
}

// ControlString returns the package in Packages file format.
func (p *Package) ControlString() string {
	var b strings.Builder

	writeField := func(name, value string) {
		if value != "" {
			b.WriteString(name)
			b.WriteString(": ")
			b.WriteString(value)
			b.WriteString("\n")
		}
	}

	writeField("Package", p.Name)
	writeField("Version", p.Version)
	writeField("Architecture", p.Architecture)
	writeField("Maintainer", p.Maintainer)
	if p.InstalledSize > 0 {
		writeField("Installed-Size", strconv.FormatInt(p.InstalledSize, 10))
	}
	writeField("Pre-Depends", p.PreDepends)
	writeField("Depends", p.Depends)
	writeField("Recommends", p.Recommends)
	writeField("Suggests", p.Suggests)
	writeField("Conflicts", p.Conflicts)
	writeField("Provides", p.Provides)
	writeField("Replaces", p.Replaces)
	writeField("Section", p.Section)
	writeField("Priority", p.Priority)
	writeField("Homepage", p.Homepage)
	writeField("Filename", p.Filename)
	writeField("Size", strconv.FormatInt(p.Size, 10))
	writeField("MD5sum", p.MD5sum)
	writeField("SHA1", p.SHA1)
	writeField("SHA256", p.SHA256)
	writeField("Description", p.Description)

	return b.String()
}

// PoolPath returns the relative path where this package should be stored in the pool.
// Format: pool/main/<first-letter>/<package-name>/<filename>
// For lib* packages: pool/main/lib<x>/<package-name>/<filename>
func (p *Package) PoolPath(filename string) string {
	var prefix string
	if strings.HasPrefix(p.Name, "lib") && len(p.Name) > 3 {
		prefix = p.Name[:4] // e.g., "liba", "libc"
	} else {
		prefix = p.Name[:1]
	}
	return filepath.Join("pool", "main", prefix, p.Name, filename)
}

// DebFilename returns the standard .deb filename for this package.
func (p *Package) DebFilename() string {
	return fmt.Sprintf("%s_%s_%s.deb", p.Name, p.Version, p.Architecture)
}

// epoch:upstream-revision version parsing
var versionRegex = regexp.MustCompile(`^(?:(\d+):)?([^-]+)(?:-(.+))?$`)

// ParseVersion parses a Debian version string into its components.
func ParseVersion(version string) (epoch int, upstream, revision string) {
	matches := versionRegex.FindStringSubmatch(version)
	if matches == nil {
		return 0, version, ""
	}

	if matches[1] != "" {
		epoch, _ = strconv.Atoi(matches[1])
	}
	upstream = matches[2]
	revision = matches[3]
	return
}
