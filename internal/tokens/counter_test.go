package tokens

import (
	"strings"
	"testing"
)

func TestCountWords(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{"Empty", "", 0},
		{"Single word", "hello", 1},
		{"Multiple words", "hello world", 2},
		{"With punctuation", "Hello, world!", 2},
		{"Multiple spaces", "hello   world", 2},
		{"With newlines", "hello\nworld", 2},
		{"Mixed whitespace", "hello\t\nworld  test", 3},
		{"Code sample", "func main() { fmt.Println(\"hello\") }", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countWords(tt.text)
			if result != tt.expected {
				t.Errorf("countWords(%q) = %d, want %d", tt.text, result, tt.expected)
			}
		})
	}
}

func TestCountLines(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{"Empty", "", 0},
		{"Single line", "hello", 1},
		{"Two lines", "hello\nworld", 2},
		{"Trailing newline", "hello\n", 1},
		{"Multiple newlines", "hello\n\nworld", 3},
		{"Windows line endings", "hello\r\nworld", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countLines(tt.text)
			if result != tt.expected {
				t.Errorf("countLines(%q) = %d, want %d", tt.text, result, tt.expected)
			}
		})
	}
}

func TestCounter_Count(t *testing.T) {
	text := "The quick brown fox jumps over the lazy dog."

	counter := NewCounter(CounterOptions{
		CharsPerToken: 4.0,
		WordsPerToken: 0.75,
	})

	result, err := counter.Count(text, "", true)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}

	if result.Characters != len(text) {
		t.Errorf("Characters = %d, want %d", result.Characters, len(text))
	}

	if result.Words != 9 {
		t.Errorf("Words = %d, want 9", result.Words)
	}

	if result.Lines != 1 {
		t.Errorf("Lines = %d, want 1", result.Lines)
	}

	if len(result.Methods) < 3 {
		t.Errorf("Expected at least 3 counting methods, got %d", len(result.Methods))
	}

	hasCharBased := false
	hasWordBased := false
	for _, method := range result.Methods {
		if strings.Contains(method.DisplayName, "Character-based") {
			hasCharBased = true
		}
		if strings.Contains(method.DisplayName, "Word-based") {
			hasWordBased = true
		}
	}

	if !hasCharBased {
		t.Error("Missing character-based approximation")
	}
	if !hasWordBased {
		t.Error("Missing word-based approximation")
	}
}

func TestGetApproximations(t *testing.T) {
	counter := NewCounter(CounterOptions{
		CharsPerToken: 4.0,
		WordsPerToken: 0.75,
	})

	text := "This is a test text with twenty characters and six words."
	methods := counter.getApproximations(text)

	if len(methods) != 3 {
		t.Errorf("Expected 3 approximation methods, got %d", len(methods))
	}

	expectedCharTokens := len(text) / 4
	found := false
	for _, method := range methods {
		if strings.Contains(method.DisplayName, "Character-based") {
			found = true
			if method.Tokens != expectedCharTokens {
				t.Errorf("Character-based tokens = %d, want %d", method.Tokens, expectedCharTokens)
			}
		}
	}
	if !found {
		t.Error("Character-based approximation not found")
	}
}

func TestCounterOptions(t *testing.T) {
	text := "Test text"

	counter := NewCounter(CounterOptions{
		CharsPerToken: 2.0,
		WordsPerToken: 0.5,
	})

	methods := counter.getApproximations(text)

	for _, method := range methods {
		if strings.Contains(method.DisplayName, "÷2.0") {
			expectedTokens := len(text) / 2
			if method.Tokens != expectedTokens {
				t.Errorf("Custom char ratio: got %d tokens, want %d", method.Tokens, expectedTokens)
			}
			return
		}
	}

	t.Error("Custom character ratio not applied")
}

func TestInitializeTokenizers(t *testing.T) {
	counter := NewCounter(CounterOptions{})
	counter.initializeTokenizers()

	expectedModels := []string{
		// OpenAI - GPT-5 series
		"gpt-5",
		"gpt-5-mini",
		// OpenAI - GPT-4.1 series
		"gpt-4.1",
		"gpt-4.1-mini",
		"gpt-4.1-nano",
		// OpenAI - GPT-4o series
		"gpt-4o",
		"gpt-4o-mini",
		// OpenAI - o-series
		"o3",
		"o3-mini",
		"o4-mini",
		// OpenAI - Legacy
		"gpt-4",
		"gpt-4-turbo",
		"gpt-3.5-turbo",
		// Anthropic - Claude
		"claude-4-opus",
		"claude-4-sonnet",
		"claude-4.5-sonnet",
		"claude-3.7-sonnet",
		"claude-3.5-sonnet",
		"claude-3-opus",
		"claude-3-sonnet",
		"claude-3-haiku",
		"claude-3", // legacy
		// Meta - Llama
		"llama-3.1-8b",
		"llama-3.1-70b",
		"llama-3.1-405b",
		"llama-4-scout",
		"llama-4-maverick",
		// DeepSeek
		"deepseek-v2",
		"deepseek-v3",
		"deepseek-coder-v2",
		// Alibaba - Qwen
		"qwen-2.5-7b",
		"qwen-2.5-14b",
		"qwen-2.5-72b",
		"qwen-3-72b",
		// Microsoft - Phi
		"phi-3-mini",
		"phi-3-small",
		"phi-3-medium",
	}

	for _, model := range expectedModels {
		if _, ok := counter.tokenizers[model]; !ok {
			t.Errorf("Model %q not registered in initializeTokenizers()", model)
		}
	}

	if len(counter.tokenizers) != len(expectedModels) {
		t.Errorf("Expected %d tokenizers, got %d", len(expectedModels), len(counter.tokenizers))
	}
}

func BenchmarkCountWords(b *testing.B) {
	text := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = countWords(text)
	}
}

func BenchmarkCounter_Count(b *testing.B) {
	text := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 100)
	counter := NewCounter(CounterOptions{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = counter.Count(text, "", false)
	}
}
