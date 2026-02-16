package tokens

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/lancekrogers/go-token-counter/internal/errors"
)

// CountResult represents the result of token counting.
type CountResult struct {
	FilePath    string         `json:"file_path"`
	IsDirectory bool           `json:"is_directory,omitempty"`
	FileCount   int            `json:"file_count,omitempty"`
	FileSize    int            `json:"file_size"`
	Characters  int            `json:"characters"`
	Words       int            `json:"words"`
	Lines       int            `json:"lines"`
	Methods     []MethodResult `json:"methods"`
	Costs       []CostEstimate `json:"costs,omitempty"`
}

// MethodResult represents token count for a specific method.
type MethodResult struct {
	Name          string `json:"name"`
	DisplayName   string `json:"display_name"`
	Tokens        int    `json:"tokens"`
	IsExact       bool   `json:"is_exact"`
	ContextWindow int    `json:"context_window,omitempty"`
}

// CostEstimate represents cost estimation for a model.
type CostEstimate struct {
	Model     string  `json:"model"`
	Tokens    int     `json:"tokens"`
	Cost      float64 `json:"cost"`
	RatePer1M float64 `json:"rate_per_1m"`
}

// CounterOptions configures the counter.
type CounterOptions struct {
	CharsPerToken float64
	WordsPerToken float64
	VocabFile     string
	Provider      string
}

// Counter handles token counting.
type Counter struct {
	charsPerToken float64
	wordsPerToken float64
	vocabFile     string
	provider      string
	tokenizers    map[string]Tokenizer
}

// Tokenizer interface for different tokenization methods.
type Tokenizer interface {
	CountTokens(text string) (int, error)
	Name() string
	DisplayName() string
	IsExact() bool
}

// NewCounter creates a new token counter.
func NewCounter(opts CounterOptions) *Counter {
	if opts.CharsPerToken == 0 {
		opts.CharsPerToken = 4.0
	}
	if opts.WordsPerToken == 0 {
		opts.WordsPerToken = 0.75
	}

	return &Counter{
		charsPerToken: opts.CharsPerToken,
		wordsPerToken: opts.WordsPerToken,
		vocabFile:     opts.VocabFile,
		provider:      opts.Provider,
		tokenizers:    make(map[string]Tokenizer),
	}
}

// Count performs token counting using specified methods.
func (c *Counter) Count(text string, model string, all bool) (*CountResult, error) {
	result := &CountResult{
		Characters: len(text),
		Words:      countWords(text),
		Lines:      countLines(text),
		Methods:    []MethodResult{},
	}

	// Initialize tokenizers if needed
	c.initializeTokenizers()

	if all || model == "" {
		result.Methods = c.countAllMethods(text)
	} else {
		methods, err := c.countSpecificModel(text, model)
		if err != nil {
			return nil, errors.Wrap(err, "counting tokens for model").WithField("model", model)
		}
		result.Methods = methods
	}

	return result, nil
}

// countAllMethods counts tokens using all available encodings (deduplicated).
func (c *Counter) countAllMethods(text string) []MethodResult {
	methods := []MethodResult{}
	seen := make(map[string]bool)

	// Collect encoding keys sorted for deterministic output
	keys := make([]string, 0, len(c.tokenizers))
	for k := range c.tokenizers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, encoding := range keys {
		tokenizer := c.tokenizers[encoding]

		// Filter by provider if specified
		if c.provider != "" && c.provider != "all" {
			if !encodingMatchesProvider(encoding, c.provider) {
				continue
			}
		}

		if seen[encoding] {
			continue
		}
		seen[encoding] = true

		if count, err := tokenizer.CountTokens(text); err == nil {
			methods = append(methods, MethodResult{
				Name:        tokenizer.Name(),
				DisplayName: tokenizer.DisplayName(),
				Tokens:      count,
				IsExact:     tokenizer.IsExact(),
			})
		}
	}

	methods = append(methods, c.getApproximations(text)...)

	return methods
}

// encodingMatchesProvider checks if an encoding should be included for a provider filter.
func encodingMatchesProvider(encoding string, provider string) bool {
	switch encoding {
	case "o200k_base":
		return provider == "openai"
	case "cl100k_base":
		// cl100k_base is used by openai legacy + open source models
		return provider == "openai" || provider == "meta" || provider == "deepseek" || provider == "alibaba" || provider == "microsoft"
	case "claude_approx":
		return provider == "anthropic"
	}
	return false
}

