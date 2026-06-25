package profiler

import (
	"strings"
	"unicode"
)

// CaseVariants returns the distinct casing transformations of s that humans
// commonly apply when crafting passwords: the value as supplied, all
// lowercase, all uppercase, and title case (first rune capitalized, the
// remainder lowercase). Duplicates that collapse together (e.g. a
// single-letter input) are removed. Returns nil for an empty input.
func CaseVariants(s string) []string {
	if s == "" {
		return nil
	}

	seen := map[string]struct{}{
		s:                  {},
		strings.ToLower(s): {},
		strings.ToUpper(s): {},
		titleCase(s):       {},
	}

	out := make([]string, 0, len(seen))
	for v := range seen {
		out = append(out, v)
	}
	return out
}

// titleCase upper-cases the first rune of s and lower-cases the rest, e.g.
// "jOHN" -> "John". Operates on runes so it is safe for non-ASCII names.
func titleCase(s string) string {
	if s == "" {
		return s
	}
	r := []rune(strings.ToLower(s))
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// AppendSuffixes returns base+suffix for every suffix in suffixes. An empty
// suffix in the list naturally reproduces the base value unchanged, which is
// intentional: it lets a single suffix list double as an "include the bare
// token" toggle.
func AppendSuffixes(base string, suffixes []string) []string {
	if base == "" || len(suffixes) == 0 {
		return nil
	}
	out := make([]string, 0, len(suffixes))
	for _, suf := range suffixes {
		out = append(out, base+suf)
	}
	return out
}

// Combine joins two non-empty strings using each separator in separators,
// producing both orderings (a+sep+b and b+sep+a). This models the very
// common habit of gluing together two pieces of personal significance, e.g.
// a pet's name and a partner's name: "max_sarah", "sarah_max", "maxsarah".
// Returns nil if either input is empty.
func Combine(a, b string, separators []string) []string {
	if a == "" || b == "" {
		return nil
	}
	out := make([]string, 0, len(separators)*2)
	for _, sep := range separators {
		out = append(out, a+sep+b, b+sep+a)
	}
	return out
}
