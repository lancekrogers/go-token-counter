package tokens

import (
	"os"
	"path/filepath"
	"testing"
)

// testModelPath returns the path to a test SentencePiece .model file.
// Tests that need a real model should call t.Skip if the file doesn't exist.
func testModelPath(t *testing.T) string {
	t.Helper()
	path := filepath.Join("testdata", "test.model")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("test .model file not available in testdata/, skipping")
	}
	return path
}

func TestNewSentencePieceTokenizer(t *testing.T) {
	tests := []struct {
		name      string
		modelPath string
		wantError bool
		errMsg    string
	}{
		{
			name:      "empty path",
			modelPath: "",
			wantError: true,
			errMsg:    "model path is required",
		},
		{
			name:      "non-existent file",
			modelPath: "nonexistent.model",
			wantError: true,
			errMsg:    "vocab file not found",
		},
		{
			name:      "non-existent nested path",
			modelPath: "/tmp/does/not/exist/model.model",
			wantError: true,
			errMsg:    "vocab file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tok, err := NewSentencePieceTokenizer(tt.modelPath)
			if tt.wantError {
				if err == nil {
					t.Error("expected error but got none")
				}
				if tok != nil {
					t.Error("expected nil tokenizer on error")
				}
				if tt.errMsg != "" && err != nil {
					if got := err.Error(); !contains(got, tt.errMsg) {
						t.Errorf("error %q should contain %q", got, tt.errMsg)
					}
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

func TestNewSentencePieceTokenizer_CorruptFile(t *testing.T) {
	// Create a temporary corrupt file
	tmpDir := t.TempDir()
	corruptPath := filepath.Join(tmpDir, "corrupt.model")
	if err := os.WriteFile(corruptPath, []byte("not a valid sentencepiece model"), 0644); err != nil {
		t.Fatalf("failed to create corrupt test file: %v", err)
	}

	tok, err := NewSentencePieceTokenizer(corruptPath)
	if err == nil {
		t.Error("expected error for corrupt model file")
	}
	if tok != nil {
		t.Error("expected nil tokenizer for corrupt model file")
	}
}

func TestSentencePieceTokenizer_Name(t *testing.T) {
	// This test doesn't need a real model - we test the method contract
	// by verifying expected output
	want := "sentencepiece"

	// Create with real model if available
	path := filepath.Join("testdata", "test.model")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("test .model file not available, skipping")
	}

	tok, err := NewSentencePieceTokenizer(path)
	if err != nil {
		t.Fatalf("failed to create tokenizer: %v", err)
	}

	if got := tok.Name(); got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

func TestSentencePieceTokenizer_DisplayName(t *testing.T) {
	path := filepath.Join("testdata", "test.model")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("test .model file not available, skipping")
	}

	tok, err := NewSentencePieceTokenizer(path)
	if err != nil {
		t.Fatalf("failed to create tokenizer: %v", err)
	}

	want := "SentencePiece"
	if got := tok.DisplayName(); got != want {
		t.Errorf("DisplayName() = %q, want %q", got, want)
	}
}

func TestSentencePieceTokenizer_IsExact(t *testing.T) {
	path := filepath.Join("testdata", "test.model")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("test .model file not available, skipping")
	}

	tok, err := NewSentencePieceTokenizer(path)
	if err != nil {
		t.Fatalf("failed to create tokenizer: %v", err)
	}

	if !tok.IsExact() {
		t.Error("SentencePieceTokenizer should return true for IsExact()")
	}
}

func TestSentencePieceTokenizer_CountTokens(t *testing.T) {
	modelPath := testModelPath(t)

	tok, err := NewSentencePieceTokenizer(modelPath)
	if err != nil {
		t.Fatalf("failed to create tokenizer: %v", err)
	}

	tests := []struct {
		name         string
		text         string
		wantPositive bool
	}{
		{
			name:         "simple text",
			text:         "Hello world",
			wantPositive: true,
		},
		{
			name:         "empty text",
			text:         "",
			wantPositive: false,
		},
		{
			name:         "longer text",
			text:         "The quick brown fox jumps over the lazy dog.",
			wantPositive: true,
		},
		{
			name:         "unicode text",
			text:         "Hello 世界",
			wantPositive: true,
		},
		{
			name:         "code snippet",
			text:         "func main() { fmt.Println(\"hello\") }",
			wantPositive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := tok.CountTokens(tt.text)
			if err != nil {
				t.Fatalf("CountTokens failed: %v", err)
			}

			if tt.wantPositive && count <= 0 {
				t.Errorf("expected positive count for %q, got: %d", tt.text, count)
			}
		})
	}
}

func TestSentencePieceTokenizer_ImplementsInterface(t *testing.T) {
	path := filepath.Join("testdata", "test.model")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("test .model file not available, skipping")
	}

	tok, err := NewSentencePieceTokenizer(path)
	if err != nil {
		t.Fatalf("failed to create tokenizer: %v", err)
	}

	// Verify it satisfies the Tokenizer interface
	var _ Tokenizer = tok
}

// contains checks if s contains substr (helper to avoid importing strings in test).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
