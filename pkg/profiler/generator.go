package profiler

import (
	"sort"
	"strings"
)

// Options controls the breadth and constraints of the generation engine.
// The zero value is not directly usable for Suffixes/Separators; call
// DefaultOptions() to get sane defaults and override individual fields.
type Options struct {
	// Suffixes are appended to single tokens and to pairwise combinations,
	// e.g. "123", "!", "2026". Include "" in the list to also keep the
	// un-suffixed form.
	Suffixes []string

	// Separators are used when gluing two distinct tokens together, e.g.
	// "", "_", ".", "-".
	Separators []string

	// MinLength discards any candidate shorter than this many characters.
	MinLength int

	// MaxLength discards any candidate longer than this many characters.
	// 0 means unlimited.
	MaxLength int

	// IncludePairs enables pairwise combination of distinct tokens (e.g.
	// PetName+PartnerName). Disabling this drastically reduces output size
	// and runtime for large profiles.
	IncludePairs bool

	// MaxCandidates caps the number of returned candidates after sorting.
	// 0 means unlimited. This is a safety valve against combinatorial
	// explosion on profiles with many populated fields.
	MaxCandidates int

	// IncludeToggleCase enables per-character case toggling, producing all
	// 2^n upper/lower combinations for each letter token (e.g. "john",
	// "John", "jOhn", "JOhn", ...). Can significantly increase output size
	// for long tokens.
	IncludeToggleCase bool

	// IncludeLeet enables leet-speak substitutions (a→@/4, e→3, i→1/!,
	// o→0, s→$/5, t→7, …) applied in 1-, 2-, and 3-rule combinations.
	// This models the common habit of swapping letters to satisfy "complexity"
	// requirements: "j0hn", "j@ne", "p@$$w0rd", etc.
	IncludeLeet bool
}

// DefaultOptions returns a reasonable, common-habit-driven default
// configuration: a curated suffix list (sequential digits, punctuation,
// recent and near-future years), a handful of glue separators, and length
// bounds matching typical password policies (4-32 characters).
func DefaultOptions() Options {
	return Options{
		Suffixes: []string{
			// bare form
			"",
			// pure digit runs — the most common suffix class
			"0", "1", "2", "01", "07", "12", "23", "69", "99", "00",
			"11", "22", "123", "007", "1234",
			// symbol-only (often added to satisfy complexity rules)
			"!", "!!", "?", "@", "#", "*",
			// number+symbol combos that satisfy complexity requirements
			"1!", "12!", "123!", "1234!", "@1", "@12", "@123", "#1", "#123",
			// years — current and recent
			"2020", "2021", "2022", "2023", "2024", "2025", "2026",
		},
		Separators:    []string{"", "_", ".", "-"},
		MinLength:     4,
		MaxLength:     32,
		IncludePairs:      true,
		IncludeToggleCase: true,
		IncludeLeet:       true,
		MaxCandidates:     0,
	}
}

// Generate runs the full mutation/combinatorics pipeline against the
// supplied Profile and returns a deduplicated, length-filtered, sorted slice
// of password candidates. The returned slice never contains duplicates.
//
// Pipeline overview:
//  1. Extract every non-empty token from the profile (names, nickname,
//     birth-date derivatives, etc.).
//  2. For each token, generate case variants, then append the configured
//     suffixes to each variant.
//  3. If IncludePairs is enabled, combine every distinct pair of tokens
//     using the configured separators, then apply case variants and
//     suffixes to each combination as well.
//  4. Deduplicate via a set, filter by length, sort for deterministic
//     output, and optionally truncate to MaxCandidates.
func Generate(p Profile, opts Options) []string {
	candidates := make(map[string]struct{})

	tokens := p.Tokens()

	// Stage 1: individual token mutations + suffixes.
	for _, t := range tokens {
		// Base case variants (lower, upper, title, original).
		variants := CaseVariants(t.Value)

		// Per-character toggle case (all 2^n upper/lower combos).
		if opts.IncludeToggleCase {
			variants = append(variants, ToggleCaseVariants(t.Value)...)
		}

		// Leet-speak substitution variants; apply CaseVariants to each so
		// "j0hn" also yields "J0hn", "J0HN", etc.
		if opts.IncludeLeet {
			for _, lv := range LeetVariants(t.Value) {
				variants = append(variants, CaseVariants(lv)...)
			}
		}

		// Deduplicate across all variant sources before suffix expansion.
		seen := make(map[string]struct{}, len(variants))
		for _, cv := range variants {
			if _, dup := seen[cv]; dup {
				continue
			}
			seen[cv] = struct{}{}
			addFiltered(candidates, cv, opts)
			for _, withSuffix := range AppendSuffixes(cv, opts.Suffixes) {
				addFiltered(candidates, withSuffix, opts)
			}
		}
	}

	// Stage 2: pairwise combinations of distinct tokens.
	if opts.IncludePairs {
		for i := 0; i < len(tokens); i++ {
			for j := i + 1; j < len(tokens); j++ {
				a := strings.ToLower(tokens[i].Value)
				b := strings.ToLower(tokens[j].Value)

				for _, combo := range Combine(a, b, opts.Separators) {
					addFiltered(candidates, combo, opts)
					for _, cv := range CaseVariants(combo) {
						addFiltered(candidates, cv, opts)
					}
					for _, withSuffix := range AppendSuffixes(combo, opts.Suffixes) {
						addFiltered(candidates, withSuffix, opts)
					}
				}
			}
		}
	}

	out := make([]string, 0, len(candidates))
	for w := range candidates {
		out = append(out, w)
	}
	sort.Strings(out)

	if opts.MaxCandidates > 0 && len(out) > opts.MaxCandidates {
		out = out[:opts.MaxCandidates]
	}

	return out
}

// addFiltered inserts s into the candidate set only if it satisfies the
// configured length constraints. Centralizing the check here keeps the
// filtering logic in exactly one place.
func addFiltered(set map[string]struct{}, s string, opts Options) {
	if s == "" {
		return
	}
	if opts.MinLength > 0 && len(s) < opts.MinLength {
		return
	}
	if opts.MaxLength > 0 && len(s) > opts.MaxLength {
		return
	}
	set[s] = struct{}{}
}
