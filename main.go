// Command valence generates a targeted wordlist of password candidates
// from a structured personal profile, for use during authorized security
// awareness audits and penetration tests.
//
// Usage:
//
//	valence -first John -last Smith -pet Max -partner Sarah \
//	    -birthdate 1990-05-15 -o johnsmith.txt
//
// Run `valence -h` for the full flag list.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/g4m3m4g/valence/pkg/profiler"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		firstName     string
		lastName      string
		nickname      string
		birthDateStr  string
		partnerName   string
		petName       string
		favoriteThing string
		outputPath    string
		minLen        int
		maxLen        int
		suffixesRaw   string
		separatorsRaw string
		noPairs       bool
		noLeet        bool
		noPrefixes    bool
		prefixesRaw   string
		noWords       bool
		wordsRaw      string
		maxCandidates int
	)

	defaults := profiler.DefaultOptions()

	flag.StringVar(&firstName, "first", "", "Target's first name")
	flag.StringVar(&lastName, "last", "", "Target's last name")
	flag.StringVar(&nickname, "nick", "", "Target's nickname or alias")
	flag.StringVar(&birthDateStr, "birthdate", "", "Target's date of birth, format YYYY-MM-DD")
	flag.StringVar(&partnerName, "partner", "", "Name of target's partner/spouse")
	flag.StringVar(&petName, "pet", "", "Name of target's pet")
	flag.StringVar(&favoriteThing, "favorite", "", "Target's favorite team, hobby, or band")

	// Both -o and -output write to the same variable; either spelling works.
	flag.StringVar(&outputPath, "o", "", "Output file path (default: stdout)")
	flag.StringVar(&outputPath, "output", "", "Output file path (default: stdout)")

	flag.IntVar(&minLen, "minlen", defaults.MinLength, "Minimum candidate length")
	flag.IntVar(&maxLen, "maxlen", defaults.MaxLength, "Maximum candidate length (0 = unlimited)")
	flag.StringVar(&suffixesRaw, "suffixes", strings.Join(defaults.Suffixes, ","),
		"Comma-separated suffixes to append to candidates (include an empty entry to also keep the bare form)")
	flag.StringVar(&separatorsRaw, "separators", strings.Join(defaults.Separators, ","),
		"Comma-separated separators used when combining two profile fields")
	flag.BoolVar(&noPairs, "no-pairs", false, "Disable pairwise combination of distinct profile fields (smaller, faster output)")
	flag.BoolVar(&noLeet, "no-leet", false, "Disable leet-speak substitutions (a→@/4, e→3, i→1/!, o→0, s→$/5, t→7, …)")
	flag.BoolVar(&noPrefixes, "no-prefixes", false, "Disable prefix prepending (!, 1, 123, … before each candidate)")
	flag.StringVar(&prefixesRaw, "prefixes", strings.Join(defaults.Prefixes, ","),
		"Comma-separated prefixes to prepend to candidates")
	flag.BoolVar(&noWords, "no-words", false, "Disable common-word mixing (love, god, dragon, … paired with profile tokens)")
	flag.StringVar(&wordsRaw, "words", strings.Join(defaults.CommonWords, ","),
		"Comma-separated common words to mix with profile tokens")
	flag.IntVar(&maxCandidates, "max", 0, "Maximum number of candidates to output (0 = unlimited)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "valence - targeted wordlist generator for security awareness audits & pentests\n\n")
		fmt.Fprintf(os.Stderr, "For AUTHORIZED security testing and awareness training only.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n  %s [flags]\n\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	profile := profiler.Profile{
		FirstName:     firstName,
		LastName:      lastName,
		Nickname:      nickname,
		PartnerName:   partnerName,
		PetName:       petName,
		FavoriteThing: favoriteThing,
	}

	if birthDateStr != "" {
		bd, err := profiler.ParseBirthDate(birthDateStr)
		if err != nil {
			return err
		}
		profile.BirthDate = bd
	}

	// No profile data supplied — enter interactive mode.
	if len(profile.Tokens()) == 0 {
		var interactiveOutput string
		var err error
		profile, birthDateStr, interactiveOutput, err = promptProfile()
		if err != nil {
			return err
		}
		if birthDateStr != "" {
			bd, err := profiler.ParseBirthDate(birthDateStr)
			if err != nil {
				return err
			}
			profile.BirthDate = bd
		}
		if len(profile.Tokens()) == 0 {
			return fmt.Errorf("at least one profile field must be provided")
		}
		// Only use the prompted output path if -o was not explicitly passed.
		if outputPath == "" {
			outputPath = interactiveOutput
		}
	}

	opts := defaults
	opts.MinLength = minLen
	opts.MaxLength = maxLen
	opts.IncludePairs = !noPairs
	opts.IncludeLeet = !noLeet
	opts.IncludePrefixes = !noPrefixes
	opts.Prefixes = splitNonEmpty(prefixesRaw, false)
	opts.IncludeCommonWords = !noWords
	opts.CommonWords = splitNonEmpty(wordsRaw, false)
	opts.MaxCandidates = maxCandidates
	opts.Suffixes = splitNonEmpty(suffixesRaw, true) // allow "" entries through deliberately
	opts.Separators = splitNonEmpty(separatorsRaw, true)

	start := time.Now()
	candidates := profiler.Generate(profile, opts)
	elapsed := time.Since(start)

	if err := writeOutput(candidates, outputPath); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	// Execution metadata always goes to stderr so stdout stays clean for piping.
	fmt.Fprintf(os.Stderr, "Generated %d unique permutations in %s\n", len(candidates), elapsed)
	if outputPath != "" {
		fmt.Fprintf(os.Stderr, "Wordlist written to %s\n", outputPath)
	}

	return nil
}

