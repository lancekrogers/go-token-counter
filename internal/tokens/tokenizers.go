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
	return fmt.Sprintf("GPT (%s)", t.model)
}

// IsExact returns true for tiktoken tokenizers.
func (t *TiktokenTokenizer) IsExact() bool {
	return true
}

// getEncodingForModel maps model names to encoding types.
func getEncodingForModel(model string) string {
	model = strings.ToLower(model)

	if strings.Contains(model, "gpt-4") || strings.Contains(model, "gpt-3.5") {
		return "cl100k_base"
	}

	if strings.Contains(model, "davinci") || strings.Contains(model, "curie") {
		return "p50k_base"
	}

	return "cl100k_base"
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
	return "Claude-3 (approx)"
}

// IsExact returns false for approximations.
func (c *ClaudeApproximator) IsExact() bool {
	return false
}
