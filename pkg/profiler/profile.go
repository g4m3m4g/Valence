// Package profiler implements the core mutation/combinatorics engine used to
// turn a structured personal profile into a list of realistic password
// candidates. It is intentionally dependency-free (standard library only) so
// it can be embedded in other tools or audited quickly.
package profiler

import (
	"fmt"
	"strings"
	"time"
)

// Token represents a single named, non-empty piece of information about a
// target that can be used as raw material for password candidates (e.g. a
// first name, a pet's name, or a derived birth-year string).
type Token struct {
	Label string // human-readable origin, useful for debugging/-verbose output
	Value string // the raw string value
}

// BirthDate is a minimal, parser-friendly representation of a date of birth.
// A zero-value BirthDate (all fields 0) is treated as "not supplied".
type BirthDate struct {
	Year  int
	Month int // 1-12
	Day   int // 1-31
}

// ParseBirthDate parses a date string in YYYY-MM-DD format (e.g. "1990-05-15")
// into a BirthDate. It returns a descriptive error on malformed input so the
// CLI layer can surface it directly to the operator.
func ParseBirthDate(s string) (BirthDate, error) {
	s = strings.TrimSpace(s)
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return BirthDate{}, fmt.Errorf("invalid birthdate %q, expected format YYYY-MM-DD: %w", s, err)
	}
	return BirthDate{Year: t.Year(), Month: int(t.Month()), Day: t.Day()}, nil
}

// IsZero reports whether the BirthDate carries no information.
func (b BirthDate) IsZero() bool {
	return b.Year == 0 && b.Month == 0 && b.Day == 0
}

// Tokens expands a BirthDate into the numeric strings people commonly weave
// into passwords: full year, two-digit year, and several day/month orderings.
// Returns nil if the BirthDate is zero-valued.
func (b BirthDate) Tokens() []Token {
	if b.IsZero() {
		return nil
	}

	yyyy := fmt.Sprintf("%04d", b.Year)
	yy := yyyy[len(yyyy)-2:]
	mm := fmt.Sprintf("%02d", b.Month)
	dd := fmt.Sprintf("%02d", b.Day)

	return []Token{
		{Label: "BirthYear", Value: yyyy},
		{Label: "BirthYearShort", Value: yy},
		{Label: "BirthDayMonth", Value: dd + mm},   // e.g. 1505
		{Label: "BirthMonthDay", Value: mm + dd},   // e.g. 0515
		{Label: "BirthFullNumeric", Value: yyyy + mm + dd},
		{Label: "BirthFullNumericShort", Value: dd + mm + yy},
	}
}

// Profile holds the structured personal information collected about an
// audit/pentest target. All fields are optional; empty fields are simply
// excluded from the generation process.
type Profile struct {
	FirstName     string
	LastName      string
	Nickname      string
	BirthDate     BirthDate
	PartnerName   string
	PetName       string
	FavoriteThing string // e.g. sports team, hobby, band
}

// Tokens flattens every non-empty field of the Profile (including any
// birth-date derivatives) into a slice of Token. This is the single entry
// point the generation engine uses to discover raw material.
func (p Profile) Tokens() []Token {
	var tokens []Token

	add := func(label, value string) {
		v := strings.TrimSpace(value)
		if v != "" {
			tokens = append(tokens, Token{Label: label, Value: v})
		}
	}

	add("FirstName", p.FirstName)
	add("LastName", p.LastName)
	add("Nickname", p.Nickname)
	add("PartnerName", p.PartnerName)
	add("PetName", p.PetName)
	add("FavoriteThing", p.FavoriteThing)

	tokens = append(tokens, p.BirthDate.Tokens()...)

	return tokens
}
