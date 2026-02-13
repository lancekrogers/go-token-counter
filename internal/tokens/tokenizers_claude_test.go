package tokens

import (
	"context"
	"testing"
	"time"
)

func TestNewClaudeAPITokenizer(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		model     string
		wantError bool
	}{
		{
			name:      "valid inputs",
			apiKey:    "sk-ant-test123",
			model:     "claude-4-sonnet",
			wantError: false,
		},
		{
			name:      "missing API key",
			apiKey:    "",
			model:     "claude-4-sonnet",
			wantError: true,
		},
		{
			name:      "missing model",
			apiKey:    "sk-ant-test123",
			model:     "",
			wantError: true,
		},
		{
			name:      "missing both",
			apiKey:    "",
			model:     "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tok, err := NewClaudeAPITokenizer(tt.apiKey, tt.model)
			if tt.wantError {
				if err == nil {
					t.Error("expected error but got none")
				}
				if tok != nil {
					t.Error("expected nil tokenizer on error")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tok == nil {
					t.Error("expected tokenizer but got nil")
				}
			}
		})
	}
}

func TestClaudeAPITokenizer_IsExact(t *testing.T) {
	tok, err := NewClaudeAPITokenizer("sk-ant-test123", "claude-4-sonnet")
	if err != nil {
		t.Fatalf("failed to create tokenizer: %v", err)
	}

	if !tok.IsExact() {
		t.Error("ClaudeAPITokenizer should return true for IsExact()")
	}
}

func TestClaudeAPITokenizer_Name(t *testing.T) {
	tok, err := NewClaudeAPITokenizer("sk-ant-test123", "claude-4-sonnet")
	if err != nil {
		t.Fatalf("failed to create tokenizer: %v", err)
	}

	want := "claude_api_claude_4_sonnet"
	if got := tok.Name(); got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

func TestClaudeAPITokenizer_DisplayName(t *testing.T) {
	tok, err := NewClaudeAPITokenizer("sk-ant-test123", "claude-4-sonnet")
	if err != nil {
		t.Fatalf("failed to create tokenizer: %v", err)
	}

	want := "Claude API (claude-4-sonnet)"
	if got := tok.DisplayName(); got != want {
		t.Errorf("DisplayName() = %q, want %q", got, want)
	}
}

func TestClaudeAPITokenizer_CountTokens_ContextCancellation(t *testing.T) {
	tok, err := NewClaudeAPITokenizer("sk-ant-test123", "claude-4-sonnet")
	if err != nil {
		t.Fatalf("failed to create tokenizer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = tok.CountTokensWithContext(ctx, "test text")
	if err == nil {
		t.Error("expected error with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}

func TestClaudeAPITokenizer_CountTokens_Timeout(t *testing.T) {
	tok, err := NewClaudeAPITokenizer("sk-ant-test123", "claude-4-sonnet")
	if err != nil {
		t.Fatalf("failed to create tokenizer: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(1 * time.Millisecond)

	_, err = tok.CountTokensWithContext(ctx, "test text")
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestClaudeAPITokenizer_ImplementsInterface(t *testing.T) {
	tok, err := NewClaudeAPITokenizer("sk-ant-test123", "claude-4-sonnet")
	if err != nil {
		t.Fatalf("failed to create tokenizer: %v", err)
	}

	// Verify it satisfies the Tokenizer interface
	var _ Tokenizer = tok
}
