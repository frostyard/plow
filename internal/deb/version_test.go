package deb

import "testing"

func TestCompare(t *testing.T) {
	tests := []struct {
		a, b     string
		expected int
	}{
		// Basic version comparison
		{"1.0", "2.0", -1},
		{"2.0", "1.0", 1},
		{"1.0", "1.0", 0},

		// Multi-part versions
		{"1.0.1", "1.0.2", -1},
		{"1.0.10", "1.0.9", 1},
		{"1.0.0", "1.0", 1}, // 1.0.0 > 1.0 because extra .0 component

		// With epoch
		{"1:1.0", "2.0", 1},
		{"2.0", "1:1.0", -1},
		{"1:1.0", "1:2.0", -1},
		{"2:1.0", "1:2.0", 1},

		// With revision
		{"1.0-1", "1.0-2", -1},
		{"1.0-2", "1.0-1", 1},
		{"1.0-1", "1.0-1", 0},

		// Full version strings
		{"1:1.0-1", "1:1.0-2", -1},
		{"2:1.0-1", "1:2.0-1", 1},

		// Tilde versions (sort before everything)
		{"1.0~beta", "1.0", -1},
		{"1.0~alpha", "1.0~beta", -1},
		{"1.0~1", "1.0", -1},

		// Alpha versions
		{"1.0a", "1.0b", -1},
		{"1.0alpha", "1.0beta", -1},

		// Mixed
		{"1.0.0~rc1", "1.0.0", -1},
		{"1.0.0", "1.0.0~rc1", 1},
	}

	for _, tc := range tests {
		t.Run(tc.a+"_vs_"+tc.b, func(t *testing.T) {
			result := Compare(tc.a, tc.b)
			if result != tc.expected {
				t.Errorf("Compare(%q, %q) = %d, want %d", tc.a, tc.b, result, tc.expected)
			}
		})
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		version  string
		epoch    int
		upstream string
		revision string
	}{
		{"1.0", 0, "1.0", ""},
		{"1.0-1", 0, "1.0", "1"},
		{"1:1.0", 1, "1.0", ""},
		{"1:1.0-1", 1, "1.0", "1"},
		{"2:1.0.0-ubuntu1", 2, "1.0.0", "ubuntu1"},
	}

	for _, tc := range tests {
		t.Run(tc.version, func(t *testing.T) {
			epoch, upstream, revision := ParseVersion(tc.version)
			if epoch != tc.epoch {
				t.Errorf("epoch = %d, want %d", epoch, tc.epoch)
			}
			if upstream != tc.upstream {
				t.Errorf("upstream = %q, want %q", upstream, tc.upstream)
			}
			if revision != tc.revision {
				t.Errorf("revision = %q, want %q", revision, tc.revision)
			}
		})
	}
}

func TestSortVersions(t *testing.T) {
	versions := []string{"1.0", "2.0", "1.5", "1.0~rc1", "2.0.1"}
	SortVersions(versions)

	expected := []string{"2.0.1", "2.0", "1.5", "1.0", "1.0~rc1"}
	for i, v := range versions {
		if v != expected[i] {
			t.Errorf("versions[%d] = %q, want %q", i, v, expected[i])
		}
	}
}
