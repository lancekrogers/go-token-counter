package tokens

import (
	"testing"

	pkoukkTiktoken "github.com/pkoukk/tiktoken-go"
	embeddedTiktoken "github.com/tiktoken-go/tokenizer"
)

const benchmarkText = `
The quick brown fox jumps over the lazy dog. This is a standard pangram used in
typography and testing. Now let us add some more complex text to make this benchmark
representative of real-world usage patterns.

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    client := NewClient(WithRetry(3), WithTimeout(10*time.Second))
    resp, err := client.Do(ctx, &Request{
        Method: "POST",
        URL:    "https://api.example.com/v1/tokens",
        Body:   map[string]interface{}{"text": "Hello, world!"},
    })
    if err != nil {
        log.Fatalf("request failed: %v", err)
    }
    fmt.Printf("Response: %+v\n", resp)
}

The transformer architecture, introduced in the paper "Attention Is All You Need"
by Vaswani et al. (2017), has become the foundation for modern language models.
Key innovations include multi-head self-attention mechanisms, positional encodings,
and layer normalization. Models like GPT-4, Claude, and Llama build upon these
principles while introducing novel training techniques and scaling laws.

JSON example:
{
    "model": "gpt-4o",
    "messages": [
        {"role": "system", "content": "You are a helpful assistant."},
        {"role": "user", "content": "Explain quantum computing in simple terms."}
    ],
    "temperature": 0.7,
    "max_tokens": 1024,
    "stream": true
}

Unicode text: こんにちは世界 (Hello World in Japanese)
Mathematical notation: ∑(i=1 to n) of x_i² = n(n+1)(2n+1)/6
Emoji: 🚀 💻 🤖 🧠 ⚡

Error handling patterns in Go:
if err := validate(input); err != nil {
    return fmt.Errorf("validation failed for %q: %w", input.Name, err)
}

The tokenizer must handle all these cases correctly including whitespace,
special characters, code blocks, and multilingual content. Performance is
critical for applications processing millions of tokens per second.
`

// BenchmarkPkoukkColdStart benchmarks pkoukk/tiktoken-go initialization + first encode.
func BenchmarkPkoukkColdStart(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		enc, err := pkoukkTiktoken.GetEncoding("cl100k_base")
		if err != nil {
			b.Fatal(err)
		}
		_ = enc.Encode(benchmarkText, nil, nil)
	}
}

// BenchmarkEmbeddedColdStart benchmarks tiktoken-go/tokenizer initialization + first encode.
func BenchmarkEmbeddedColdStart(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		enc, err := embeddedTiktoken.Get(embeddedTiktoken.Cl100kBase)
		if err != nil {
			b.Fatal(err)
		}
		_, _, _ = enc.Encode(benchmarkText)
	}
}

// BenchmarkPkoukkThroughput benchmarks pkoukk/tiktoken-go steady-state encoding.
func BenchmarkPkoukkThroughput(b *testing.B) {
	enc, err := pkoukkTiktoken.GetEncoding("cl100k_base")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = enc.Encode(benchmarkText, nil, nil)
	}
}

// BenchmarkEmbeddedThroughput benchmarks tiktoken-go/tokenizer steady-state encoding.
func BenchmarkEmbeddedThroughput(b *testing.B) {
	enc, err := embeddedTiktoken.Get(embeddedTiktoken.Cl100kBase)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _ = enc.Encode(benchmarkText)
	}
}

// BenchmarkPkoukkO200k benchmarks pkoukk with o200k_base encoding.
func BenchmarkPkoukkO200k(b *testing.B) {
	enc, err := pkoukkTiktoken.GetEncoding("o200k_base")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = enc.Encode(benchmarkText, nil, nil)
	}
}

// BenchmarkEmbeddedO200k benchmarks embedded tokenizer with o200k_base encoding.
func BenchmarkEmbeddedO200k(b *testing.B) {
	enc, err := embeddedTiktoken.Get(embeddedTiktoken.O200kBase)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _ = enc.Encode(benchmarkText)
	}
}
