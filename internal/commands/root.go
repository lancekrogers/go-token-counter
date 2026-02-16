package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lancekrogers/go-token-counter/internal/errors"
	"github.com/lancekrogers/go-token-counter/internal/fileops"
	"github.com/lancekrogers/go-token-counter/internal/tokens"
	"github.com/lancekrogers/go-token-counter/internal/ui"
)

var (
	noColor bool
	verbose bool
)

type countOptions struct {
	model         string
	vocabFile     string
	provider      string
	all           bool
	jsonOutput    bool
	showCost      bool
	showModels    bool
	recursive     bool
	charsPerToken float64
	wordsPerToken float64
}

// Execute runs the root command.
func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	opts := &countOptions{}

	cmd := &cobra.Command{
		Use:   "tcount [file|directory]",
		Short: "Count tokens in files using various LLM tokenizers",
		Long: `Count tokens in a file or directory using multiple tokenization methods.

Provides token counts using different LLM tokenizers and approximation methods,
helping you understand token usage and estimate costs.

Supports all modern OpenAI models (GPT-5, GPT-4.1, GPT-4o, o-series) and
Anthropic Claude models (Claude 4, Claude 3 series).

When counting a directory with --recursive, the command:
  - Respects .gitignore files
  - Skips binary files automatically
  - Returns aggregated totals for all text files`,
		Example: `  tcount document.md                                       # Count tokens in a file
  tcount --model gpt-4o doc.md                             # Use GPT-4o tokenizer
  tcount --model gpt-5 doc.md                              # Use GPT-5 tokenizer
  tcount --model claude-4-sonnet doc.md                    # Use Claude 4 Sonnet
  tcount --model llama-3.1-8b --vocab-file tokenizer.model doc.md  # SentencePiece
  tcount --all --cost doc.md                               # Show all methods with costs
  tcount --json doc.md                                     # Output as JSON
  tcount -r ./src                                          # Count all files in directory
  tcount -r --models ./project                             # Show encoding→model lookup`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCount(cmd.Context(), args[0], opts)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output")
	cmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable verbose output")

	cmd.Flags().StringVar(&opts.model, "model", "", `specific model to use

OpenAI Models:
  GPT-5 series:     gpt-5, gpt-5-mini
  GPT-4.1 series:   gpt-4.1, gpt-4.1-mini, gpt-4.1-nano
  GPT-4o series:    gpt-4o, gpt-4o-mini
  o-series:         o3, o3-mini, o4-mini
  Legacy:           gpt-4, gpt-4-turbo, gpt-3.5-turbo

Anthropic Models:
  Claude 4 series:  claude-4-opus, claude-4-sonnet, claude-4.5-sonnet
  Claude 3 series:  claude-3.7-sonnet, claude-3.5-sonnet, claude-3-opus, claude-3-sonnet, claude-3-haiku

Open Source Models (tiktoken approximation):
  Llama:            llama-3.1-8b, llama-3.1-70b, llama-3.1-405b, llama-4-scout, llama-4-maverick
  DeepSeek:         deepseek-v2, deepseek-v3, deepseek-coder-v2
  Qwen:             qwen-2.5-7b, qwen-2.5-14b, qwen-2.5-72b, qwen-3-72b
  Phi:              phi-3-mini, phi-3-small, phi-3-medium`)
	cmd.Flags().StringVar(&opts.vocabFile, "vocab-file", "", `path to SentencePiece .model file for exact tokenization
Required for models that use SentencePiece (e.g., llama-3.1-8b)
Download vocab files from HuggingFace (see error messages for URLs)`)
	cmd.Flags().StringVar(&opts.provider, "provider", "all", `filter models by provider (openai, anthropic, meta, deepseek, alibaba, microsoft, all)`)
	cmd.Flags().BoolVar(&opts.all, "all", false, "show all counting methods")
	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "output in JSON format")
	cmd.Flags().BoolVar(&opts.showCost, "cost", false, "include cost estimates")
	cmd.Flags().BoolVarP(&opts.showModels, "models", "m", false, "show encoding-to-model lookup table")
	cmd.Flags().BoolVarP(&opts.recursive, "recursive", "r", false, "recursively count tokens in directory")
	cmd.Flags().BoolVarP(&opts.recursive, "directory", "d", false, "alias for --recursive")
	cmd.Flags().Float64Var(&opts.charsPerToken, "chars-per-token", 4.0, "characters per token ratio")
	cmd.Flags().Float64Var(&opts.wordsPerToken, "words-per-token", 0.75, "words per token ratio")

	return cmd
}

