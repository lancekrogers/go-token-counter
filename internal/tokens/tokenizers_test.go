package tokens

import (
	"testing"
)

func TestGetEncodingForModel(t *testing.T) {
	tests := []struct {
		name         string
		model        string
		wantEncoding string
		reason       string
	}{
		// o200k_base models
		{
			name:         "gpt-5",
			model:        "gpt-5",
			wantEncoding: "o200k_base",
			reason:       "GPT-5 uses o200k_base",
		},
		{
			name:         "gpt-5-mini",
			model:        "gpt-5-mini",
			wantEncoding: "o200k_base",
			reason:       "GPT-5 mini uses o200k_base",
		},
		{
			name:         "gpt-4.1",
			model:        "gpt-4.1",
			wantEncoding: "o200k_base",
			reason:       "GPT-4.1 uses o200k_base",
		},
		{
			name:         "gpt-4.1-mini",
			model:        "gpt-4.1-mini",
			wantEncoding: "o200k_base",
			reason:       "GPT-4.1-mini uses o200k_base",
		},
		{
			name:         "gpt-4.1-nano",
			model:        "gpt-4.1-nano",
			wantEncoding: "o200k_base",
			reason:       "GPT-4.1-nano uses o200k_base",
		},
		{
			name:         "gpt-4o",
			model:        "gpt-4o",
			wantEncoding: "o200k_base",
			reason:       "GPT-4o uses o200k_base (this was the bug!)",
		},
		{
			name:         "gpt-4o-mini",
			model:        "gpt-4o-mini",
			wantEncoding: "o200k_base",
			reason:       "GPT-4o-mini uses o200k_base",
		},
		{
			name:         "o3",
			model:        "o3",
			wantEncoding: "o200k_base",
			reason:       "o3 uses o200k_base",
		},
		{
			name:         "o3-mini",
			model:        "o3-mini",
			wantEncoding: "o200k_base",
			reason:       "o3-mini uses o200k_base",
		},
		{
			name:         "o4-mini",
			model:        "o4-mini",
			wantEncoding: "o200k_base",
			reason:       "o4-mini uses o200k_base",
		},

		// cl100k_base models (legacy)
		{
			name:         "gpt-4",
			model:        "gpt-4",
			wantEncoding: "cl100k_base",
			reason:       "Legacy GPT-4 uses cl100k_base",
		},
		{
			name:         "gpt-4-turbo",
			model:        "gpt-4-turbo",
			wantEncoding: "cl100k_base",
			reason:       "GPT-4-turbo uses cl100k_base",
		},
		{
			name:         "gpt-3.5-turbo",
			model:        "gpt-3.5-turbo",
			wantEncoding: "cl100k_base",
			reason:       "GPT-3.5-turbo uses cl100k_base",
		},

		// p50k_base models
		{
			name:         "text-davinci-003",
			model:        "text-davinci-003",
			wantEncoding: "p50k_base",
			reason:       "Davinci uses p50k_base",
		},

		// Unknown model (should default to o200k_base)
		{
			name:         "unknown-future-model",
			model:        "unknown-future-model",
			wantEncoding: "o200k_base",
			reason:       "Unknown models default to o200k_base",
		},

		// Open source models (cl100k_base approximation)
		{
			name:         "llama-3.1-8b",
			model:        "llama-3.1-8b",
			wantEncoding: "cl100k_base",
			reason:       "Llama uses cl100k_base approximation",
		},
		{
			name:         "deepseek-v3",
			model:        "deepseek-v3",
			wantEncoding: "cl100k_base",
			reason:       "DeepSeek uses cl100k_base approximation",
		},
		{
			name:         "qwen-2.5-7b",
			model:        "qwen-2.5-7b",
			wantEncoding: "cl100k_base",
			reason:       "Qwen uses cl100k_base approximation",
		},
		{
			name:         "phi-3-mini",
			model:        "phi-3-mini",
			wantEncoding: "cl100k_base",
			reason:       "Phi uses cl100k_base approximation",
		},

		// Case insensitivity
		{
			name:         "GPT-4O uppercase",
			model:        "GPT-4O",
			wantEncoding: "o200k_base",
			reason:       "Case insensitive matching works",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getEncodingForModel(tt.model)
			if got != tt.wantEncoding {
				t.Errorf("getEncodingForModel(%q) = %q, want %q\nReason: %s",
					tt.model, got, tt.wantEncoding, tt.reason)
			}
		})
	}
}
