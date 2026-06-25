package profiler

import (
	"strings"
	"testing"
)

func TestCaseVariants(t *testing.T) {
	got := CaseVariants("John")
	want := map[string]bool{"John": true, "john": true, "JOHN": true}
	if len(got) != len(want) {
		t.Fatalf("CaseVariants(John) = %v, want 3 distinct variants", got)
	}
	for _, v := range got {
		if !want[v] {
			t.Errorf("unexpected variant %q", v)
		}
	}

	if got := CaseVariants(""); got != nil {
		t.Errorf("CaseVariants(\"\") = %v, want nil", got)
	}
}

func TestAppendSuffixes(t *testing.T) {
	got := AppendSuffixes("max", []string{"", "123", "!"})
	want := []string{"max", "max123", "max!"}
	if len(got) != len(want) {
		t.Fatalf("AppendSuffixes returned %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("AppendSuffixes[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCombine(t *testing.T) {
	got := Combine("max", "sarah", []string{"", "_"})
	wantContains := []string{"maxsarah", "sarahmax", "max_sarah", "sarah_max"}
	for _, w := range wantContains {
		found := false
		for _, g := range got {
			if g == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Combine result missing %q, got %v", w, got)
		}
	}

	if got := Combine("", "sarah", []string{""}); got != nil {
		t.Errorf("Combine with empty operand should return nil, got %v", got)
	}
}

func TestParseBirthDate(t *testing.T) {
	bd, err := ParseBirthDate("1990-05-15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bd.Year != 1990 || bd.Month != 5 || bd.Day != 15 {
		t.Errorf("got %+v, want Year=1990 Month=5 Day=15", bd)
	}

	if _, err := ParseBirthDate("not-a-date"); err == nil {
		t.Error("expected error for malformed birthdate, got nil")
	}
}

func TestBirthDateTokens(t *testing.T) {
	bd := BirthDate{Year: 1990, Month: 5, Day: 15}
	tokens := bd.Tokens()
	values := make(map[string]bool)
	for _, tok := range tokens {
		values[tok.Value] = true
	}
	for _, want := range []string{"1990", "90", "1505", "0515", "19900515", "150590"} {
		if !values[want] {
			t.Errorf("BirthDate.Tokens() missing expected value %q in %v", want, tokens)
		}
	}

	if got := (BirthDate{}).Tokens(); got != nil {
		t.Errorf("zero-value BirthDate.Tokens() = %v, want nil", got)
	}
}

func TestGenerate_NoDuplicates(t *testing.T) {
	p := Profile{
		FirstName:   "John",
		LastName:    "Smith",
		Nickname:    "Johnny",
		PartnerName: "Sarah",
		PetName:     "Max",
		BirthDate:   BirthDate{Year: 1990, Month: 5, Day: 15},
	}

	out := Generate(p, DefaultOptions())

	seen := make(map[string]bool, len(out))
	for _, w := range out {
		if seen[w] {
			t.Fatalf("duplicate candidate found: %q", w)
		}
		seen[w] = true
	}

	if len(out) == 0 {
		t.Fatal("Generate returned no candidates for a fully populated profile")
	}
}

func TestGenerate_RespectsLengthBounds(t *testing.T) {
	p := Profile{FirstName: "Jo", PetName: "Al"}
	opts := DefaultOptions()
	opts.MinLength = 6
	opts.MaxLength = 10

	out := Generate(p, opts)
	for _, w := range out {
		if len(w) < 6 || len(w) > 10 {
			t.Errorf("candidate %q (len=%d) violates length bounds [6,10]", w, len(w))
		}
	}
}

func TestGenerate_MaxCandidatesCap(t *testing.T) {
	p := Profile{
		FirstName:   "John",
		LastName:    "Smith",
		Nickname:    "Johnny",
		PartnerName: "Sarah",
		PetName:     "Max",
		BirthDate:   BirthDate{Year: 1990, Month: 5, Day: 15},
	}
	opts := DefaultOptions()
	opts.MaxCandidates = 10

	out := Generate(p, opts)
	if len(out) > 10 {
		t.Errorf("Generate returned %d candidates, want at most 10", len(out))
	}
}

func TestGenerate_EmptyProfile(t *testing.T) {
	out := Generate(Profile{}, DefaultOptions())
	if len(out) != 0 {
		t.Errorf("Generate(empty profile) = %v, want empty slice", out)
	}
}

func TestGenerate_ContainsExpectedPattern(t *testing.T) {
	p := Profile{FirstName: "John", BirthDate: BirthDate{Year: 1990, Month: 5, Day: 15}}
	out := Generate(p, DefaultOptions())

	found := false
	for _, w := range out {
		if strings.Contains(strings.ToLower(w), "john") && strings.Contains(w, "1990") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected a candidate combining 'john' and birth year '1990', got %v", out)
	}
}
