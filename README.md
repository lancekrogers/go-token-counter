# tcount

A fast, multi-method token counter for LLM workflows. Count tokens in files and directories using exact OpenAI tokenizers, Claude approximations, and generic estimation methods — all from a single command.

## Why tcount?

When working with LLMs, you constantly need to know how many tokens your content uses. Different models tokenize differently, and existing tools only support one method at a time. `tcount` gives you every count at once:

- **Exact counts** via OpenAI's tiktoken (GPT-4, GPT-3.5)
- **Claude estimates** calibrated for Anthropic models
- **Generic approximations** (character-based, word-based, whitespace split)
- **Cost estimates** across major models
- **Directory scanning** with .gitignore support and binary file detection

## Installation

### From source

Requires Go 1.21+

```bash
go install github.com/lancekrogers/go-token-counter/cmd/tcount@latest
```

### Build from repo

```bash
git clone https://github.com/lancekrogers/go-token-counter.git
cd go-token-counter
go build -o bin/tcount ./cmd/tcount
```

Or with [just](https://github.com/casey/just):

```bash
just build
just install    # installs to $GOPATH/bin
```

## Quick Start

```bash
# Count tokens in a file
tcount myfile.txt

# Count with cost estimates
tcount --all --cost prompt.md

# Recursive directory count
tcount -r ./src

# JSON output for scripting
tcount --json document.md

# Specific model
tcount --model gpt-4 prompt.txt
```

## Usage

```
tcount [file|directory] [flags]
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--model` | | Specific model tokenizer (`gpt-4`, `gpt-3.5-turbo`, `claude-3`) |
| `--all` | | Show all counting methods |
| `--json` | | Output as JSON |
| `--cost` | | Include cost estimates |
| `--recursive` | `-r` | Recursively count files in a directory |
| `--directory` | `-d` | Alias for `--recursive` |
| `--chars-per-token` | | Character/token ratio for approximation (default: 4.0) |
| `--words-per-token` | | Words/token ratio for approximation (default: 0.75) |
| `--verbose` | | Show additional details |
| `--no-color` | | Disable color output |

## Examples

### Single file

```
$ tcount internal/tokens/counter.go

Token Count Report for: internal/tokens/counter.go
═══════════════════════════════════════════════════════

Basic Statistics:
  Characters:     5451
  Words:          662
  Lines:          222

Token Counts by Method:
  ┌─────────────────────────┬──────────┬────────────┐
  │ Method                  │ Tokens   │ Accuracy   │
  ├─────────────────────────┼──────────┼────────────┤
  │ GPT (gpt-4)             │ 1445     │ Exact      │
  │ GPT (gpt-3.5-turbo)     │ 1445     │ Exact      │
  │ Claude-3 (approx)       │ 1434     │ Estimated  │
  │ Character-based (÷4.0)  │ 1362     │ Approx     │
  │ Word-based (×1.33)      │ 882      │ Approx     │
  │ Whitespace split        │ 662      │ Approx     │
  └─────────────────────────┴──────────┴────────────┘
```

### With cost estimates

```
$ tcount --all --cost internal/tokens/counter.go

Token Count Report for: internal/tokens/counter.go
═══════════════════════════════════════════════════════

Basic Statistics:
  Characters:     5451
  Words:          662
  Lines:          222

Token Counts by Method:
  ┌─────────────────────────┬──────────┬────────────┐
  │ Method                  │ Tokens   │ Accuracy   │
  ├─────────────────────────┼──────────┼────────────┤
  │ GPT (gpt-4)             │ 1445     │ Exact      │
  │ GPT (gpt-3.5-turbo)     │ 1445     │ Exact      │
  │ Claude-3 (approx)       │ 1434     │ Estimated  │
  │ Character-based (÷4.0)  │ 1362     │ Approx     │
  │ Word-based (×1.33)      │ 882      │ Approx     │
  │ Whitespace split        │ 662      │ Approx     │
  └─────────────────────────┴──────────┴────────────┘

Cost Estimates (Input):
  gpt-4:           $0.014 ($0.0100/1K tokens)
  gpt-3.5-turbo:   $0.001 ($0.0005/1K tokens)
  claude-3-opus:   $0.022 ($0.0150/1K tokens)
  claude-3-sonnet: $0.004 ($0.0030/1K tokens)
```

### Recursive directory scan

```
$ tcount -r --verbose internal/tokens/

Found 4 text files (skipped 0 binary, 0 ignored)
Token Count Report for: internal/tokens/ (directory)
═══════════════════════════════════════════════════════

Basic Statistics:
  Files:          4
  Characters:     14929
  Words:          1906
  Lines:          612

Token Counts by Method:
  ┌─────────────────────────┬──────────┬────────────┐
  │ Method                  │ Tokens   │ Accuracy   │
  ├─────────────────────────┼──────────┼────────────┤
  │ GPT (gpt-4)             │ 4206     │ Exact      │
  │ GPT (gpt-3.5-turbo)     │ 4206     │ Exact      │
  │ Claude-3 (approx)       │ 3928     │ Estimated  │
  │ Character-based (÷4.0)  │ 3732     │ Approx     │
  │ Word-based (×1.33)      │ 2541     │ Approx     │
  │ Whitespace split        │ 1906     │ Approx     │
  └─────────────────────────┴──────────┴────────────┘
```

### JSON output

```
$ tcount --json document.md
{
  "file_path": "document.md",
  "file_size": 5451,
  "characters": 5451,
  "words": 662,
  "lines": 222,
  "methods": [
    {
      "name": "tiktoken_gpt_4",
      "display_name": "GPT (gpt-4)",
      "tokens": 1445,
      "is_exact": true
    },
    ...
  ]
}
```

JSON output is designed for piping into other tools:

```bash
# Get just the GPT-4 token count
tcount --json myfile.txt | jq '.methods[] | select(.name == "tiktoken_gpt_4") | .tokens'

# Batch count all markdown files
for f in docs/*.md; do tcount --json "$f"; done | jq -s '.'
```

## Tokenization Methods

| Method | Accuracy | Description |
|--------|----------|-------------|
| GPT (gpt-4) | Exact | Uses OpenAI's tiktoken with cl100k_base encoding |
| GPT (gpt-3.5-turbo) | Exact | Uses OpenAI's tiktoken with cl100k_base encoding |
| Claude-3 (approx) | Estimated | Calibrated character-based approximation (÷3.8) |
| Character-based | Approximate | Characters ÷ configurable ratio (default 4.0) |
| Word-based | Approximate | Words × configurable multiplier (default ×1.33) |
| Whitespace split | Approximate | Raw word count as lower bound |

## Directory Scanning

When using `-r`/`--recursive`, tcount:

- Respects `.gitignore` rules in the target directory
- Skips binary files (detected by extension and null-byte check)
- Skips `.git` directories
- Aggregates all text file contents for a combined count
- Reports file counts and skip statistics with `--verbose`

## Development

Requires [just](https://github.com/casey/just) for the build system.

```bash
just              # List all recipes
just build        # Build (with fmt + vet)
just build-only   # Build only
just fmt          # Format code
just vet          # Run go vet
just lint         # Run golangci-lint

just testing unit          # Run tests
just testing coverage      # Coverage report
just testing coverage-html # HTML coverage report
just testing bench         # Benchmarks

just cross linux       # Build for Linux amd64
just cross darwin      # Build for macOS amd64
just cross darwin-arm64 # Build for macOS arm64
just cross windows     # Build for Windows amd64
just cross all         # Build all platforms
```

## License

MIT License. See [LICENSE](LICENSE) for details.