// countSpecificModel counts tokens for a specific model.
func (c *Counter) countSpecificModel(text string, model string) ([]MethodResult, error) {
	methods := []MethodResult{}

	// Look up the model's encoding and find the right tokenizer
	meta := GetModelMetadata(model)
	if meta != nil {
		if tokenizer, ok := c.tokenizers[meta.Encoding]; ok {
			count, err := tokenizer.CountTokens(text)
			if err != nil {
				return nil, err
			}
			methods = append(methods, MethodResult{
				Name:        fmt.Sprintf("bpe_%s", strings.ReplaceAll(model, "-", "_")),
				DisplayName: fmt.Sprintf("%s (%s)", meta.Encoding, model),
				Tokens:      count,
				IsExact:     tokenizer.IsExact(),
				ContextWindow: meta.ContextWindow,
			})
			return methods, nil
		}
	}

	// Direct tokenizer lookup fallback (handles models registered directly)
	if tokenizer, ok := c.tokenizers[model]; ok {
		count, err := tokenizer.CountTokens(text)
		if err != nil {
			return nil, err
		}
		result := MethodResult{
			Name:        tokenizer.Name(),
			DisplayName: tokenizer.DisplayName(),
			Tokens:      count,
			IsExact:     tokenizer.IsExact(),
		}
		if meta != nil {
			result.ContextWindow = meta.ContextWindow
		}
		methods = append(methods, result)
		return methods, nil
	}

	// Fall back to approximations for unknown models
	methods = append(methods, c.getApproximations(text)...)
	return methods, nil
}

// getApproximations returns approximation-based token counts.
func (c *Counter) getApproximations(text string) []MethodResult {
	chars := len(text)
	words := countWords(text)

	multiplier := 1.0 / c.wordsPerToken
	multiplierStr := fmt.Sprintf("%.0f", multiplier*100)

	return []MethodResult{
		{
			Name:        fmt.Sprintf("character_based_div%.0f", c.charsPerToken),
			DisplayName: fmt.Sprintf("Character-based (÷%.1f)", c.charsPerToken),
			Tokens:      int(float64(chars) / c.charsPerToken),
			IsExact:     false,
		},
		{
			Name:        fmt.Sprintf("word_based_mul%s", multiplierStr),
			DisplayName: fmt.Sprintf("Word-based (×%.2f)", multiplier),
			Tokens:      int(float64(words) / c.wordsPerToken),
			IsExact:     false,
		},
		{
			Name:        "whitespace_split",
			DisplayName: "Whitespace split",
			Tokens:      words,
			IsExact:     false,
		},
	}
}

// initializeTokenizers sets up one tokenizer per unique encoding.
func (c *Counter) initializeTokenizers() {
	if len(c.tokenizers) > 0 {
		return
	}

	// One tokenizer per encoding, not one per model
	if tokenizer, err := NewBPETokenizerByEncoding("o200k_base"); err == nil {
		c.tokenizers["o200k_base"] = tokenizer
	}
	if tokenizer, err := NewBPETokenizerByEncoding("cl100k_base"); err == nil {
		c.tokenizers["cl100k_base"] = tokenizer
	}
	c.tokenizers["claude_approx"] = NewClaudeApproximator()

	// SentencePiece tokenizer (when vocab file is provided)
	if c.vocabFile != "" {
		if tokenizer, err := NewSPMTokenizer(c.vocabFile); err == nil {
			c.tokenizers["spm"] = tokenizer
		}
	}
}

// ModelsByEncoding returns a map of encoding name to sorted model names.
func ModelsByEncoding() map[string][]string {
	result := make(map[string][]string)
	for name, meta := range modelRegistry {
		result[meta.Encoding] = append(result[meta.Encoding], name)
	}
	for enc := range result {
		sort.Strings(result[enc])
	}
	return result
}

// countWords counts words in text.
func countWords(text string) int {
	words := 0
	inWord := false

	for _, r := range text {
		if unicode.IsSpace(r) || unicode.IsPunct(r) {
			if inWord {
				words++
				inWord = false
			}
		} else {
			inWord = true
		}
	}

	if inWord {
		words++
	}

	return words
}

// countLines counts lines in text.
func countLines(text string) int {
	if len(text) == 0 {
		return 0
	}

	lines := strings.Count(text, "\n")
	if text[len(text)-1] != '\n' {
		lines++
	}

	return lines
}
