# tcount

A fast, multi-method token counter for LLM workflows. Count tokens in files and directories using exact OpenAI tokenizers, Claude approximations, SentencePiece vocabularies, and generic estimation methods — all from a single command.

## Why tcount?

When working with LLMs, you constantly need to know how many tokens your content uses. Different models tokenize differently, and existing tools only support one method at a time. `tcount` gives you every count at once:

- **Exact counts** via tiktoken (GPT-5, GPT-4.1, GPT-4o, o-series, and legacy GPT-4/3.5)
- **Claude estimates** calibrated for Anthropic models (Claude 4, 3.7, 3.5, 3)
- **SentencePiece tokenization** for Llama and other open-source models
- **Context window usage** showing what percentage of a model's context you're consuming
- **Cost estimates** with 2026 per-1M token pricing
- **Provider filtering** to compare models from a specific provider
- **Directory scanning** with .gitignore support and binary file detection

## Supported Models

### OpenAI
| Model | Encoding | Context Window |
|-------|----------|----------------|
| `gpt-5`, `gpt-5-mini` | o200k_base | 200K |
| `gpt-4.1`, `gpt-4.1-mini`, `gpt-4.1-nano` | o200k_base | 128K |
| `gpt-4o`, `gpt-4o-mini` | o200k_base | 128K |
| `o3`, `o3-mini`, `o4-mini` | o200k_base | 200K |
| `gpt-4`, `gpt-4-turbo` (legacy) | cl100k_base | 8K/128K |
| `gpt-3.5-turbo` (legacy) | cl100k_base | 16K |

### Anthropic
| Model | Method | Context Window |
|-------|--------|----------------|
| `claude-4-opus`, `claude-4-sonnet` | Approximation | 200K |
| `claude-4.5-sonnet` | Approximation | 200K |
| `claude-3.7-sonnet`, `claude-3.5-sonnet` | Approximation | 200K |
| `claude-3-opus`, `claude-3-sonnet`, `claude-3-haiku` | Approximation | 200K |

### Meta (Llama)
| Model | Method | Context Window |
|-------|--------|----------------|
| `llama-4-scout`, `llama-4-maverick` | tiktoken approx / SentencePiece | 128K |
| `llama-3.1-8b`, `llama-3.1-70b`, `llama-3.1-405b` | tiktoken approx / SentencePiece | 128K |

### DeepSeek
| Model | Method | Context Window |
|-------|--------|----------------|
| `deepseek-v2`, `deepseek-v3`, `deepseek-coder-v2` | tiktoken approx | 128K |

### Alibaba (Qwen)
| Model | Method | Context Window |
|-------|--------|----------------|
| `qwen-2.5-7b`, `qwen-2.5-14b`, `qwen-2.5-72b` | tiktoken approx | 33K |
| `qwen-3-72b` | tiktoken approx | 33K |

### Microsoft (Phi)
| Model | Method | Context Window |
|-------|--------|----------------|
| `phi-3-mini`, `phi-3-small`, `phi-3-medium` | tiktoken approx | 128K |

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

# Use a specific model
tcount --model gpt-5 prompt.md

# Show all methods with costs
tcount --all --cost prompt.md

# Filter by provider
tcount --provider openai prompt.md

# Recursive directory count
tcount -r ./src

# JSON output for scripting
tcount --json document.md
```

## Usage

```
tcount [file|directory] [flags]
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--model` | | Specific model tokenizer (see supported models above) |
| `--provider` | | Filter by provider: `openai`, `anthropic`, `meta`, `deepseek`, `alibaba`, `microsoft`, `all` (default) |
| `--vocab-file` | | Path to SentencePiece `.model` file for exact Llama tokenization |
| `--all` | | Show all counting methods |
| `--json` | | Output as JSON |
| `--cost` | | Include cost estimates (per 1M tokens) |
| `--recursive` | `-r` | Recursively count files in a directory |
| `--directory` | `-d` | Alias for `--recursive` |
| `--chars-per-token` | | Character/token ratio for approximation (default: 4.0) |
| `--words-per-token` | | Words/token ratio for approximation (default: 0.75) |
| `--verbose` | | Show additional details |
| `--no-color` | | Disable color output |

## Examples

### Single file with context window display

```
$ tcount --model gpt-5 document.md

Token Count Report for: document.md
═══════════════════════════════════════════════════════

Basic Statistics:
  Characters:     5451
  Words:          662
  Lines:          222

Token Counts by Method:
  ┌─────────────────────────┬──────────┬────────────┬──────────────────┐
  │ Method                  │ Tokens   │ Accuracy   │ Context Usage    │
  ├─────────────────────────┼──────────┼────────────┼──────────────────┤
  │ GPT (gpt-5)             │ 1445     │ Exact      │ 0.7% of 200K     │
  └─────────────────────────┴──────────┴────────────┴──────────────────┘
```

### All methods with cost estimates

```
$ tcount --all --cost document.md

Token Count Report for: document.md
═══════════════════════════════════════════════════════

Basic Statistics:
  Characters:     5451
  Words:          662
  Lines:          222

