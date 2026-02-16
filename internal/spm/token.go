// Package spm implements SentencePiece BPE tokenization from .model files.
package spm

import "fmt"

// Token represents a single token from the input text.
type Token struct {
	ID   int
	Text string
}

func (t Token) String() string {
	return fmt.Sprintf("Token{ID: %v, Text: %q}", t.ID, t.Text)
}
