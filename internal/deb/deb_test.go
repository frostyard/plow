package deb

import "testing"

func TestPackagePoolPath(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{"myapp", "myapp_1.0.0_amd64.deb", "pool/main/m/myapp/myapp_1.0.0_amd64.deb"},
		{"libc", "libc_1.0.0_amd64.deb", "pool/main/libc/libc/libc_1.0.0_amd64.deb"},
		{"libfoo", "libfoo_1.0.0_amd64.deb", "pool/main/libf/libfoo/libfoo_1.0.0_amd64.deb"},
		{"apache2", "apache2_2.4.0_amd64.deb", "pool/main/a/apache2/apache2_2.4.0_amd64.deb"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pkg := &Package{Name: tc.name}
			result := pkg.PoolPath(tc.filename)
			if result != tc.expected {
				t.Errorf("PoolPath(%q) = %q, want %q", tc.filename, result, tc.expected)
			}
		})
	}
}

func TestPackageDebFilename(t *testing.T) {
	pkg := &Package{
		Name:         "myapp",
		Version:      "1.0.0",
		Architecture: "amd64",
	}

	expected := "myapp_1.0.0_amd64.deb"
	if result := pkg.DebFilename(); result != expected {
		t.Errorf("DebFilename() = %q, want %q", result, expected)
	}
}

func TestPackageControlString(t *testing.T) {
	pkg := &Package{
		Name:         "myapp",
		Version:      "1.0.0",
		Architecture: "amd64",
		Maintainer:   "Test <test@example.com>",
		Description:  "A test package",
		Size:         1024,
		MD5sum:       "abc123",
		SHA1:         "def456",
		SHA256:       "ghi789",
		Filename:     "pool/main/m/myapp/myapp_1.0.0_amd64.deb",
	}

	control := pkg.ControlString()

	// Check required fields are present
	required := []string{
		"Package: myapp",
		"Version: 1.0.0",
		"Architecture: amd64",
		"Filename: pool/main/m/myapp/myapp_1.0.0_amd64.deb",
		"Size: 1024",
		"MD5sum: abc123",
		"SHA1: def456",
		"SHA256: ghi789",
	}

	for _, field := range required {
		if !containsLine(control, field) {
			t.Errorf("ControlString() missing %q", field)
		}
	}
}

func containsLine(s, substr string) bool {
	for _, line := range splitLines(s) {
		if line == substr {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
