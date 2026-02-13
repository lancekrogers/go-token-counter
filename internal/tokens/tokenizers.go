package tokens

import (
	"fmt"
	"strings"

	"github.com/pkoukk/tiktoken-go"

	"github.com/lancekrogers/go-token-counter/internal/errors"
)

// TiktokenTokenizer implements exact tokenization for OpenAI models.
type TiktokenTokenizer struct {
	model    string
	encoding *tiktoken.Tiktoken
}

// NewTiktokenTokenizer creates a new tiktoken-based tokenizer.
func NewTiktokenTokenizer(model string) (*TiktokenTokenizer, error) {
	encodingName := getEncodingForModel(model)

	encoding, err := tiktoken.GetEncoding(encodingName)
	if err != nil {
		encoding, err = tiktoken.EncodingForModel(model)
		if err != nil {
			return nil, errors.Wrap(err, "getting encoding for model").WithField("model", model)
		}
	}

	return &TiktokenTokenizer{
		model:    model,
		encoding: encoding,
	}, nil
}

// CountTokens counts tokens using tiktoken.
func (t *TiktokenTokenizer) CountTokens(text string) (int, error) {
	tokens := t.encoding.Encode(text, nil, nil)
	return len(tokens), nil
}

// Name returns the machine-readable tokenizer identifier.
func (t *TiktokenTokenizer) Name() string {
	modelName := strings.ReplaceAll(t.model, "-", "_")
	modelName = strings.ReplaceAll(modelName, ".", "_")
	return fmt.Sprintf("tiktoken_%s", modelName)
}

// DisplayName returns the human-readable tokenizer name.
func (t *TiktokenTokenizer) DisplayName() string {
	if meta := GetModelMetadata(t.model); meta != nil {
		switch meta.Provider {
		case ProviderMeta:
			return fmt.Sprintf("Llama (%s)", t.model)
		case ProviderDeepSeek:
			return fmt.Sprintf("DeepSeek (%s)", t.model)
		case ProviderAlibaba:
			return fmt.Sprintf("Qwen (%s)", t.model)
		case ProviderMicrosoft:
			return fmt.Sprintf("Phi (%s)", t.model)
		}
	}
	return fmt.Sprintf("GPT (%s)", t.model)
}

// IsExact returns true for tiktoken tokenizers.
func (t *TiktokenTokenizer) IsExact() bool {
	return true
}

// getEncodingForModel maps model names to encoding types.
// Order matters: check o200k_base models FIRST, then fall back to cl100k_base.
func getEncodingForModel(model string) string {
	model = strings.ToLower(model)

	// o200k_base models (check these FIRST to avoid prefix collisions)
	// GPT-5 series
	if strings.HasPrefix(model, "gpt-5") {
		return "o200k_base"
	}
	// GPT-4.1 series
	if strings.HasPrefix(model, "gpt-4.1") {
		return "o200k_base"
	}
	// GPT-4o series (must check before "gpt-4")
	if strings.HasPrefix(model, "gpt-4o") {
		return "o200k_base"
	}
	// o-series models (o3, o3-mini, o4-mini)
	if strings.HasPrefix(model, "o3") || strings.HasPrefix(model, "o4") {
		return "o200k_base"
	}

	// cl100k_base models (legacy)
	if strings.HasPrefix(model, "gpt-4") || strings.HasPrefix(model, "gpt-3.5") {
		return "cl100k_base"
	}

	// Open source models (Llama, DeepSeek, Qwen, Phi)
	// These use cl100k_base as a reasonable approximation
	if strings.HasPrefix(model, "llama-") ||
		strings.HasPrefix(model, "deepseek-") ||
		strings.HasPrefix(model, "qwen-") ||
		strings.HasPrefix(model, "phi-") {
		return "cl100k_base"
	}

	// p50k_base models (older models)
	if strings.Contains(model, "davinci") || strings.Contains(model, "curie") {
		return "p50k_base"
	}

	// Default to o200k_base for unknown modern models
	return "o200k_base"
}

// ClaudeApproximator provides approximation for Claude models.
type ClaudeApproximator struct{}

// NewClaudeApproximator creates a new Claude approximator.
func NewClaudeApproximator() *ClaudeApproximator {
	return &ClaudeApproximator{}
}

// CountTokens approximates token count for Claude.
func (c *ClaudeApproximator) CountTokens(text string) (int, error) {
	chars := len(text)
	tokens := int(float64(chars) / 3.8)
	return tokens, nil
}

// Name returns the machine-readable tokenizer identifier.
func (c *ClaudeApproximator) Name() string {
	return "claude_3_approx"
}

// DisplayName returns the human-readable tokenizer name.
func (c *ClaudeApproximator) DisplayName() string {
	return "Claude (approx)"
}

// IsExact returns false for approximations.
func (c *ClaudeApproximator) IsExact() bool {
	return false
}
