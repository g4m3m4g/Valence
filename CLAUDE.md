# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build -o valence .

# Run all tests (verbose)
go test ./... -v

# Run a single test
go test ./pkg/profiler -run TestGenerate_NoDuplicates -v

# Run the tool
./valence -first John -last Smith -pet Max -birthdate 1990-05-15
```

## Architecture

Valence is split into two layers:

**`pkg/profiler/`** — the core engine (zero external dependencies, importable as a library):
- `profile.go` — `Profile` and `BirthDate` types. `Profile.Tokens()` is the single entry point that flattens all populated fields (including birthdate derivatives like `YYYY`, `YY`, `DDMM`, etc.) into `[]Token`. This is what the generator consumes.
- `mutator.go` — three pure functions: `CaseVariants`, `AppendSuffixes`, `Combine`. These are the atomic building blocks; they do no filtering or deduplication themselves.
- `generator.go` — `Generate(Profile, Options) []string` orchestrates the pipeline: tokenize → case-mutate → suffix → pair-combine → deduplicate via `map[string]struct{}` → sort → cap. `DefaultOptions()` is the canonical starting point for options; callers override individual fields.

**`main.go`** — CLI-only concerns: flag parsing, `Profile` construction, calling `Generate`, and writing output. Metadata (count, elapsed time) always goes to `stderr`; candidates go to `stdout` (or a file via `-o`). This keeps the tool pipeable into `hashcat`, `sort`, etc.

### Key design invariants

- `stdout` is reserved for wordlist candidates only; all metadata/errors use `stderr`.
- An empty suffix entry (`""`) in the suffix list is intentional — it means "also emit the bare/un-suffixed form". The comma-separated flag default starts with `,` for this reason.
- `addFiltered` in `generator.go` is the single choke point for length filtering — all candidates pass through it before entering the dedup set.
- Pairwise combination (`-no-pairs` flag / `IncludePairs` option) is the main driver of output size; disabling it drastically reduces both count and runtime.
