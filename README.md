<div align="center">
  <img width="1280" height="640" alt="Valence" src="https://github.com/user-attachments/assets/e2a18c7f-c006-4581-92db-a61fb7c731d0" />
</div>

# Valence

> **Valence** (/ˈveɪ.ləns/) is inspired by the chemistry concept of the ability of atoms to bond and form new structures.
>
> Just like atoms combine in different ways, Valence treats personal data (names, dates, hobbies, etc.) as building blocks. It intelligently mixes, mutates, and combines them using rules, suffixes, and separators to generate a wide range of possible password combinations.

A targeted wordlist generator and password-profiling tool designed for use during **authorized** security awareness audits and penetration tests.

By inputting a structured personal profile (names, nicknames, birthdates, pets, partners, hobbies, and more), Valence automatically generates a deduplicated list of potential password candidates. The output mirrors common human password-creation habits — case mutations, leet-speak substitutions, prefix/suffix padding, word combinations, reversals, and common-word mixing.

> [!WARNING]
> **Ethical use only.** This tool is strictly intended for security professionals testing systems and personnel they are explicitly authorized to audit (e.g., contracted pentests or internal security awareness campaigns). Unauthorized use against individuals or accounts is illegal in most jurisdictions.

---

## Installation

### Homebrew (macOS and Linux)

```bash
brew tap g4m3m4g/tap
brew install valence
```

### curl one-liner (Linux and macOS, no Go required)

```bash
curl -fsSL https://raw.githubusercontent.com/g4m3m4g/Valence/main/scripts/install.sh | sh
```

Downloads the correct pre-built binary for your OS and architecture, installs it to `/usr/local/bin`.

### Go install

```bash
go install github.com/g4m3m4g/valence@latest
```

Requires Go 1.22+.

### Build from source

```bash
git clone https://github.com/g4m3m4g/Valence.git
cd Valence
go build -o valence .
```

---

## Key Features

- **Zero External Dependencies:** Built entirely on the Go standard library — auditable and embeddable with no overhead.
- **Modular Architecture:** The core engine (`pkg/profiler`) is completely separated from the CLI, allowing it to be imported into web services, TUIs, or Burp Suite extensions.
- **Interactive Mode:** Run `valence` with no flags to enter a guided profile builder — no flag memorization required.
- **Stream-Friendly:** Outputs to `stdout` by default while pushing metadata to `stderr`. Pipes directly into `hashcat`, `sort`, `uniq`, etc.
- **Leet-Speak Mutations:** Applies 1-, 2-, and 3-rule substitution combinations (`a→@/4`, `e→3`, `i→1/!`, `o→0`, `s→$/5`, `t→7` …).
- **Per-Character Toggle Case:** Generates all 2ⁿ upper/lower combinations for each token (`jOhN`, `JoHn`, …).
- **Prefix + Suffix Patterns:** Produces `prefix__`, `__suffix`, and `prefix__suffix` forms (`!john123`, `123smith!`).
- **Common-Word Mixing:** Pairs profile tokens with real-world breach-corpus words (`johnlove`, `dragonsmith`, `ilovejohn`).
- **Phone Number Derivation:** Automatically extracts last-4, last-6, area code, and exchange digits from a raw phone number.
- **Reversed Tokens:** Generates reversed spellings of name-like fields (`nhoj`, `htims`).
- **Initials Combinations:** Derives `JSmith` and `JohnS` style tokens automatically.
- **Progress Spinner:** Animated terminal indicator during generation and file write — degrades gracefully when piped.

---

## Project Layout

```text
valence/
├── .github/
│   └── workflows/
│       └── release.yml      # Automated release via GoReleaser on v* tag push
├── scripts/
│   └── install.sh           # curl one-liner installer
├── .goreleaser.yaml         # Cross-platform build and Homebrew tap config
├── go.mod
├── main.go                  # CLI entrypoint: flags, interactive mode, I/O
├── spinner.go               # Terminal progress spinner
├── README.md
└── pkg/
    └── profiler/            # Core engine (standard library only)
        ├── profile.go       # Profile, BirthDate, Token types and derivations
        ├── mutator.go       # CaseVariants, LeetVariants, ToggleCaseVariants,
        │                    # PrependPrefixes, AppendSuffixes, Combine
        ├── generator.go     # Options, DefaultOptions, Generate pipeline
        └── generator_test.go
```

---

## Usage

### Interactive mode

Run without flags — Valence walks you through each field and asks for an output filename:

```bash
valence
```

```
Valence — interactive profile builder
Leave any field blank to skip it.

  First name:           John
  Last name:            Smith
  Nickname / alias:
  Partner's name:
  Pet's name:           Max
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

```bash
# Pipe directly into Hashcat
valence -first John -pet Max | hashcat -a 0 -m 1000 hashes.txt

