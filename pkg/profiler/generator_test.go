package profiler

import (
	"strings"
	"testing"
)

func TestPrependPrefixes(t *testing.T) {
	got := PrependPrefixes("john", []string{"!", "1", "123"})
	want := []string{"!john", "1john", "123john"}
	if len(got) != len(want) {
		t.Fatalf("PrependPrefixes returned %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("PrependPrefixes[%d] = %q, want %q", i, got[i], want[i])
		}
	}

	if got := PrependPrefixes("", []string{"!"}); got != nil {
		t.Errorf("PrependPrefixes with empty base should return nil, got %v", got)
	}
	if got := PrependPrefixes("john", nil); got != nil {
		t.Errorf("PrependPrefixes with nil prefixes should return nil, got %v", got)
	}
}

func TestGenerate_PrefixPatterns(t *testing.T) {
	p := Profile{FirstName: "John"}
	opts := DefaultOptions()
	out := setOf(Generate(p, opts))

	// prefix__ : "!john", "123john"
	for _, want := range []string{"!john", "123john", "1john", "@john"} {
		if !out[want] {
			t.Errorf("expected prefix-only candidate %q", want)
		}
	}

	// prefix__suffix : "!john123", "123john!"
	for _, want := range []string{"!john123", "123john!", "1john!"} {
		if !out[want] {
			t.Errorf("expected prefix+suffix candidate %q", want)
		}
	}

	// __suffix still present
	for _, want := range []string{"john123", "john!"} {
		if !out[want] {
			t.Errorf("expected suffix-only candidate %q", want)
		}
	}
}

func TestGenerate_PrefixDisabled(t *testing.T) {
	p := Profile{FirstName: "John"}
	opts := DefaultOptions()
	opts.IncludePrefixes = false
	out := setOf(Generate(p, opts))

	for _, unwanted := range []string{"!john", "123john", "1john"} {
		if out[unwanted] {
			t.Errorf("prefix candidate %q should not appear when IncludePrefixes=false", unwanted)
		}
	}
	// suffix-only should still be present
	if !out["john123"] {
		t.Error("suffix-only candidate john123 should still appear when prefixes disabled")
	}
}

func TestGenerate_CustomPrefixes(t *testing.T) {
	p := Profile{FirstName: "Jane"}
	opts := DefaultOptions()
	opts.Prefixes = []string{"xx"}
	out := setOf(Generate(p, opts))

	if !out["xxjane"] {
		t.Error("expected custom prefix candidate xxjane")
	}
	if !out["xxjane!"] {
		t.Error("expected custom prefix+suffix candidate xxjane!")
	}
	// default prefix should not appear
	if out["!jane"] {
		t.Error("default prefix !jane should not appear when custom prefixes override it")
	}
}

func TestGenerate_NoDuplicatesWithPrefixes(t *testing.T) {
	p := Profile{FirstName: "John", LastName: "Smith"}
	out := Generate(p, DefaultOptions())
	seen := make(map[string]bool, len(out))
	for _, w := range out {
		if seen[w] {
			t.Fatalf("duplicate candidate found: %q", w)
		}
		seen[w] = true
	}
}

