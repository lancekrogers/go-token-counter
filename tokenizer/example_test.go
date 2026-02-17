package tokenizer_test

import (
	"context"
	"fmt"

	"github.com/lancekrogers/go-token-counter/tokenizer"
)

func ExampleNewCounter() {
	counter := tokenizer.NewCounter(tokenizer.CounterOptions{})
	result, err := counter.Count("Hello, world!", "gpt-4o", false)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	for _, m := range result.Methods {
		if m.IsExact {
			fmt.Printf("Tokens: %d (exact)\n", m.Tokens)
		}
	}
	// Output: Tokens: 4 (exact)
}

func ExampleCounter_Count() {
	counter := tokenizer.NewCounter(tokenizer.CounterOptions{})
	result, err := counter.Count("The quick brown fox jumps over the lazy dog.", "", true)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("Characters: %d\n", result.Characters)
	fmt.Printf("Words: %d\n", result.Words)
	fmt.Printf("Methods: %d\n", len(result.Methods))
	// Output:
	// Characters: 44
	// Words: 9
	// Methods: 6
}

func ExampleCounter_CountFile() {
	counter := tokenizer.NewCounter(tokenizer.CounterOptions{})
	ctx := context.Background()
	_, err := counter.CountFile(ctx, "nonexistent.txt", "gpt-4o", false)
	if err != nil {
		fmt.Println("File not found (expected)")
	}
	// Output: File not found (expected)
}

func ExampleNewBPETokenizer() {
	tok, err := tokenizer.NewBPETokenizer("gpt-4o")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	count, err := tok.CountTokens("Hello, world!")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("Tokens: %d\n", count)
	fmt.Printf("Exact: %v\n", tok.IsExact())
	// Output:
	// Tokens: 4
	// Exact: true
}

func ExampleGetModelMetadata() {
	meta := tokenizer.GetModelMetadata("gpt-4o")
	if meta != nil {
		fmt.Printf("Model: %s\n", meta.Name)
		fmt.Printf("Provider: %s\n", meta.Provider)
		fmt.Printf("Encoding: %s\n", meta.Encoding)
		fmt.Printf("Context: %d\n", meta.ContextWindow)
	}
	// Output:
	// Model: gpt-4o
	// Provider: openai
	// Encoding: o200k_base
	// Context: 128000
}

func ExampleCalculateCosts() {
	counter := tokenizer.NewCounter(tokenizer.CounterOptions{})
	result, _ := counter.Count("Hello, world!", "gpt-4o", false)
	costs := tokenizer.CalculateCosts(result.Methods)
	for _, c := range costs {
		fmt.Printf("%s: $%.6f\n", c.Model, c.Cost)
	}
}