// splitNonEmpty splits a comma-separated flag value. When keepEmpty is true,
// empty entries produced by a leading/trailing/doubled comma (e.g. ",123")
// are preserved, since an empty suffix/separator is a meaningful choice
// (it means "also keep the bare/un-joined form").
func splitNonEmpty(raw string, keepEmpty bool) []string {
	if raw == "" {
		if keepEmpty {
			return []string{""}
		}
		return nil
	}
	parts := strings.Split(raw, ",")
	if keepEmpty {
		return parts
	}
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// promptProfile interactively asks the user to fill in each profile field and
// choose an output filename. Returns the filled Profile, the raw birthdate
// string (for parsing), the chosen output path, and any read error.
// Fields left blank are silently skipped; the output filename falls back to a
// computed default when left blank.
func promptProfile() (profiler.Profile, string, string, error) {
	sc := bufio.NewScanner(os.Stdin)

	ask := func(label string) string {
		fmt.Fprintf(os.Stderr, "  %-20s ", label+":")
		if !sc.Scan() {
			return ""
		}
		return strings.TrimSpace(sc.Text())
	}

	fmt.Fprintf(os.Stderr, "\nValence — interactive profile builder\n")
	fmt.Fprintf(os.Stderr, "Leave any field blank to skip it.\n\n")

	p := profiler.Profile{
		FirstName:     ask("First name"),
		LastName:      ask("Last name"),
		Nickname:      ask("Nickname / alias"),
		PartnerName:   ask("Partner's name"),
		PetName:       ask("Pet's name"),
		FavoriteThing: ask("Favorite (team/band/hobby)"),
	}
	birthdate := ask("Date of birth (YYYY-MM-DD)")

	// Build the default output filename from whatever names we have.
	defaultOut := outputDefault(p.FirstName, p.LastName)
	fmt.Fprintf(os.Stderr, "\n  %-20s ", fmt.Sprintf("Output file [%s]:", defaultOut))
	var outPath string
	if sc.Scan() {
		outPath = strings.TrimSpace(sc.Text())
	}
	if outPath == "" {
		outPath = defaultOut
	}

	fmt.Fprintf(os.Stderr, "\n")

	if err := sc.Err(); err != nil {
		return profiler.Profile{}, "", "", fmt.Errorf("reading input: %w", err)
	}
	return p, birthdate, outPath, nil
}

// outputDefault returns a sensible default output filename derived from the
// supplied name parts. Falls back to "output.txt" when both are blank.
func outputDefault(firstName, lastName string) string {
	fn := strings.ToLower(strings.TrimSpace(firstName))
	ln := strings.ToLower(strings.TrimSpace(lastName))
	switch {
	case fn != "" && ln != "":
		return fn + "_" + ln + ".txt"
	case fn != "":
		return fn + ".txt"
	case ln != "":
		return ln + ".txt"
	default:
		return "output.txt"
	}
}

// writeOutput streams one candidate per line to either the given file path
// or, if path is empty, to stdout.
func writeOutput(words []string, path string) error {
	var w *bufio.Writer
	if path == "" {
		w = bufio.NewWriter(os.Stdout)
	} else {
		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}
		defer f.Close()
		w = bufio.NewWriter(f)
	}

	for _, word := range words {
		if _, err := w.WriteString(word); err != nil {
			return err
		}
		if err := w.WriteByte('\n'); err != nil {
			return err
		}
	}
	return w.Flush()
}
