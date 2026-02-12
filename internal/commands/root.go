package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Obedience-Corp/go-token-counter/internal/errors"
	"github.com/Obedience-Corp/go-token-counter/internal/fileops"
	"github.com/Obedience-Corp/go-token-counter/internal/tokens"
	"github.com/Obedience-Corp/go-token-counter/internal/ui"
)

var (
	noColor bool
	verbose bool
)

type countOptions struct {
	model         string
	all           bool
	jsonOutput    bool
	showCost      bool
	recursive     bool
	charsPerToken float64
	wordsPerToken float64
}

// Execute runs the root command.
func Execute() {
	if err := newRootCmd().Execute(); err != nil {
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

When counting a directory with --recursive, the command:
  - Respects .gitignore files
  - Skips binary files automatically
  - Returns aggregated totals for all text files`,
		Example: `  tcount document.md              # Count tokens in a file
  tcount --model gpt-4 doc.md     # Use specific model tokenizer
  tcount --all --cost doc.md      # Show all methods with costs
  tcount --json doc.md            # Output as JSON
  tcount -r ./src                 # Count all files in directory
  tcount -r --json ./project      # Directory with JSON output`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCount(cmd.Context(), args[0], opts)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output")
	cmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable verbose output")

	cmd.Flags().StringVar(&opts.model, "model", "", "specific model to use (gpt-4, gpt-3.5-turbo, claude-3)")
	cmd.Flags().BoolVar(&opts.all, "all", false, "show all counting methods")
	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "output in JSON format")
	cmd.Flags().BoolVar(&opts.showCost, "cost", false, "include cost estimates")
	cmd.Flags().BoolVarP(&opts.recursive, "recursive", "r", false, "recursively count tokens in directory")
	cmd.Flags().BoolVarP(&opts.recursive, "directory", "d", false, "alias for --recursive")
	cmd.Flags().Float64Var(&opts.charsPerToken, "chars-per-token", 4.0, "characters per token ratio")
	cmd.Flags().Float64Var(&opts.wordsPerToken, "words-per-token", 0.75, "words per token ratio")

	return cmd
}

func runCount(ctx context.Context, path string, opts *countOptions) error {
	display := ui.New(noColor, verbose)

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

	counter := tokens.NewCounter(tokens.CounterOptions{
		CharsPerToken: opts.charsPerToken,
		WordsPerToken: opts.wordsPerToken,
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

	return outputTable(display, result)
}

func outputJSON(result *tokens.CountResult) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func outputTable(display *ui.UI, result *tokens.CountResult) error {
	if result.IsDirectory {
		display.Info("Token Count Report for: %s (directory)", result.FilePath)
	} else {
		display.Info("Token Count Report for: %s", result.FilePath)
	}
	display.Info("%s", strings.Repeat("═", 55))
	fmt.Println()

	display.Info("Basic Statistics:")
	if result.IsDirectory {
		display.Info("  Files:          %d", result.FileCount)
	}
	display.Info("  Characters:     %d", result.Characters)
	display.Info("  Words:          %d", result.Words)
	display.Info("  Lines:          %d", result.Lines)
	fmt.Println()

	display.Info("Token Counts by Method:")
	display.Info("  ┌─────────────────────────┬──────────┬────────────┐")
	display.Info("  │ Method                  │ Tokens   │ Accuracy   │")
	display.Info("  ├─────────────────────────┼──────────┼────────────┤")

	for _, method := range result.Methods {
		accuracy := "Approx"
		if method.IsExact {
			accuracy = "Exact"
		} else if method.Name == "claude_3_approx" {
			accuracy = "Estimated"
		}

		display.Info("  │ %-23s │ %-8d │ %-10s │",
			method.DisplayName, method.Tokens, accuracy)
	}

	display.Info("  └─────────────────────────┴──────────┴────────────┘")

	if len(result.Costs) > 0 {
		fmt.Println()
		display.Info("Cost Estimates (Input):")

		for _, cost := range result.Costs {
			display.Info("  %-16s $%.3f ($%.4f/1K tokens)",
				cost.Model+":", cost.Cost, cost.RatePer1K)
		}
	}

	return nil
}
