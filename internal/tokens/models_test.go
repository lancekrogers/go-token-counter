package tokens

import (
	"testing"
)

func TestGetModelMetadata(t *testing.T) {
	tests := []struct {
		name         string
		modelName    string
		wantProvider Provider
		wantEncoding string
		wantNil      bool
	}{
		{
			name:         "gpt-4o",
			modelName:    "gpt-4o",
			wantProvider: ProviderOpenAI,
			wantEncoding: "o200k_base",
		},
		{
			name:         "gpt-5",
			modelName:    "gpt-5",
			wantProvider: ProviderOpenAI,
			wantEncoding: "o200k_base",
		},
		{
			name:         "gpt-4",
			modelName:    "gpt-4",
			wantProvider: ProviderOpenAI,
			wantEncoding: "cl100k_base",
		},
		{
			name:         "claude-4-sonnet",
			modelName:    "claude-4-sonnet",
			wantProvider: ProviderAnthropic,
			wantEncoding: "claude_approx",
		},
		{
			name:      "unknown model",
			modelName: "unknown-model",
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := GetModelMetadata(tt.modelName)
			if tt.wantNil {
				if meta != nil {
					t.Errorf("GetModelMetadata(%q) = %+v, want nil", tt.modelName, meta)
				}
				return
			}
			if meta == nil {
				t.Fatalf("GetModelMetadata(%q) = nil, want non-nil", tt.modelName)
			}
			if meta.Provider != tt.wantProvider {
				t.Errorf("Provider = %v, want %v", meta.Provider, tt.wantProvider)
			}
			if meta.Encoding != tt.wantEncoding {
				t.Errorf("Encoding = %v, want %v", meta.Encoding, tt.wantEncoding)
			}
		})
	}
}

func TestListModelsByProvider(t *testing.T) {
	openAIModels := ListModelsByProvider(ProviderOpenAI)
	if len(openAIModels) < 10 {
		t.Errorf("Expected at least 10 OpenAI models, got %d", len(openAIModels))
	}

	anthropicModels := ListModelsByProvider(ProviderAnthropic)
	if len(anthropicModels) < 8 {
		t.Errorf("Expected at least 8 Anthropic models, got %d", len(anthropicModels))
	}
}

func TestAllOpenAIModelsRegistered(t *testing.T) {
	requiredModels := []string{
		"gpt-5", "gpt-5-mini",
		"gpt-4.1", "gpt-4.1-mini", "gpt-4.1-nano",
		"gpt-4o", "gpt-4o-mini",
		"o3", "o3-mini", "o4-mini",
		"gpt-4", "gpt-4-turbo", "gpt-3.5-turbo",
	}

	for _, modelName := range requiredModels {
		meta := GetModelMetadata(modelName)
		if meta == nil {
			t.Errorf("Model %q not registered in modelRegistry", modelName)
			continue
		}
		if meta.Provider != ProviderOpenAI {
			t.Errorf("Model %q has wrong provider: got %v, want %v",
				modelName, meta.Provider, ProviderOpenAI)
		}
		if meta.Encoding == "" {
			t.Errorf("Model %q has empty encoding", modelName)
		}
		if meta.ContextWindow == 0 {
			t.Errorf("Model %q has zero context window", modelName)
		}
	}
}

func TestOpenAIEncodingCorrectness(t *testing.T) {
	tests := []struct {
		model        string
		wantEncoding string
	}{
		{"gpt-5", "o200k_base"},
		{"gpt-5-mini", "o200k_base"},
		{"gpt-4.1", "o200k_base"},
		{"gpt-4.1-mini", "o200k_base"},
		{"gpt-4.1-nano", "o200k_base"},
		{"gpt-4o", "o200k_base"},
		{"gpt-4o-mini", "o200k_base"},
		{"o3", "o200k_base"},
		{"o3-mini", "o200k_base"},
		{"o4-mini", "o200k_base"},
		{"gpt-4", "cl100k_base"},
		{"gpt-4-turbo", "cl100k_base"},
		{"gpt-3.5-turbo", "cl100k_base"},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			meta := GetModelMetadata(tt.model)
			if meta == nil {
				t.Fatalf("Model %q not found in registry", tt.model)
			}
			if meta.Encoding != tt.wantEncoding {
				t.Errorf("Model %q encoding = %q, want %q",
					tt.model, meta.Encoding, tt.wantEncoding)
			}
		})
	}
}

func TestRegistryMatchesTokenizer(t *testing.T) {
	models := []string{"gpt-4o", "gpt-5", "gpt-4", "gpt-3.5-turbo"}

	for _, modelName := range models {
		t.Run(modelName, func(t *testing.T) {
			meta := GetModelMetadata(modelName)
			if meta == nil {
				t.Fatalf("Model %q not in registry", modelName)
			}
			registryEncoding := meta.Encoding
			tokenizerEncoding := getEncodingForModel(modelName)

			if registryEncoding != tokenizerEncoding {
				t.Errorf("Encoding mismatch for %q: registry=%q, tokenizer=%q",
					modelName, registryEncoding, tokenizerEncoding)
			}
		})
	}
}