func TestLeetVariants(t *testing.T) {
	t.Run("single substitution", func(t *testing.T) {
		if !setOf(LeetVariants("john"))["j0hn"] {
			t.Error("expected j0hn (o→0)")
		}
	})

	t.Run("two substitutions", func(t *testing.T) {
		got := setOf(LeetVariants("jane"))
		for _, want := range []string{"j@ne", "j4ne", "jan3", "j@n3", "j4n3"} {
			if !got[want] {
				t.Errorf("expected leet variant %q", want)
			}
		}
	})

	t.Run("three substitutions", func(t *testing.T) {
		got := setOf(LeetVariants("master"))
		// a→@, e→3, s→$ — three distinct char types
		if !got["m@$t3r"] {
			t.Error("expected m@$t3r (a→@, s→$, e→3)")
		}
	})

	t.Run("original excluded", func(t *testing.T) {
		for _, v := range LeetVariants("john") {
			if v == "john" || v == "John" {
				t.Errorf("LeetVariants should not include the original value, got %q", v)
			}
		}
	})

	t.Run("no substitutable chars", func(t *testing.T) {
		// "zzz" has no chars in the leet table
		if got := LeetVariants("zzz"); got != nil {
			t.Errorf("expected nil for non-substitutable input, got %v", got)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		if got := LeetVariants(""); got != nil {
			t.Errorf("LeetVariants(\"\") = %v, want nil", got)
		}
	})

	t.Run("no duplicates", func(t *testing.T) {
		got := LeetVariants("assassin")
		seen := map[string]bool{}
		for _, v := range got {
			if seen[v] {
				t.Errorf("duplicate leet variant: %q", v)
			}
			seen[v] = true
		}
	})
}

func TestToggleCaseVariants(t *testing.T) {
	t.Run("all 2^n combos present", func(t *testing.T) {
		got := setOf(ToggleCaseVariants("ab"))
		for _, want := range []string{"ab", "Ab", "aB", "AB"} {
			if !got[want] {
				t.Errorf("missing toggle variant %q", want)
			}
		}
		if len(got) != 4 {
			t.Errorf("want 4 variants for 2-letter input, got %d: %v", len(got), got)
		}
	})

	t.Run("non-letter runes unchanged", func(t *testing.T) {
		got := setOf(ToggleCaseVariants("a1b"))
		// digit position fixed; only a and b toggle → 4 variants
		for _, want := range []string{"a1b", "A1b", "a1B", "A1B"} {
			if !got[want] {
				t.Errorf("missing toggle variant %q", want)
			}
		}
	})

	t.Run("no duplicates", func(t *testing.T) {
		got := ToggleCaseVariants("John")
		seen := map[string]bool{}
		for _, v := range got {
			if seen[v] {
				t.Errorf("duplicate toggle variant: %q", v)
			}
			seen[v] = true
		}
		if len(got) != 16 { // 2^4 for 4 letters
			t.Errorf("want 16 variants for John, got %d", len(got))
		}
	})

	t.Run("empty input", func(t *testing.T) {
		if got := ToggleCaseVariants(""); got != nil {
			t.Errorf("ToggleCaseVariants(\"\") = %v, want nil", got)
		}
	})
}

func TestProfileTokens_Initials(t *testing.T) {
	p := Profile{FirstName: "John", LastName: "Smith"}
	vals := tokenValues(p.Tokens())

	if !vals["JSmith"] {
		t.Error("expected InitialLastName token JSmith")
	}
	if !vals["JohnS"] {
		t.Error("expected FirstNameInitial token JohnS")
	}
}

func TestProfileTokens_Reversed(t *testing.T) {
	p := Profile{FirstName: "John", LastName: "Smith", PetName: "Max"}
	vals := tokenValues(p.Tokens())

	if !vals["nhoj"] {
		t.Error("expected reversed FirstName token nhoj")
	}
	if !vals["htims"] {
		t.Error("expected reversed LastName token htims")
	}
	if !vals["xam"] {
		t.Error("expected reversed PetName token xam")
	}
}

func TestProfileTokens_PalindromeSkipped(t *testing.T) {
	p := Profile{FirstName: "aba"}
	vals := tokenValues(p.Tokens())
	if vals["aba"] {
		// "aba" reversed is still "aba" — should not appear as a *Rev token
		// (it may appear as the FirstName token itself, but we check the Rev label)
	}
	for _, tok := range p.Tokens() {
		if tok.Label == "FirstNameRev" {
			t.Errorf("palindrome %q should not produce a Rev token, got label %q value %q", "aba", tok.Label, tok.Value)
		}
	}
}

func TestGenerate_LeetVariantsPresent(t *testing.T) {
	p := Profile{FirstName: "John"}
	out := setOf(Generate(p, DefaultOptions()))

	for _, want := range []string{"j0hn", "J0hn", "J0HN"} {
		if !out[want] {
			t.Errorf("expected leet variant %q in output", want)
		}
	}
}

func TestGenerate_LeetDisabled(t *testing.T) {
	p := Profile{FirstName: "John"}
	opts := DefaultOptions()
	opts.IncludeLeet = false
	out := setOf(Generate(p, opts))

	if out["j0hn"] {
		t.Error("leet variant j0hn should not appear when IncludeLeet=false")
	}
}

func TestGenerate_InitialsAndReversedPresent(t *testing.T) {
	p := Profile{FirstName: "John", LastName: "Smith"}
	out := setOf(Generate(p, DefaultOptions()))

	for _, want := range []string{"JSmith", "JohnS", "nhoj", "htims"} {
		if !out[want] {
			t.Errorf("expected derived token %q in output", want)
		}
	}
}

// helpers

func setOf(ss []string) map[string]bool {
	m := make(map[string]bool, len(ss))
	for _, s := range ss {
		m[s] = true
	}
	return m
}

func tokenValues(tokens []Token) map[string]bool {
	m := make(map[string]bool, len(tokens))
	for _, tok := range tokens {
		m[tok.Value] = true
	}
	return m
}

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