// validModels returns the list of valid model names.
func validModels() []string {
	return []string{
		"gpt-5", "gpt-5-mini",
		"gpt-4.1", "gpt-4.1-mini", "gpt-4.1-nano",
		"gpt-4o", "gpt-4o-mini",
		"o3", "o3-mini", "o4-mini",
		"gpt-4", "gpt-4-turbo", "gpt-3.5-turbo",
		"claude-4-opus", "claude-4-sonnet", "claude-4.5-sonnet",
		"claude-3.7-sonnet", "claude-3.5-sonnet",
		"claude-3-opus", "claude-3-sonnet", "claude-3-haiku", "claude-3",
		"llama-3.1-8b", "llama-3.1-70b", "llama-3.1-405b", "llama-4-scout", "llama-4-maverick",
		"deepseek-v2", "deepseek-v3", "deepseek-coder-v2",
		"qwen-2.5-7b", "qwen-2.5-14b", "qwen-2.5-72b", "qwen-3-72b",
		"phi-3-mini", "phi-3-small", "phi-3-medium",
	}
}

// isValidModel checks if a model name is valid.
func isValidModel(model string) bool {
	if model == "" {
		return true
	}
	for _, valid := range validModels() {
		if model == valid {
			return true
		}
	}
	return false
}

// sentencePieceVocabURLs maps model prefixes to their HuggingFace vocab download URLs.
var sentencePieceVocabURLs = map[string]string{
	"llama-3.1": "https://huggingface.co/meta-llama/Llama-3.1-8B/blob/main/original/tokenizer.model",
	"llama-4":   "https://huggingface.co/meta-llama/Llama-4-Scout-17B-16E/blob/main/tokenizer.model",
}

// isValidProvider checks if a provider name is valid.
func isValidProvider(provider string) bool {
	for _, valid := range validProviders {
		if provider == valid {
			return true
		}
	}
	return false
}

// requiresSentencePiece checks if a model can use SentencePiece tokenization
// and returns the download URL for the vocab file.
func requiresSentencePiece(model string) (bool, string) {
	for prefix, url := range sentencePieceVocabURLs {
		if strings.HasPrefix(model, prefix) {
			return true, url
		}
	}
	return false, ""
}

// validProviders lists accepted values for the --provider flag.
var validProviders = []string{"openai", "anthropic", "meta", "deepseek", "alibaba", "microsoft", "all"}

func runCount(ctx context.Context, path string, opts *countOptions) error {
	display := ui.New(noColor, verbose)

	if !isValidProvider(opts.provider) {
		return fmt.Errorf("invalid provider %q, valid options: %s", opts.provider, strings.Join(validProviders, ", "))
	}

	if !isValidModel(opts.model) {
		display.Warning("Unknown model '%s', using approximation methods", opts.model)
	}

	info, err := os.Stat(path)
	if err != nil {
		return errors.IO("accessing path", err).WithField("path", path)
	}

	var content []byte
	var fileCount int
	isDirectory := info.IsDir()

	if isDirectory {
		if !opts.recursive {
			return errors.Validation("path is a directory — use --recursive flag to count tokens in all files").WithField("path", path)
		}

		walkResult, err := fileops.WalkDirectory(ctx, path)
		if err != nil {
			return errors.IO("walking directory", err).WithField("path", path)
		}

		if len(walkResult.Files) == 0 {
			return errors.NotFound("text files in directory").WithField("path", path)
		}

		if verbose {
			display.Info("Found %d text files (skipped %d binary, %d ignored)",
				len(walkResult.Files), walkResult.SkippedBinary, walkResult.SkippedIgnore)
		}

		content, err = fileops.AggregateFileContents(ctx, walkResult.Files)
		if err != nil {
			return errors.IO("reading files", err).WithField("path", path)
		}

		fileCount = len(walkResult.Files)
	} else {
		content, err = os.ReadFile(path)
		if err != nil {
			return errors.IO("reading file", err).WithField("path", path)
		}
		fileCount = 1
	}

	// Check if model requires SentencePiece and validate vocab-file flag
	if needsSP, downloadURL := requiresSentencePiece(opts.model); needsSP && opts.vocabFile == "" {
		return fmt.Errorf(
			"model %s requires a SentencePiece vocab file\n\n"+
				"Download the tokenizer.model file from:\n"+
				"  %s\n\n"+
				"Then run:\n"+
				"  tcount --model %s --vocab-file /path/to/tokenizer.model <input>",
			opts.model, downloadURL, opts.model,
		)
	}

	counter := tokens.NewCounter(tokens.CounterOptions{
		CharsPerToken: opts.charsPerToken,
		WordsPerToken: opts.wordsPerToken,
		VocabFile:     opts.vocabFile,
		Provider:      opts.provider,
	})

	result, err := counter.Count(string(content), opts.model, opts.all)
	if err != nil {
		return errors.Wrap(err, "counting tokens")
	}

	result.FilePath = path
	result.FileSize = len(content)
	result.IsDirectory = isDirectory
	if isDirectory {
		result.FileCount = fileCount
	}

	if opts.showCost {
		result.Costs = tokens.CalculateCosts(result.Methods)
	}

	if opts.jsonOutput {
		return outputJSON(result)
	}

	return outputTable(display, result, opts.showModels)
}

