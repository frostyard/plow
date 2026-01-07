package deb

import (
	"strconv"
	"strings"
	"unicode"
)

// Compare compares two Debian version strings.
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
// Implements the Debian version comparison algorithm.
func Compare(a, b string) int {
	epochA, upstreamA, revisionA := ParseVersion(a)
	epochB, upstreamB, revisionB := ParseVersion(b)

	// Compare epochs
	if epochA != epochB {
		if epochA < epochB {
			return -1
		}
		return 1
	}

	// Compare upstream versions
	if cmp := compareVersionPart(upstreamA, upstreamB); cmp != 0 {
		return cmp
	}

	// Compare revisions
	return compareVersionPart(revisionA, revisionB)
}

// compareVersionPart compares version parts using Debian's algorithm.
// The algorithm splits the string into alternating non-digit and digit parts,
// comparing them appropriately.
func compareVersionPart(a, b string) int {
	for len(a) > 0 || len(b) > 0 {
		// Extract and compare non-digit parts
		nonDigitA, restA := extractNonDigits(a)
		nonDigitB, restB := extractNonDigits(b)

		if cmp := compareNonDigits(nonDigitA, nonDigitB); cmp != 0 {
			return cmp
		}

		a, b = restA, restB

		// Extract and compare digit parts
		digitA, restA := extractDigits(a)
		digitB, restB := extractDigits(b)

		if cmp := compareDigits(digitA, digitB); cmp != 0 {
			return cmp
		}

		a, b = restA, restB
	}

	return 0
}

func extractNonDigits(s string) (string, string) {
	i := 0
	for i < len(s) && !unicode.IsDigit(rune(s[i])) {
		i++
	}
	return s[:i], s[i:]
}

func extractDigits(s string) (string, string) {
	i := 0
	for i < len(s) && unicode.IsDigit(rune(s[i])) {
		i++
	}
	return s[:i], s[i:]
}

// compareNonDigits compares non-digit parts according to Debian rules.
// Letters sort before non-letters (except ~), and ~ sorts before everything.
func compareNonDigits(a, b string) int {
	for i := 0; i < len(a) || i < len(b); i++ {
		var ca, cb int
		if i < len(a) {
			ca = charOrder(a[i])
		}
		if i < len(b) {
			cb = charOrder(b[i])
		}
		if ca != cb {
			if ca < cb {
				return -1
			}
			return 1
		}
	}
	return 0
}

// charOrder returns the sort order for a character in Debian version comparison.
// ~ sorts before everything (including empty)
// Letters sort before non-letters
func charOrder(c byte) int {
	switch {
	case c == '~':
		return -1
	case c >= 'A' && c <= 'Z':
		return int(c)
	case c >= 'a' && c <= 'z':
		return int(c)
	default:
		return int(c) + 256
	}
}

// compareDigits compares digit strings numerically.
func compareDigits(a, b string) int {
	// Trim leading zeros for numeric comparison
	a = strings.TrimLeft(a, "0")
	b = strings.TrimLeft(b, "0")

	if a == "" {
		a = "0"
	}
	if b == "" {
		b = "0"
	}

	numA, _ := strconv.ParseInt(a, 10, 64)
	numB, _ := strconv.ParseInt(b, 10, 64)

	if numA < numB {
		return -1
	}
	if numA > numB {
		return 1
	}
	return 0
}

// SortVersions sorts a slice of version strings in descending order (newest first).
func SortVersions(versions []string) {
	for i := 0; i < len(versions)-1; i++ {
		for j := i + 1; j < len(versions); j++ {
			if Compare(versions[i], versions[j]) < 0 {
				versions[i], versions[j] = versions[j], versions[i]
			}
		}
	}
}