Token Counts by Method:
  ┌─────────────────────────┬──────────┬────────────┬──────────────────┐
  │ Method                  │ Tokens   │ Accuracy   │ Context Usage    │
  ├─────────────────────────┼──────────┼────────────┼──────────────────┤
  │ GPT (gpt-5)             │ 1445     │ Exact      │ 0.7% of 200K     │
  │ GPT (gpt-4o)            │ 1445     │ Exact      │ 1.1% of 128K     │
  │ Claude (approx)         │ 1434     │ Estimated  │ 0.7% of 200K     │
  │ Llama (llama-3.1-8b)    │ 1445     │ Exact      │ 1.1% of 128K     │
  │ Character-based (÷4.0)  │ 1362     │ Approx     │                  │
  │ Word-based (×1.33)      │ 882      │ Approx     │                  │
  │ Whitespace split        │ 662      │ Approx     │                  │
  └─────────────────────────┴──────────┴────────────┴──────────────────┘

Cost Estimates (Input):
  gpt-5:           $0.0018 ($1.25/1M tokens)
  gpt-4o:          $0.0036 ($2.50/1M tokens)
  claude-4.5-sonnet: $0.0043 ($3.00/1M tokens)
  claude-4-sonnet: $0.0043 ($3.00/1M tokens)
```

### Provider filtering

```bash
# Show only OpenAI models
tcount --provider openai document.md

# Show only Anthropic models
tcount --provider anthropic document.md

# Show only Meta/Llama models
tcount --provider meta document.md
```

### SentencePiece vocabulary for Llama

For exact Llama tokenization, download the tokenizer model and use `--vocab-file`:

```bash
# Download from HuggingFace (requires authentication):
# Llama 3.1: https://huggingface.co/meta-llama/Llama-3.1-8B/blob/main/original/tokenizer.model
# Llama 4:   https://huggingface.co/meta-llama/Llama-4-Scout-17B-16E/blob/main/tokenizer.model

tcount --model llama-3.1-8b --vocab-file /path/to/tokenizer.model document.md
```

Without `--vocab-file`, Llama models use a tiktoken-based approximation. With it, you get exact SentencePiece tokenization.

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
  ┌─────────────────────────┬──────────┬────────────┬──────────────────┐
  │ Method                  │ Tokens   │ Accuracy   │ Context Usage    │
  ├─────────────────────────┼──────────┼────────────┼──────────────────┤
  │ GPT (gpt-5)             │ 4206     │ Exact      │ 2.1% of 200K     │
  │ Claude (approx)         │ 3928     │ Estimated  │ 2.0% of 200K     │
  │ Character-based (÷4.0)  │ 3732     │ Approx     │                  │
  │ Word-based (×1.33)      │ 2541     │ Approx     │                  │
  │ Whitespace split        │ 1906     │ Approx     │                  │
  └─────────────────────────┴──────────┴────────────┴──────────────────┘
```

### JSON output

```
$ tcount --json --model gpt-5 document.md
{
  "file_path": "document.md",
  "file_size": 5451,
  "characters": 5451,
  "words": 662,
  "lines": 222,
  "methods": [
    {
      "name": "tiktoken_gpt_5",
      "display_name": "GPT (gpt-5)",
      "tokens": 1445,
      "is_exact": true,
      "context_window": 200000
    }
  ]
}
```

JSON output is designed for piping into other tools:

```bash
# Get just the GPT-5 token count
tcount --json myfile.txt | jq '.methods[] | select(.name == "tiktoken_gpt_5") | .tokens'

# Batch count all markdown files
for f in docs/*.md; do tcount --json "$f"; done | jq -s '.'
```

## Tokenization Methods

| Method | Accuracy | Models |
|--------|----------|--------|
| tiktoken (o200k_base) | Exact | GPT-5, GPT-4.1, GPT-4o, o3, o4-mini |
| tiktoken (cl100k_base) | Exact | GPT-4, GPT-3.5 (legacy) |
| Claude approximation | Estimated | All Claude models (÷3.8 character ratio) |
| SentencePiece | Exact | Llama (with `--vocab-file`) |
| tiktoken approximation | Approximate | Llama, DeepSeek, Qwen, Phi (without vocab file) |
| Character-based | Approximate | Any (characters ÷ configurable ratio, default 4.0) |
| Word-based | Approximate | Any (words × configurable multiplier, default ×1.33) |
| Whitespace split | Approximate | Any (raw word count as lower bound) |

## Pricing

Cost estimates use 2026 per-1M token rates from official provider pricing pages.

| Model | Input/1M | Output/1M |
|-------|----------|-----------|
| GPT-5 | $1.25 | $10.00 |
| GPT-5-mini | $0.25 | $2.00 |
| GPT-4.1 | $2.00 | $8.00 |
| GPT-4o | $2.50 | $10.00 |
| Claude 4.5 Sonnet | $3.00 | $15.00 |
| Claude 4 Opus | $15.00 | $75.00 |
| Claude 4 Sonnet | $3.00 | $15.00 |

Use `--cost` to see estimates for your content. Default display shows GPT-5, GPT-4o, Claude 4.5 Sonnet, and Claude 4 Sonnet.

*Pricing last updated: 2026-02-13*

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