# Filter and save
valence -first John -birthdate 1990-05-15 | sort -u > candidates.txt
```

---

## Configuration Flags

> [!NOTE]
> At least one profile field must be provided. All fields are optional and can be combined freely.

### Profile fields

| Flag          | Description                                              |
| ------------- | -------------------------------------------------------- |
| `-first`      | Target's first name                                      |
| `-last`       | Target's last name                                       |
| `-nick`       | Nickname or alias                                        |
| `-birthdate`  | Date of birth (`YYYY-MM-DD`)                             |
| `-partner`    | Partner or spouse's name                                 |
| `-pet`        | Pet's name                                               |
| `-child`      | Child's or sibling's name                                |
| `-favorite`   | Favorite sports team, hobby, or band                     |
| `-number`     | Favorite or lucky number (jersey number, PIN seed, etc.) |
| `-city`       | Hometown or current city                                 |
| `-username`   | Social media handle or gaming alias                      |
| `-phone`      | Phone number — non-digits are stripped automatically     |

### Output

| Flag            | Description                          | Default  |
| --------------- | ------------------------------------ | -------- |
| `-o`, `-output` | Output file path                     | `stdout` |
| `-max`          | Cap total candidates (`0` unlimited) | `0`      |
| `-minlen`       | Minimum candidate length             | `4`      |
| `-maxlen`       | Maximum candidate length             | `32`     |

### Mutation toggles

| Flag           | Description                                                      | Default |
| -------------- | ---------------------------------------------------------------- | ------- |
| `-no-pairs`    | Disable pairwise token combinations                              | `false` |
| `-no-leet`     | Disable leet-speak substitutions                                 | `false` |
| `-no-prefixes` | Disable prefix prepending (`!john`, `123smith`)                  | `false` |
| `-no-words`    | Disable common-word mixing (`johnlove`, `dragonsmith`)           | `false` |

### Custom lists

| Flag           | Description                                           | Default               |
| -------------- | ----------------------------------------------------- | --------------------- |
| `-suffixes`    | Comma-separated suffixes to append                    | `,1,12,123,!,@,…`     |
| `-prefixes`    | Comma-separated prefixes to prepend                   | `!,1,12,123,0,@`      |
| `-separators`  | Separators used when joining token pairs              | `,_,.,-`              |
| `-words`       | Comma-separated common words to mix with profile tokens | `love,dragon,qwerty,…` |

---

## How Generation Works

```
[Profile] → [Tokenize] → [Mutate] → [Prefix/Suffix] → [Pair] → [Word-Mix] → [Deduplicate & Filter] → [Wordlist]
```

1. **Tokenize** — Each non-empty field becomes a labeled `Token`. BirthDate expands into multiple numeric forms (`YYYY`, `YY`, `DDMM`, `MMDD`, …). Phone numbers expand into last-4, last-6, area code, and full digits. Initials combinations (`JSmith`, `JohnS`) and reversed spellings (`nhoj`) are derived automatically.

2. **Mutate** — Every token receives:
   - Basic case variants: lowercase, UPPERCASE, Title
   - Per-character toggle case: all 2ⁿ combinations (`jOhN`, `JoHn`, …)
   - Leet-speak substitutions in 1-, 2-, and 3-rule combinations (`j0hn`, `j@ne`, `$m1th`)

3. **Prefix / Suffix** — Each mutated form `cv` produces four candidates:
   - `cv` (bare)
   - `cv + suffix` → `john123`, `john!`
   - `prefix + cv` → `!john`, `123john`
   - `prefix + cv + suffix` → `!john123`

4. **Pair** — Every distinct token pair is joined with each separator in both orderings (`john_smith`, `smith_john`), then case-mutated and suffixed.

5. **Word-Mix** — Each profile token is paired with every word in the common-words list (`johnlove`, `dragonsmith`), then the same case/suffix/prefix pipeline is applied. Words are never paired with each other.

6. **Deduplicate & Filter** — A `map[string]struct{}` eliminates duplicates in O(1). Length bounds (`-minlen`/`-maxlen`) are enforced at a single choke point. Results are sorted and optionally capped.

---

## Testing

```bash
go test ./... -v
```

The test suite covers case mutations, leet variants, toggle-case combinations, phone derivations, suffix/prefix handling, combinator matrices, birthdate extraction, common-word mixing, and all core boundary protections.

---

# ⚖️ Usage Regulation, Disclaimer & Responsibility Agreement

## 1. Authorized Educational & Defensive Use Only

This tool is developed, distributed, and maintained **solely for authorized security research, educational purposes, and defensive security auditing**.

It is designed strictly to assist:

- **System Administrators & IAM Teams:** To audit password strength policies and validate that user credentials do not contain easily guessable corporate or personal identifiers.
- **Authorized Penetration Testers:** To demonstrate risk to clients during explicitly scoped, legally contracted security engagements.
- **Academic Researchers:** To study human credential-generation behavioral patterns and defensive mitigation strategies.

## 2. Strict Prohibition of Unauthorized Activity

**Any utilization of this tool for unauthorized access, malicious credential stuffing, corporate espionage, harassment, stalking, or any form of cyber-attacks is strictly prohibited.**

You must **never** ingest personal details or generate wordlists targeting any individual or organization unless you have obtained **explicit, written, and legally binding authorization** from the rightful owner of those assets.

## 3. Legal Compliance & Jurisdiction

As the user, you bear sole responsibility for ensuring that your use of this software complies with all applicable local, national, and international laws. Unauthorized use of this tool may violate major cybercrime statutes, including but not limited to:

- **United States:** Computer Fraud and Abuse Act (CFAA) (18 U.S.C. § 1030)
- **United Kingdom:** Computer Misuse Act 1990
- **European Union:** Cybercrime Convention/Directives
- **Thailand:** Computer Crimes Act B.E. 2550 (and its amendments)

## 4. Limitation of Liability & Exclusion of Warranty

```text
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.

IN NO EVENT SHALL THE AUTHORS, CONTRIBUTORS, OR COPYRIGHT HOLDERS BE LIABLE
FOR ANY CLAIM, DAMAGES, OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
TORT, OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
```

By compiling, downloading, or executing this software, you explicitly agree that **the authors and contributors accept absolutely no liability** for any misuse, damage, data breaches, service disruptions, or legal consequences caused by your deployment of the tool.

## 5. Ethical OSINT Ingestion Warning

This tool relies on user-supplied Open Source Intelligence (OSINT) data. Users are strictly cautioned against violating privacy laws, data protection regulations (such as GDPR, CCPA, or PDPA), or social media platform Terms of Service when gathering profiling data. Scraping or collecting personal details without consent may constitute an independent legal violation.
