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
	FirstName      string
	LastName       string
	Nickname       string
	BirthDate      BirthDate
	PartnerName    string
	PetName        string
	FavoriteThing  string // e.g. sports team, hobby, band
	PhoneNumber    string // raw; digits are extracted automatically
	City           string // hometown or current city
	Username       string // social media handle, gaming alias, etc.
	ChildName      string // name of child or sibling
	FavoriteNumber string // jersey number, lucky number, PIN seed
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
	add("City", p.City)
	add("Username", p.Username)
	add("ChildName", p.ChildName)
	add("FavoriteNumber", p.FavoriteNumber)

	// Derived: phone number digit sequences people embed in passwords.
	if digits := extractDigits(p.PhoneNumber); len(digits) >= 4 {
		add("PhoneFull", digits)
		add("PhoneLast4", digits[len(digits)-4:])
		if len(digits) >= 6 {
			add("PhoneLast6", digits[len(digits)-6:])
			add("PhoneAreaCode", digits[:3])
		}
		if len(digits) >= 10 {
			add("PhoneExchange", digits[3:6])
		}
	}

	// Derived: initial-based combinations people use to shorten their name.
	fn := strings.TrimSpace(p.FirstName)
	ln := strings.TrimSpace(p.LastName)
	if fn != "" && ln != "" {
		fnR := []rune(fn)
		lnR := []rune(ln)
		add("InitialLastName", strings.ToUpper(string(fnR[:1]))+ln)  // JSmith
		add("FirstNameInitial", fn+strings.ToUpper(string(lnR[:1]))) // JohnS
	}

	// Derived: reversed spellings of name-like fields (some people do this).
	for _, pair := range []struct{ label, value string }{
		{"FirstNameRev", p.FirstName},
		{"LastNameRev", p.LastName},
		{"NicknameRev", p.Nickname},
		{"PetNameRev", p.PetName},
		{"PartnerNameRev", p.PartnerName},
		{"CityRev", p.City},
		{"UsernameRev", p.Username},
		{"ChildNameRev", p.ChildName},
	} {
		trimmed := strings.TrimSpace(pair.value)
		if trimmed == "" {
			continue
		}
		lower := strings.ToLower(trimmed)
		rev := reverseString(lower)
		if rev != lower { // skip palindromes
			add(pair.label, rev)
		}
	}

	tokens = append(tokens, p.BirthDate.Tokens()...)

	return tokens
}

// extractDigits strips all non-digit characters from s and returns the
// remaining digit string, e.g. "(555) 123-4567" → "5551234567".
func extractDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func reverseString(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
