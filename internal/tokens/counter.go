package tokens

import (
	"fmt"
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
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Tokens      int    `json:"tokens"`
	IsExact     bool   `json:"is_exact"`
}

// CostEstimate represents cost estimation for a model.
type CostEstimate struct {
	Model     string  `json:"model"`
	Tokens    int     `json:"tokens"`
	Cost      float64 `json:"cost"`
	RatePer1K float64 `json:"rate_per_1k"`
}

// CounterOptions configures the counter.
type CounterOptions struct {
	CharsPerToken float64
	WordsPerToken float64
	VocabFile     string
}

// Counter handles token counting.
type Counter struct {
	charsPerToken float64
	wordsPerToken float64
	vocabFile     string
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

// countAllMethods counts tokens using all available methods.
func (c *Counter) countAllMethods(text string) []MethodResult {
	methods := []MethodResult{}

	for _, tokenizer := range c.tokenizers {
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

// countSpecificModel counts tokens for a specific model.
func (c *Counter) countSpecificModel(text string, model string) ([]MethodResult, error) {
	methods := []MethodResult{}

	if tokenizer, ok := c.tokenizers[model]; ok {
		count, err := tokenizer.CountTokens(text)
		if err != nil {
			return nil, err
		}
		methods = append(methods, MethodResult{
			Name:        tokenizer.Name(),
			DisplayName: tokenizer.DisplayName(),
			Tokens:      count,
			IsExact:     tokenizer.IsExact(),
		})
	} else {
		methods = append(methods, c.getApproximations(text)...)
	}

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

// initializeTokenizers sets up available tokenizers.
func (c *Counter) initializeTokenizers() {
	// OpenAI Models - GPT-5 series (o200k_base)
	if tokenizer, err := NewTiktokenTokenizer("gpt-5"); err == nil {
		c.tokenizers["gpt-5"] = tokenizer
	}
	if tokenizer, err := NewTiktokenTokenizer("gpt-5-mini"); err == nil {
		c.tokenizers["gpt-5-mini"] = tokenizer
	}

	// OpenAI Models - GPT-4.1 series (o200k_base)
	if tokenizer, err := NewTiktokenTokenizer("gpt-4.1"); err == nil {
		c.tokenizers["gpt-4.1"] = tokenizer
	}
	if tokenizer, err := NewTiktokenTokenizer("gpt-4.1-mini"); err == nil {
		c.tokenizers["gpt-4.1-mini"] = tokenizer
	}
	if tokenizer, err := NewTiktokenTokenizer("gpt-4.1-nano"); err == nil {
		c.tokenizers["gpt-4.1-nano"] = tokenizer
	}

	// OpenAI Models - GPT-4o series (o200k_base)
	if tokenizer, err := NewTiktokenTokenizer("gpt-4o"); err == nil {
		c.tokenizers["gpt-4o"] = tokenizer
	}
	if tokenizer, err := NewTiktokenTokenizer("gpt-4o-mini"); err == nil {
		c.tokenizers["gpt-4o-mini"] = tokenizer
	}

	// OpenAI Models - o-series (o200k_base)
	if tokenizer, err := NewTiktokenTokenizer("o3"); err == nil {
		c.tokenizers["o3"] = tokenizer
	}
	if tokenizer, err := NewTiktokenTokenizer("o3-mini"); err == nil {
		c.tokenizers["o3-mini"] = tokenizer
	}
	if tokenizer, err := NewTiktokenTokenizer("o4-mini"); err == nil {
		c.tokenizers["o4-mini"] = tokenizer
	}

	// OpenAI Models - Legacy (cl100k_base)
	if tokenizer, err := NewTiktokenTokenizer("gpt-4"); err == nil {
		c.tokenizers["gpt-4"] = tokenizer
	}
	if tokenizer, err := NewTiktokenTokenizer("gpt-4-turbo"); err == nil {
		c.tokenizers["gpt-4-turbo"] = tokenizer
	}
	if tokenizer, err := NewTiktokenTokenizer("gpt-3.5-turbo"); err == nil {
		c.tokenizers["gpt-3.5-turbo"] = tokenizer
	}

	// Anthropic Models - Claude (approximation)
	c.tokenizers["claude-4-opus"] = NewClaudeApproximator()
	c.tokenizers["claude-4-sonnet"] = NewClaudeApproximator()
	c.tokenizers["claude-4.5-sonnet"] = NewClaudeApproximator()
	c.tokenizers["claude-3.7-sonnet"] = NewClaudeApproximator()
	c.tokenizers["claude-3.5-sonnet"] = NewClaudeApproximator()
	c.tokenizers["claude-3-opus"] = NewClaudeApproximator()
	c.tokenizers["claude-3-sonnet"] = NewClaudeApproximator()
	c.tokenizers["claude-3-haiku"] = NewClaudeApproximator()
	// Keep legacy name for backward compatibility
	c.tokenizers["claude-3"] = NewClaudeApproximator()

	// Meta Models - Llama (tiktoken approximation)
	if tokenizer, err := NewTiktokenTokenizer("llama-3.1-8b"); err == nil {
		c.tokenizers["llama-3.1-8b"] = tokenizer
	}
	if tokenizer, err := NewTiktokenTokenizer("llama-3.1-70b"); err == nil {
		c.tokenizers["llama-3.1-70b"] = tokenizer
	}
	if tokenizer, err := NewTiktokenTokenizer("llama-3.1-405b"); err == nil {
		c.tokenizers["llama-3.1-405b"] = tokenizer
	}
	if tokenizer, err := NewTiktokenTokenizer("llama-4-scout"); err == nil {
		c.tokenizers["llama-4-scout"] = tokenizer
	}
	if tokenizer, err := NewTiktokenTokenizer("llama-4-maverick"); err == nil {
		c.tokenizers["llama-4-maverick"] = tokenizer
	}

	// DeepSeek Models
	if tokenizer, err := NewTiktokenTokenizer("deepseek-v2"); err == nil {
		c.tokenizers["deepseek-v2"] = tokenizer
	}
	if tokenizer, err := NewTiktokenTokenizer("deepseek-v3"); err == nil {
		c.tokenizers["deepseek-v3"] = tokenizer
	}
	if tokenizer, err := NewTiktokenTokenizer("deepseek-coder-v2"); err == nil {
		c.tokenizers["deepseek-coder-v2"] = tokenizer
	}

	// Alibaba Models - Qwen
	if tokenizer, err := NewTiktokenTokenizer("qwen-2.5-7b"); err == nil {
		c.tokenizers["qwen-2.5-7b"] = tokenizer
	}
	if tokenizer, err := NewTiktokenTokenizer("qwen-2.5-14b"); err == nil {
		c.tokenizers["qwen-2.5-14b"] = tokenizer
	}
	if tokenizer, err := NewTiktokenTokenizer("qwen-2.5-72b"); err == nil {
		c.tokenizers["qwen-2.5-72b"] = tokenizer
	}
	if tokenizer, err := NewTiktokenTokenizer("qwen-3-72b"); err == nil {
		c.tokenizers["qwen-3-72b"] = tokenizer
	}

	// Microsoft Models - Phi
	if tokenizer, err := NewTiktokenTokenizer("phi-3-mini"); err == nil {
		c.tokenizers["phi-3-mini"] = tokenizer
	}
	if tokenizer, err := NewTiktokenTokenizer("phi-3-small"); err == nil {
		c.tokenizers["phi-3-small"] = tokenizer
	}
	if tokenizer, err := NewTiktokenTokenizer("phi-3-medium"); err == nil {
		c.tokenizers["phi-3-medium"] = tokenizer
	}

	// SentencePiece tokenizer (when vocab file is provided)
	if c.vocabFile != "" {
		if tokenizer, err := NewSentencePieceTokenizer(c.vocabFile); err == nil {
			// Register for all models that use SentencePiece
			spModels := []string{
				"llama-3.1-8b", "llama-3.1-70b", "llama-3.1-405b",
				"llama-4-scout", "llama-4-maverick",
			}
			for _, model := range spModels {
				c.tokenizers[model] = tokenizer
			}
		}
	}
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
