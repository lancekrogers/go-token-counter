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

	// 3 encodings (o200k, cl100k, claude) + 3 approximations = 6
	if len(result.Methods) != 6 {
		t.Errorf("Expected 6 counting methods, got %d", len(result.Methods))
		for _, m := range result.Methods {
			t.Logf("  %s (%s)", m.DisplayName, m.Name)
		}
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

	// Now we have 3 encodings: o200k_base, cl100k_base, claude_approx
	expectedEncodings := []string{
		"o200k_base",
		"cl100k_base",
		"claude_approx",
	}

	for _, enc := range expectedEncodings {
		if _, ok := counter.tokenizers[enc]; !ok {
			t.Errorf("Encoding %q not registered in initializeTokenizers()", enc)
		}
	}

	if len(counter.tokenizers) != len(expectedEncodings) {
		t.Errorf("Expected %d tokenizers, got %d", len(expectedEncodings), len(counter.tokenizers))
	}
}

func TestInitializeTokenizers_Idempotent(t *testing.T) {
	counter := NewCounter(CounterOptions{})
	counter.initializeTokenizers()
	count1 := len(counter.tokenizers)
	counter.initializeTokenizers()
	count2 := len(counter.tokenizers)

	if count1 != count2 {
		t.Errorf("initializeTokenizers not idempotent: first=%d, second=%d", count1, count2)
	}
}

func TestCounter_CountSpecificModel(t *testing.T) {
	text := "The quick brown fox jumps over the lazy dog."

	counter := NewCounter(CounterOptions{})

	result, err := counter.Count(text, "gpt-5", false)
	if err != nil {
		t.Fatalf("Count with specific model failed: %v", err)
	}

	if len(result.Methods) != 1 {
		t.Errorf("Expected 1 method for specific model, got %d", len(result.Methods))
	}

	if !result.Methods[0].IsExact {
		t.Error("GPT-5 tokenizer should be exact")
	}

	if result.Methods[0].ContextWindow == 0 {
		t.Error("Expected context window to be populated for gpt-5")
	}
}

func TestCounter_CountSpecificModel_Claude(t *testing.T) {
	text := "The quick brown fox jumps over the lazy dog."

	counter := NewCounter(CounterOptions{})

	result, err := counter.Count(text, "claude-4-sonnet", false)
	if err != nil {
		t.Fatalf("Count with claude model failed: %v", err)
	}

	if len(result.Methods) != 1 {
		t.Errorf("Expected 1 method for specific model, got %d", len(result.Methods))
	}

	if result.Methods[0].IsExact {
		t.Error("Claude approximator should not be exact")
	}

	if result.Methods[0].ContextWindow == 0 {
		t.Error("Expected context window to be populated for claude-4-sonnet")
	}
}

func TestCounter_CountSpecificModel_Unknown(t *testing.T) {
	text := "The quick brown fox."

	counter := NewCounter(CounterOptions{})

	result, err := counter.Count(text, "unknown-model-xyz", false)
	if err != nil {
		t.Fatalf("Count with unknown model failed: %v", err)
	}

	// Should fall back to approximations
	if len(result.Methods) < 3 {
		t.Errorf("Expected at least 3 approximation methods for unknown model, got %d", len(result.Methods))
	}
}

func TestCounter_ProviderFilter(t *testing.T) {
	text := "The quick brown fox jumps over the lazy dog."

	counter := NewCounter(CounterOptions{Provider: "openai"})

	result, err := counter.Count(text, "", true)
	if err != nil {
		t.Fatalf("Count with provider filter failed: %v", err)
	}

	for _, method := range result.Methods {
		name := strings.ToLower(method.Name)
		// Approximation methods don't have providers, skip them
		if strings.Contains(name, "character_based") ||
			strings.Contains(name, "word_based") ||
			strings.Contains(name, "whitespace") {
			continue
		}
		// Claude should not be in the results for openai filter
		if strings.Contains(name, "claude") {
			t.Errorf("Provider filter 'openai' should exclude %s", method.Name)
		}
	}
}

func TestCounter_ProviderFilter_Anthropic(t *testing.T) {
	text := "The quick brown fox jumps over the lazy dog."

	counter := NewCounter(CounterOptions{Provider: "anthropic"})

	result, err := counter.Count(text, "", true)
	if err != nil {
		t.Fatalf("Count with provider filter failed: %v", err)
	}

	hasClaudeMethod := false
	for _, method := range result.Methods {
		if strings.Contains(method.Name, "claude") {
			hasClaudeMethod = true
		}
		// BPE encodings should not appear for anthropic filter
		if strings.Contains(method.Name, "bpe_") {
			t.Errorf("Provider filter 'anthropic' should exclude %s", method.Name)
		}
	}

	if !hasClaudeMethod {
		t.Error("Provider filter 'anthropic' should include claude_3_approx")
	}
}

func TestCounter_NoDuplicates(t *testing.T) {
	text := "The quick brown fox jumps over the lazy dog."

	counter := NewCounter(CounterOptions{})

	result, err := counter.Count(text, "", true)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}

	seen := make(map[string]bool)
	for _, method := range result.Methods {
		if seen[method.Name] {
			t.Errorf("Duplicate method name: %s", method.Name)
		}
		seen[method.Name] = true
	}

	// Exactly 6 methods: o200k, cl100k, claude, char, word, whitespace
	if len(result.Methods) != 6 {
		t.Errorf("Expected exactly 6 methods, got %d", len(result.Methods))
		for _, m := range result.Methods {
			t.Logf("  %s (%s)", m.DisplayName, m.Name)
		}
	}
}

func TestModelsByEncoding(t *testing.T) {
	byEnc := ModelsByEncoding()

	if len(byEnc) == 0 {
		t.Fatal("ModelsByEncoding() returned empty map")
	}

	// Check expected encodings exist
	for _, enc := range []string{"o200k_base", "cl100k_base", "claude_approx"} {
		models, ok := byEnc[enc]
		if !ok {
			t.Errorf("Missing encoding %q", enc)
			continue
		}
		if len(models) == 0 {
			t.Errorf("No models for encoding %q", enc)
		}
		// Verify sorted
		for i := 1; i < len(models); i++ {
			if models[i] < models[i-1] {
				t.Errorf("Models for %q not sorted: %v", enc, models)
				break
			}
		}
	}
}

func TestEncodingMatchesProvider(t *testing.T) {
	tests := []struct {
		encoding string
		provider string
		want     bool
	}{
		{"o200k_base", "openai", true},
		{"o200k_base", "anthropic", false},
		{"cl100k_base", "openai", true},
		{"cl100k_base", "meta", true},
		{"cl100k_base", "anthropic", false},
		{"claude_approx", "anthropic", true},
		{"claude_approx", "openai", false},
		{"unknown", "openai", false},
	}

	for _, tt := range tests {
		t.Run(tt.encoding+"_"+tt.provider, func(t *testing.T) {
			got := encodingMatchesProvider(tt.encoding, tt.provider)
			if got != tt.want {
				t.Errorf("encodingMatchesProvider(%q, %q) = %v, want %v",
					tt.encoding, tt.provider, got, tt.want)
			}
		})
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
