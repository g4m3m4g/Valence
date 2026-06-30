<div align="center">
  <img width="1280" height="640" alt="Valence" src="https://github.com/user-attachments/assets/e2a18c7f-c006-4581-92db-a61fb7c731d0" />

  <br />

  [![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
  [![Release](https://img.shields.io/github/v/release/g4m3m4g/Valence?style=flat&color=20e18b)](https://github.com/g4m3m4g/Valence/releases)
  [![License](https://img.shields.io/github/license/g4m3m4g/Valence?style=flat&color=20e18b)](LICENSE)
  [![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macOS%20%7C%20windows-lightgrey?style=flat)](https://github.com/g4m3m4g/Valence/releases)

  <br />

  **Targeted wordlist generator for authorized penetration tests and security audits.**<br />
  Feed it a personal profile. Get a deduplicated wordlist in seconds.

  <br />

  [Installation](#installation) · [Usage](#usage) · [Flags](#flags) · [How it works](#how-it-works) · [Website](https://g4m3m4g.github.io/Valence/) · [Performance](https://g4m3m4g.github.io/Valence/#performance)

</div>

---

> [!WARNING]
> **Authorized use only.** This tool is intended exclusively for security professionals testing systems they are explicitly contracted to audit. Unauthorized use against individuals or accounts is illegal in most jurisdictions.

## Overview

Valence treats personal data — names, dates, pets, hobbies — as atomic building blocks. It combines and mutates them using the same patterns real people follow when creating passwords: leet-speak substitutions, case variations, prefix/suffix padding, pairwise combinations, and common-word mixing.

The result is a deduplicated, length-filtered wordlist that reflects how a specific person likely thinks about passwords — far more focused than a generic rockyou-style list.

## Installation

**Homebrew** (macOS and Linux — recommended)
```bash
brew tap g4m3m4g/tap
brew install valence
```

**curl** (Linux and macOS, no Go required)
```bash
curl -fsSL https://raw.githubusercontent.com/g4m3m4g/Valence/main/scripts/install.sh | sh
```

**Go install**
```bash
go install github.com/g4m3m4g/valence@latest
```

**Build from source**
```bash
git clone https://github.com/g4m3m4g/Valence.git
cd Valence
make build          # version injected from git tag automatically
```

Run `make release` to cross-compile for all platforms (linux/darwin × amd64/arm64).

## Usage

### Interactive mode

Run `valence` with no flags for a guided prompt — no flag memorization required:

```
$ valence

Valence — interactive profile builder
Leave any field blank to skip it.

  First name:                John
  Last name:                 Smith
  Nickname / alias:
  Partner's name:
  Pet's name:                Max
  Child's name:
  Favorite (team/band/hobby):
  Favorite / lucky number:
  Hometown / city:
  Username / handle:
  Phone number:
  Date of birth (YYYY-MM-DD): 1990-05-15

  Output file [john_smith.txt]:
```

### Flag mode

```bash
valence -first John -last Smith -nick Johnny \
        -partner Sarah -pet Max -child Emma \
        -phone "555-123-4567" -city Bangkok \
        -username j0hn -number 7 \
        -birthdate 1990-05-15 \
        -o john_smith.txt
```

### Piping into other tools

Candidates go to `stdout`, metadata to `stderr` — the split is intentional so the output is always clean for piping.

```bash
# Feed directly into Hashcat
valence -first John -pet Max | hashcat -a 0 -m 1000 hashes.txt

# Sort and deduplicate further
valence -first John -birthdate 1990-05-15 | sort -u > candidates.txt
```

## Flags

> At least one profile field must be provided. All fields are optional and can be combined freely.

### Profile

| Flag | Description |
|------|-------------|
| `-first` | First name |
| `-last` | Last name |
| `-nick` | Nickname or alias |
| `-birthdate` | Date of birth (`YYYY-MM-DD`) |
| `-partner` | Partner or spouse's name |
| `-pet` | Pet's name |
| `-child` | Child's or sibling's name |
| `-favorite` | Favorite sports team, hobby, or band |
| `-number` | Favorite or lucky number |
| `-city` | Hometown or current city |
| `-username` | Social media handle or gaming alias |
| `-phone` | Phone number — non-digits stripped automatically |

### Output

| Flag | Description | Default |
|------|-------------|---------|
| `-o`, `-output` | Output file path | `stdout` |
| `-max` | Cap total candidates (`0` = unlimited) | `0` |
| `-minlen` | Minimum candidate length | `4` |
| `-maxlen` | Maximum candidate length | `32` |
| `-version` | Print version and exit | — |

### Mutation toggles

| Flag | Description |
|------|-------------|
| `-no-pairs` | Disable pairwise token combinations |
| `-no-leet` | Disable leet-speak substitutions |
| `-no-prefixes` | Disable prefix prepending |
| `-no-words` | Disable common-word mixing |
| `-toggle-case` | Enable all 2ⁿ per-character case combinations (greatly increases output size) |

### Custom lists

| Flag | Description | Default |
|------|-------------|---------|
| `-suffixes` | Comma-separated suffixes to append | `,1,12,123,!,@,…` |
| `-prefixes` | Comma-separated prefixes to prepend | `!,1,12,123,0,@` |
| `-separators` | Separators used when joining token pairs | `,_,.,-` |
| `-words` | Comma-separated words to mix with profile tokens | `love,dragon,qwerty,…` |

## How it works

```
Profile → Tokenize → Mutate → Prefix/Suffix → Pair → Word-mix → Dedup & Filter → Wordlist
```

**1. Tokenize** — Each non-empty field becomes a labeled token. Birthdates expand into `YYYY`, `YY`, `DDMM`, `MMDD`, and more. Phone numbers yield last-4, last-6, area code, and full digits. Initials combos (`JSmith`, `JohnS`) and reversed spellings (`nhoj`) are derived automatically.

**2. Mutate** — Every token receives lowercase, UPPERCASE, and Title variants, plus leet-speak substitutions in 1-, 2-, and 3-rule combos (`j0hn`, `j@ne`, `$m1th`). Full 2ⁿ per-character toggle-case (`jOhN`, `JoHn`) is available via `-toggle-case` but off by default — those patterns rarely appear in real passwords and inflate output significantly.

**3. Prefix / Suffix** — Each mutated form produces four candidates: bare (`john`), suffixed (`john123`), prefixed (`!john`), and wrapped (`!john123`).

**4. Pair** — Every distinct token pair is joined with each separator in both orderings (`john_smith`, `smith_john`), then case-mutated, prefixed, and suffixed (`!johnsmith`, `johnsmith123`). A leet-on-component pass also produces `j0hnsmith` and `johnsm!th` without the combinatorial explosion of leetifying the full combined string.

**5. Word-mix** — Profile tokens are paired with breach-corpus common words (`johnlove`, `dragonsmith`). The full case/prefix/suffix pipeline re-applies. Words never pair with each other.

**6. Dedup & Filter** — A `map[string]struct{}` eliminates duplicates in O(1). Length bounds are enforced at a single choke point. Results are sorted and optionally capped.

## Project layout

```
valence/
├── .github/workflows/
│   └── release.yml          # GoReleaser on v* tag push
├── scripts/
│   └── install.sh           # curl one-liner installer
├── docs/
│   └── index.html           # GitHub Pages site
├── .goreleaser.yaml          # Cross-platform builds + Homebrew tap
├── Makefile                 # build / test / install / release targets
├── main.go                  # CLI entrypoint: flags, interactive mode, I/O
├── spinner.go               # Terminal progress indicator
└── pkg/profiler/            # Core engine — zero external dependencies
    ├── profile.go           # Profile, BirthDate, Token types
    ├── mutator.go           # CaseVariants, LeetVariants, AppendSuffixes, Combine
    ├── generator.go         # Options, DefaultOptions, Generate pipeline
    └── generator_test.go
```

## Testing & Building

```bash
make test          # run full test suite
make build         # build with version injected from git tag
make release       # cross-compile for linux/darwin × amd64/arm64
make install       # install to /usr/local/bin
```

Or directly with Go:
```bash
go test ./... -v
go build -o valence .
```

## Legal & Ethics

This tool is developed for **authorized security research, educational purposes, and defensive auditing only**. Intended users: penetration testers on contracted engagements, IAM teams auditing internal password policies, and academic researchers studying credential-generation patterns.

**Unauthorized use is prohibited.** You must never generate wordlists targeting any individual or organization without explicit, written authorization. Unauthorized use may violate:

- **United States** — Computer Fraud and Abuse Act (CFAA), 18 U.S.C. § 1030
- **United Kingdom** — Computer Misuse Act 1990
- **European Union** — Directive on Attacks Against Information Systems
- **Thailand** — Computer Crimes Act B.E. 2550

The software is provided "as is" without warranty of any kind. The authors accept no liability for misuse or damages arising from use of this tool. By using Valence, you accept sole responsibility for legal compliance in your jurisdiction.
