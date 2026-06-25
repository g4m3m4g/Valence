# Valence

A targeted wordlist generator and password-profiling tool designed for use during **authorized** security awareness audits and penetration tests.

By inputting a structured personal profile (names, nicknames, birthdates, pets, partners, or hobbies), Valence automatically generates a deduplicated list of potential password candidates. The output mirrors common human password-creation habits, including case mutations, padding, word combinations, and year suffixes.

> [!WARNING]
> **Ethical use only.** This tool is strictly intended for security professionals testing systems and personnel they are explicitly authorized to audit (e.g., contracted pentests or internal security awareness campaigns). Unauthorized use against individuals or accounts is illegal in most jurisdictions.

---

## Key Features

- **Zero Dependencies:** Built entirely using the Go standard library with no external dependencies (But might import some in the future maybe lol).
- **Modular Architecture:** The core engine (`pkg/profiler`) is completely separated from the CLI logic, allowing it to be imported into web services, TUIs, or Burp Suite extensions without overhead.
- **Stream-Friendly Design:** Outputs to `stdout` by default while pushing execution metadata to `stderr`. It pipes seamlessly into standard utilities like `sort`, `uniq`, or `hashcat`.
- **Targeted Mutations:** Dynamically extracts dates, applies capitalization variations, handles word-pairing sequences, and appends custom delimiters and suffixes.

---

## Project Layout

```text
valence/
├── go.mod
├── README.md
├── main.go               # CLI entrypoint: flags, I/O, and error handling
└── pkg/
    └── profiler/         # Core engine (Standard library only)
        ├── profile.go    # Profile, BirthDate, and Token types
        ├── mutator.go    # Case variants, suffixing, and combinators
        ├── generator.go  # Options orchestration & main generation loop
        └── generator_test.go

```

---

## Installation and Build

Ensure you have Go installed, then run:

```bash
go build -o valence .

```

---

## Usage and Examples

### Basic Execution

Pass personal profile details via CLI flags. By default, the list prints directly to `stdout`:

```bash
valence -first John -last Smith -nick Johnny \
        -partner Sarah -pet Max -favorite Lakers \
        -birthdate 1990-05-15

```

### Piping and Tool Integration

Because metadata (timing, total count) is cleanly separated to `stderr`, you can pipe outputs directly into password-cracking tools or formatters without polluting the data stream:

```bash
# Filter unique values and save to a file
valence -first John -pet Max -birthdate 1990-05-15 | sort -u > custom.txt

# Pipe directly into Hashcat for an attack
valence -first John -pet Max | hashcat -a 0 -m 1000 hashes.txt

```

### Direct File Output

```bash
valence -first John -last Smith -pet Max -o johnsmith.txt

```

---

## Configuration Flags

> [!NOTE]
> At least one profile field must be provided to run the generator.

| Flag            | Description                                                  | Default                                               |
| --------------- | ------------------------------------------------------------ | ----------------------------------------------------- |
| `-first`        | Target's first name                                          | —                                                     |
| `-last`         | Target's last name                                           | —                                                     |
| `-nick`         | Nickname or alias                                            | —                                                     |
| `-birthdate`    | Date of birth (`YYYY-MM-DD`)                                 | —                                                     |
| `-partner`      | Partner or spouse's name                                     | —                                                     |
| `-pet`          | Pet's name                                                   | —                                                     |
| `-favorite`     | Favorite sports team, hobby, or band                         | —                                                     |
| `-o`, `-output` | Custom output file path                                      | `stdout`                                              |
| `-minlen`       | Minimum candidate length                                     | `4`                                                   |
| `-maxlen`       | Maximum candidate length (`0` for unlimited)                 | `32`                                                  |
| `-suffixes`     | Comma-separated suffixes to append                           | `,1,12,123,1234,!,!!,?,007,69,01,2023,2024,2025,2026` |
| `-separators`   | Comma-separated characters to link word pairs                | `,_,.,-`                                              |
| `-no-pairs`     | Disable pairwise field combinations (faster, smaller output) | `false`                                               |
| `-max`          | Cap total password candidates returned (`0` for unlimited)   | `0`                                                   |

---

## How Generation Works

```
[Input Profile] -> [1. Tokenize] -> [2. Mutate] -> [3. Combine] -> [4. Filter & Deduplicate] -> [Final Wordlist]

```

1. **Tokenize:** Populated profile fields are transformed into a collection of labeled `Tokens` (e.g., raw names, plus birthdate derivations like `YYYY`, `YY`, `DDMM`, `MMDD`).
2. **Mutate:** Every token undergoes case adjustments (`John`, `john`, `JOHN`) and receives your specified trailing suffixes (`John123`, `john!`).
3. **Combine:** Unless `-no-pairs` is flagged, distinct token pairs are joined using every configured separator in both directions (`max_sarah`, `sarah_max`), then case-mutated and suffixed.
4. **Filter & Deduplicate:** Candidates are structured into a set to eliminate duplicates, filtered against your `-minlen`/`-maxlen` specifications, sorted deterministically, and optionally capped at your `-max` threshold.

---

## Testing

Run the full automated test suite to verify internal logic:

```bash
go test ./... -v

```

The test suite covers case mutations, suffix handling, combinator matrices, birthdate extraction rules, and core boundary protections (ensuring limits are respected and duplicates are dropped).

```

```
