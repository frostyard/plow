package repo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateHTMLIndexes(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "plow-html-test")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create repository structure
	r := New(tmpDir, DefaultConfig())
	if err := r.Init(); err != nil {
		t.Fatalf("init repo: %v", err)
	}

	// Create some test files
	testDeb := filepath.Join(tmpDir, "pool", "main", "t", "testpkg", "testpkg_1.0.0_amd64.deb")
	if err := os.MkdirAll(filepath.Dir(testDeb), 0755); err != nil {
		t.Fatalf("create pool dir: %v", err)
	}
	if err := os.WriteFile(testDeb, []byte("fake deb content"), 0644); err != nil {
		t.Fatalf("write test deb: %v", err)
	}

	// Generate HTML indexes
	if err := r.GenerateHTMLIndexes(); err != nil {
		t.Fatalf("generate HTML indexes: %v", err)
	}

	// Check that index.html files were created
	expectedIndexes := []string{
		filepath.Join(tmpDir, "index.html"),
		filepath.Join(tmpDir, "pool", "index.html"),
		filepath.Join(tmpDir, "pool", "main", "index.html"),
		filepath.Join(tmpDir, "pool", "main", "t", "index.html"),
		filepath.Join(tmpDir, "pool", "main", "t", "testpkg", "index.html"),
		filepath.Join(tmpDir, "dists", "index.html"),
		filepath.Join(tmpDir, "dists", "stable", "index.html"),
		filepath.Join(tmpDir, "dists", "testing", "index.html"),
	}

	for _, indexPath := range expectedIndexes {
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			t.Errorf("expected index.html at %s", indexPath)
		}
	}

	// Check content of a generated index
	rootIndex, err := os.ReadFile(filepath.Join(tmpDir, "index.html"))
	if err != nil {
		t.Fatalf("read root index: %v", err)
	}

	content := string(rootIndex)
	if !strings.Contains(content, "Index of /") {
		t.Error("root index missing title")
	}
	if !strings.Contains(content, "pool/") {
		t.Error("root index missing pool link")
	}
	if !strings.Contains(content, "dists/") {
		t.Error("root index missing dists link")
	}

	// Check package directory index
	pkgIndex, err := os.ReadFile(filepath.Join(tmpDir, "pool", "main", "t", "testpkg", "index.html"))
	if err != nil {
		t.Fatalf("read package index: %v", err)
	}

	pkgContent := string(pkgIndex)
	if !strings.Contains(pkgContent, "testpkg_1.0.0_amd64.deb") {
		t.Error("package index missing deb file")
	}
	if !strings.Contains(pkgContent, "../") {
		t.Error("package index missing parent link")
	}
	if !strings.Contains(pkgContent, "📦") {
		t.Error("package index missing deb icon")
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		result := formatSize(tt.size)
		if result != tt.expected {
			t.Errorf("formatSize(%d) = %s, want %s", tt.size, result, tt.expected)
		}
	}
}

func TestIconForFile(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"package.deb", "📦"},
		{"Packages.gz", "🗜️"},
		{"Packages.xz", "🗜️"},
		{"Release.gpg", "🔑"},
		{"public.key", "🔑"},
		{"index.html", "🌐"},
		{"Release", "📄"},
		{"Packages", "📄"},
		{"random.txt", "📄"},
	}

	for _, tt := range tests {
		result := iconForFile(tt.name)
		if result != tt.expected {
			t.Errorf("iconForFile(%s) = %s, want %s", tt.name, result, tt.expected)
		}
	}
}
