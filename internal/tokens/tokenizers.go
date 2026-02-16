package tokens

import (
	"fmt"
	"os"
	"strings"

	sentencepiece "github.com/eliben/go-sentencepiece"
	"github.com/lancekrogers/go-token-counter/internal/bpe"
	"github.com/lancekrogers/go-token-counter/internal/errors"
)

// BPETokenizerWrapper implements exact tokenization using a BPE encoding.
type BPETokenizerWrapper struct {
	encodingName string
	tokenizer    *bpe.BPETokenizer
}

// NewBPETokenizer creates a new BPE-based tokenizer for a specific model.
func NewBPETokenizer(model string) (*BPETokenizerWrapper, error) {
	encodingName := getEncodingForModel(model)
	return NewBPETokenizerByEncoding(encodingName)
}

// NewBPETokenizerByEncoding creates a tokenizer directly from an encoding name.
func NewBPETokenizerByEncoding(encodingName string) (*BPETokenizerWrapper, error) {
	tokenizer, err := bpe.NewEncoderByName(encodingName)
	if err != nil {
		return nil, errors.Wrap(err, "getting encoding").WithField("encoding", encodingName)
	}

	return &BPETokenizerWrapper{
		encodingName: encodingName,
		tokenizer:    tokenizer,
	}, nil
}

// CountTokens counts tokens using BPE tokenization.
func (t *BPETokenizerWrapper) CountTokens(text string) (int, error) {
	tokens := t.tokenizer.Encode(text, nil, nil)
	return len(tokens), nil
}

// Name returns the machine-readable tokenizer identifier.
func (t *BPETokenizerWrapper) Name() string {
	return fmt.Sprintf("bpe_%s", t.encodingName)
}

// DisplayName returns the human-readable tokenizer name.
func (t *BPETokenizerWrapper) DisplayName() string {
	return t.encodingName
}

// IsExact returns true for BPE tokenizers.
func (t *BPETokenizerWrapper) IsExact() bool {
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

// SPMTokenizerWrapper uses a .model vocab file for exact tokenization.
// Supports models like Llama 2, Mistral, and Gemma.
type SPMTokenizerWrapper struct {
	processor *sentencepiece.Processor
	modelPath string
}

// NewSPMTokenizer creates a tokenizer from a SentencePiece .model file.
// Returns an error if the model file doesn't exist, is inaccessible, or cannot be loaded.
func NewSPMTokenizer(modelPath string) (*SPMTokenizerWrapper, error) {
	if modelPath == "" {
		return nil, fmt.Errorf("model path is required for SPMTokenizerWrapper")
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

	return &SPMTokenizerWrapper{
		processor: processor,
		modelPath: modelPath,
	}, nil
}

// CountTokens returns the token count using the SentencePiece model.
func (t *SPMTokenizerWrapper) CountTokens(text string) (int, error) {
	tokens := t.processor.Encode(text)
	return len(tokens), nil
}

// Name returns the machine-readable tokenizer identifier.
func (t *SPMTokenizerWrapper) Name() string {
	return "spm"
}

// DisplayName returns the human-readable tokenizer name.
func (t *SPMTokenizerWrapper) DisplayName() string {
	return "SentencePiece"
}

// IsExact returns true because SentencePiece provides exact token counts.
func (t *SPMTokenizerWrapper) IsExact() bool {
	return true
}