func outputJSON(result *tokens.CountResult) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func outputTable(display *ui.UI, result *tokens.CountResult, showModels bool) error {
	if result.IsDirectory {
		display.Info("Token Count Report for: %s (directory)", result.FilePath)
	} else {
		display.Info("Token Count Report for: %s", result.FilePath)
	}
	display.Info("%s", strings.Repeat("=", 55))
	fmt.Println()

	display.Info("Basic Statistics:")
	if result.IsDirectory {
		display.Info("  Files:          %d", result.FileCount)
	}
	display.Info("  Characters:     %d", result.Characters)
	display.Info("  Words:          %d", result.Words)
	display.Info("  Lines:          %d", result.Lines)
	fmt.Println()

	// Calculate dynamic column widths
	methodWidth := len("Method")
	tokenWidth := len("Tokens")
	accuracyWidth := len("Accuracy")

	for _, method := range result.Methods {
		if len(method.DisplayName) > methodWidth {
			methodWidth = len(method.DisplayName)
		}
		tokenStr := fmt.Sprintf("%d", method.Tokens)
		if len(tokenStr) > tokenWidth {
			tokenWidth = len(tokenStr)
		}
	}

	// Add padding
	methodWidth += 2
	tokenWidth += 2
	if accuracyWidth < 11 {
		accuracyWidth = 11
	}

	display.Info("Token Counts by Method:")
	display.Info("  %s%s%s%s",
		corner("┌", methodWidth),
		corner("┬", tokenWidth),
		corner("┬", accuracyWidth),
		"┐")
	display.Info("  │ %-*s │ %-*s │ %-*s │",
		methodWidth-2, "Method",
		tokenWidth-2, "Tokens",
		accuracyWidth-2, "Accuracy")
	display.Info("  %s%s%s%s",
		corner("├", methodWidth),
		corner("┼", tokenWidth),
		corner("┼", accuracyWidth),
		"┤")

	for _, method := range result.Methods {
		accuracy := "Approx"
		if method.IsExact {
			accuracy = "Exact"
		} else if method.Name == "claude_3_approx" {
			accuracy = "Estimated"
		}

		display.Info("  │ %-*s │ %-*d │ %-*s │",
			methodWidth-2, method.DisplayName,
			tokenWidth-2, method.Tokens,
			accuracyWidth-2, accuracy)
	}

	display.Info("  %s%s%s%s",
		corner("└", methodWidth),
		corner("┴", tokenWidth),
		corner("┴", accuracyWidth),
		"┘")

	if len(result.Costs) > 0 {
		fmt.Println()
		display.Info("Cost Estimates (Input):")

		for _, cost := range result.Costs {
			display.Info("  %-16s $%.4f ($%.2f/1M tokens)",
				cost.Model+":", cost.Cost, cost.RatePer1M)
		}
	}

	if showModels {
		fmt.Println()
		outputModelLookup(display)
	}

	return nil
}

// corner builds a box-drawing segment: joint + repeated "─" + "─" padding.
func corner(joint string, width int) string {
	return joint + strings.Repeat("─", width)
}

// outputModelLookup prints the encoding→model mapping.
func outputModelLookup(display *ui.UI) {
	display.Info("Model Lookup:")

	byEncoding := tokens.ModelsByEncoding()

	// Deterministic order
	order := []string{"o200k_base", "cl100k_base", "claude_approx"}
	for _, enc := range order {
		models, ok := byEncoding[enc]
		if !ok {
			continue
		}
		display.Info("  %-14s %s", enc+":", strings.Join(models, ", "))
	}
}
