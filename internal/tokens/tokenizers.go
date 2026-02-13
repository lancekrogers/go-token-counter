package tokens

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	sentencepiece "github.com/eliben/go-sentencepiece"
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

// ClaudeAPITokenizer uses Anthropic's Messages.CountTokens API for exact token counting.
type ClaudeAPITokenizer struct {
	client *anthropic.Client
	model  string
}

// NewClaudeAPITokenizer creates a tokenizer that uses Anthropic's token counting API.
// Returns an error if apiKey or model is empty.
func NewClaudeAPITokenizer(apiKey, model string) (*ClaudeAPITokenizer, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required for ClaudeAPITokenizer")
	}
	if model == "" {
		return nil, fmt.Errorf("model is required for ClaudeAPITokenizer")
	}

	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &ClaudeAPITokenizer{
		client: &client,
		model:  model,
	}, nil
}

// CountTokensWithContext returns the exact token count by calling Anthropic's API.
func (t *ClaudeAPITokenizer) CountTokensWithContext(ctx context.Context, text string) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	response, err := t.client.Messages.CountTokens(ctx, anthropic.MessageCountTokensParams{
		Model: anthropic.Model(t.model),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(text)),
		},
	})
	if err != nil {
		return 0, fmt.Errorf("anthropic API request failed: %w", err)
	}

	return int(response.InputTokens), nil
}

// CountTokens implements the Tokenizer interface using a background context.
func (t *ClaudeAPITokenizer) CountTokens(text string) (int, error) {
	return t.CountTokensWithContext(context.Background(), text)
}

// Name returns the machine-readable tokenizer identifier.
func (t *ClaudeAPITokenizer) Name() string {
	modelName := strings.ReplaceAll(t.model, "-", "_")
	modelName = strings.ReplaceAll(modelName, ".", "_")
	return fmt.Sprintf("claude_api_%s", modelName)
}

// DisplayName returns the human-readable tokenizer name.
func (t *ClaudeAPITokenizer) DisplayName() string {
	return fmt.Sprintf("Claude API (%s)", t.model)
}

// IsExact returns true because this uses Anthropic's official API.
func (t *ClaudeAPITokenizer) IsExact() bool {
	return true
}

// SentencePieceTokenizer uses a .model vocab file for exact tokenization.
// Supports models like Llama 2, Mistral, and Gemma.
type SentencePieceTokenizer struct {
	processor *sentencepiece.Processor
	modelPath string
}

// NewSentencePieceTokenizer creates a tokenizer from a SentencePiece .model file.
// Returns an error if the model file doesn't exist, is inaccessible, or cannot be loaded.
func NewSentencePieceTokenizer(modelPath string) (*SentencePieceTokenizer, error) {
	if modelPath == "" {
		return nil, fmt.Errorf("model path is required for SentencePieceTokenizer")
	}

	if _, err := os.Stat(modelPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("vocab file not found: %s", modelPath)
		}
		return nil, fmt.Errorf("failed to access vocab file: %w", err)
	}

	processor, err := sentencepiece.NewProcessorFromPath(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load SentencePiece model: %w", err)
	}

	return &SentencePieceTokenizer{
		processor: processor,
		modelPath: modelPath,
	}, nil
}

// CountTokens returns the token count using the SentencePiece model.
func (t *SentencePieceTokenizer) CountTokens(text string) (int, error) {
	tokens := t.processor.Encode(text)
	return len(tokens), nil
}

// Name returns the machine-readable tokenizer identifier.
func (t *SentencePieceTokenizer) Name() string {
	return "sentencepiece"
}

// DisplayName returns the human-readable tokenizer name.
func (t *SentencePieceTokenizer) DisplayName() string {
	return "SentencePiece"
}

// IsExact returns true because SentencePiece provides exact token counts.
func (t *SentencePieceTokenizer) IsExact() bool {
	return true
}
