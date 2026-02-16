package tokens

// models.go - Centralized model metadata registry
// Provides a single source of truth for model information including
// encoding, context windows, and pricing.

// Provider represents an LLM provider.
// Supported providers:
//   - openai: OpenAI models (GPT-4, GPT-5, etc.)
//   - anthropic: Anthropic models (Claude series)
//   - meta: Meta models (Llama series)
//   - deepseek: DeepSeek models (DeepSeek v2/v3)
//   - alibaba: Alibaba Cloud models (Qwen series)
//   - microsoft: Microsoft models (Phi series)
//   - google: Google models (Gemma series)
type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
	ProviderMeta      Provider = "meta"
	ProviderDeepSeek  Provider = "deepseek"
	ProviderAlibaba   Provider = "alibaba"
	ProviderMicrosoft Provider = "microsoft"
	ProviderGoogle    Provider = "google"
)

// ModelMetadata contains comprehensive information about an LLM model.
type ModelMetadata struct {
	Name             string   // Model identifier (e.g., "gpt-4o", "claude-4-sonnet")
	Provider         Provider // Provider who created the model
	Encoding         string   // BPE encoding name (e.g., "o200k_base", "cl100k_base")
	ContextWindow    int      // Maximum context window size in tokens
	InputPricePer1M  float64  // Input price per 1M tokens in USD
	OutputPricePer1M float64  // Output price per 1M tokens in USD
}

// modelRegistry is the central registry of all supported models.
// Initialized at package init time.
var modelRegistry map[string]ModelMetadata

// GetModelMetadata retrieves metadata for a given model name.
// Returns nil if model is not found in the registry.
func GetModelMetadata(modelName string) *ModelMetadata {
	if meta, ok := modelRegistry[modelName]; ok {
		return &meta
	}
	return nil
}

// ListModels returns all registered model names.
func ListModels() []string {
	models := make([]string, 0, len(modelRegistry))
	for name := range modelRegistry {
		models = append(models, name)
	}
	return models
}

// ListModelsByProvider returns all models from a specific provider.
func ListModelsByProvider(provider Provider) []ModelMetadata {
	models := make([]ModelMetadata, 0)
	for _, meta := range modelRegistry {
		if meta.Provider == provider {
			models = append(models, meta)
		}
	}
	return models
}

// GetProviderForModel returns the provider for a given model name.
// Returns empty string if model is not registered.
func GetProviderForModel(modelName string) Provider {
	if meta := GetModelMetadata(modelName); meta != nil {
		return meta.Provider
	}
	return ""
}

// IsOpenSourceModel returns true if the model is from an open-source provider
// (not OpenAI or Anthropic).
func IsOpenSourceModel(modelName string) bool {
	provider := GetProviderForModel(modelName)
	return provider != "" &&
		provider != ProviderOpenAI &&
		provider != ProviderAnthropic
}

func init() {
	modelRegistry = make(map[string]ModelMetadata)

	// OpenAI Models - GPT-5 series (o200k_base)
	modelRegistry["gpt-5"] = ModelMetadata{
		Name:             "gpt-5",
		Provider:         ProviderOpenAI,
		Encoding:         "o200k_base",
		ContextWindow:    200000,
		InputPricePer1M:  5.00,
		OutputPricePer1M: 15.00,
	}
	modelRegistry["gpt-5-mini"] = ModelMetadata{
		Name:             "gpt-5-mini",
		Provider:         ProviderOpenAI,
		Encoding:         "o200k_base",
		ContextWindow:    200000,
		InputPricePer1M:  1.00,
		OutputPricePer1M: 3.00,
	}

	// OpenAI Models - GPT-4.1 series (o200k_base)
	modelRegistry["gpt-4.1"] = ModelMetadata{
		Name:             "gpt-4.1",
		Provider:         ProviderOpenAI,
		Encoding:         "o200k_base",
		ContextWindow:    128000,
		InputPricePer1M:  3.00,
		OutputPricePer1M: 9.00,
	}
	modelRegistry["gpt-4.1-mini"] = ModelMetadata{
		Name:             "gpt-4.1-mini",
		Provider:         ProviderOpenAI,
		Encoding:         "o200k_base",
		ContextWindow:    128000,
		InputPricePer1M:  0.60,
		OutputPricePer1M: 1.80,
	}
	modelRegistry["gpt-4.1-nano"] = ModelMetadata{
		Name:             "gpt-4.1-nano",
		Provider:         ProviderOpenAI,
		Encoding:         "o200k_base",
		ContextWindow:    128000,
		InputPricePer1M:  0.30,
		OutputPricePer1M: 0.90,
	}

	// OpenAI Models - GPT-4o series (o200k_base)
	modelRegistry["gpt-4o"] = ModelMetadata{
		Name:             "gpt-4o",
		Provider:         ProviderOpenAI,
		Encoding:         "o200k_base",
		ContextWindow:    128000,
		InputPricePer1M:  2.50,
		OutputPricePer1M: 10.00,
	}
	modelRegistry["gpt-4o-mini"] = ModelMetadata{
		Name:             "gpt-4o-mini",
		Provider:         ProviderOpenAI,
		Encoding:         "o200k_base",
		ContextWindow:    128000,
		InputPricePer1M:  0.15,
		OutputPricePer1M: 0.60,
	}

	// OpenAI Models - o-series (o200k_base)
	modelRegistry["o3"] = ModelMetadata{
		Name:             "o3",
		Provider:         ProviderOpenAI,
		Encoding:         "o200k_base",
		ContextWindow:    200000,
		InputPricePer1M:  10.00,
		OutputPricePer1M: 30.00,
	}
	modelRegistry["o3-mini"] = ModelMetadata{
		Name:             "o3-mini",
		Provider:         ProviderOpenAI,
		Encoding:         "o200k_base",
		ContextWindow:    200000,
		InputPricePer1M:  1.00,
		OutputPricePer1M: 3.00,
	}
	modelRegistry["o4-mini"] = ModelMetadata{
		Name:             "o4-mini",
		Provider:         ProviderOpenAI,
		Encoding:         "o200k_base",
		ContextWindow:    200000,
		InputPricePer1M:  1.00,
		OutputPricePer1M: 3.00,
	}

	// OpenAI Models - Legacy (cl100k_base)
	modelRegistry["gpt-4"] = ModelMetadata{
		Name:             "gpt-4",
		Provider:         ProviderOpenAI,
		Encoding:         "cl100k_base",
		ContextWindow:    8192,
		InputPricePer1M:  30.00,
		OutputPricePer1M: 60.00,
	}
	modelRegistry["gpt-4-turbo"] = ModelMetadata{
		Name:             "gpt-4-turbo",
		Provider:         ProviderOpenAI,
		Encoding:         "cl100k_base",
		ContextWindow:    128000,
		InputPricePer1M:  10.00,
		OutputPricePer1M: 30.00,
	}
	modelRegistry["gpt-3.5-turbo"] = ModelMetadata{
		Name:             "gpt-3.5-turbo",
		Provider:         ProviderOpenAI,
		Encoding:         "cl100k_base",
		ContextWindow:    16385,
		InputPricePer1M:  0.50,
		OutputPricePer1M: 1.50,
	}

	// Anthropic Models - Claude (approximation)
	modelRegistry["claude-4-opus"] = ModelMetadata{
		Name:             "claude-4-opus",
		Provider:         ProviderAnthropic,
		Encoding:         "claude_approx",
		ContextWindow:    200000,
		InputPricePer1M:  15.00,
		OutputPricePer1M: 75.00,
	}
	modelRegistry["claude-4-sonnet"] = ModelMetadata{
		Name:             "claude-4-sonnet",
		Provider:         ProviderAnthropic,
		Encoding:         "claude_approx",
		ContextWindow:    200000,
		InputPricePer1M:  3.00,
		OutputPricePer1M: 15.00,
	}
	modelRegistry["claude-4.5-sonnet"] = ModelMetadata{
		Name:             "claude-4.5-sonnet",
		Provider:         ProviderAnthropic,
		Encoding:         "claude_approx",
		ContextWindow:    200000,
		InputPricePer1M:  3.00,
		OutputPricePer1M: 15.00,
	}
	modelRegistry["claude-3.7-sonnet"] = ModelMetadata{
		Name:             "claude-3.7-sonnet",
		Provider:         ProviderAnthropic,
		Encoding:         "claude_approx",
		ContextWindow:    200000,
		InputPricePer1M:  3.00,
		OutputPricePer1M: 15.00,
	}
	modelRegistry["claude-3.5-sonnet"] = ModelMetadata{
		Name:             "claude-3.5-sonnet",
		Provider:         ProviderAnthropic,
		Encoding:         "claude_approx",
		ContextWindow:    200000,
		InputPricePer1M:  3.00,
		OutputPricePer1M: 15.00,
	}
	modelRegistry["claude-3-opus"] = ModelMetadata{
		Name:             "claude-3-opus",
		Provider:         ProviderAnthropic,
		Encoding:         "claude_approx",
		ContextWindow:    200000,
		InputPricePer1M:  15.00,
		OutputPricePer1M: 75.00,
	}
	modelRegistry["claude-3-sonnet"] = ModelMetadata{
		Name:             "claude-3-sonnet",
		Provider:         ProviderAnthropic,
		Encoding:         "claude_approx",
		ContextWindow:    200000,
		InputPricePer1M:  3.00,
		OutputPricePer1M: 15.00,
	}
	modelRegistry["claude-3-haiku"] = ModelMetadata{
		Name:             "claude-3-haiku",
		Provider:         ProviderAnthropic,
		Encoding:         "claude_approx",
		ContextWindow:    200000,
		InputPricePer1M:  0.25,
		OutputPricePer1M: 1.25,
	}
	// Legacy name for backward compatibility
	modelRegistry["claude-3"] = ModelMetadata{
		Name:             "claude-3",
		Provider:         ProviderAnthropic,
		Encoding:         "claude_approx",
		ContextWindow:    200000,
		InputPricePer1M:  3.00,
		OutputPricePer1M: 15.00,
	}

	// Meta Models - Llama 3 series (cl100k_base BPE approximation)
	// Note: Llama uses its own tokenizer, but cl100k_base provides reasonable approximation
	modelRegistry["llama-3.1-8b"] = ModelMetadata{
		Name:             "llama-3.1-8b",
		Provider:         ProviderMeta,
		Encoding:         "cl100k_base",
		ContextWindow:    128000,
		InputPricePer1M:  0.0,
		OutputPricePer1M: 0.0,
	}
	modelRegistry["llama-3.1-70b"] = ModelMetadata{
		Name:             "llama-3.1-70b",
		Provider:         ProviderMeta,
		Encoding:         "cl100k_base",
		ContextWindow:    128000,
		InputPricePer1M:  0.0,
		OutputPricePer1M: 0.0,
	}
	modelRegistry["llama-3.1-405b"] = ModelMetadata{
		Name:             "llama-3.1-405b",
		Provider:         ProviderMeta,
		Encoding:         "cl100k_base",
		ContextWindow:    128000,
		InputPricePer1M:  0.0,
		OutputPricePer1M: 0.0,
	}
	modelRegistry["llama-4-scout"] = ModelMetadata{
		Name:             "llama-4-scout",
		Provider:         ProviderMeta,
		Encoding:         "cl100k_base",
		ContextWindow:    128000,
		InputPricePer1M:  0.0,
		OutputPricePer1M: 0.0,
	}
	modelRegistry["llama-4-maverick"] = ModelMetadata{
		Name:             "llama-4-maverick",
		Provider:         ProviderMeta,
		Encoding:         "cl100k_base",
		ContextWindow:    128000,
		InputPricePer1M:  0.0,
		OutputPricePer1M: 0.0,
	}

	// DeepSeek Models (cl100k_base BPE approximation)
	modelRegistry["deepseek-v2"] = ModelMetadata{
		Name:             "deepseek-v2",
		Provider:         ProviderDeepSeek,
		Encoding:         "cl100k_base",
		ContextWindow:    128000,
		InputPricePer1M:  0.0,
		OutputPricePer1M: 0.0,
	}
	modelRegistry["deepseek-v3"] = ModelMetadata{
		Name:             "deepseek-v3",
		Provider:         ProviderDeepSeek,
		Encoding:         "cl100k_base",
		ContextWindow:    128000,
		InputPricePer1M:  0.0,
		OutputPricePer1M: 0.0,
	}
	modelRegistry["deepseek-coder-v2"] = ModelMetadata{
		Name:             "deepseek-coder-v2",
		Provider:         ProviderDeepSeek,
		Encoding:         "cl100k_base",
		ContextWindow:    128000,
		InputPricePer1M:  0.0,
		OutputPricePer1M: 0.0,
	}

	// Alibaba Models - Qwen 2/3 series (cl100k_base BPE compatible)
	modelRegistry["qwen-2.5-7b"] = ModelMetadata{
		Name:             "qwen-2.5-7b",
		Provider:         ProviderAlibaba,
		Encoding:         "cl100k_base",
		ContextWindow:    32768,
		InputPricePer1M:  0.0,
		OutputPricePer1M: 0.0,
	}
	modelRegistry["qwen-2.5-14b"] = ModelMetadata{
		Name:             "qwen-2.5-14b",
		Provider:         ProviderAlibaba,
		Encoding:         "cl100k_base",
		ContextWindow:    32768,
		InputPricePer1M:  0.0,
		OutputPricePer1M: 0.0,
	}
	modelRegistry["qwen-2.5-72b"] = ModelMetadata{
		Name:             "qwen-2.5-72b",
		Provider:         ProviderAlibaba,
		Encoding:         "cl100k_base",
		ContextWindow:    32768,
		InputPricePer1M:  0.0,
		OutputPricePer1M: 0.0,
	}
	modelRegistry["qwen-3-72b"] = ModelMetadata{
		Name:             "qwen-3-72b",
		Provider:         ProviderAlibaba,
		Encoding:         "cl100k_base",
		ContextWindow:    32768,
		InputPricePer1M:  0.0,
		OutputPricePer1M: 0.0,
	}

	// Microsoft Models - Phi-3 series (cl100k_base BPE compatible)
	modelRegistry["phi-3-mini"] = ModelMetadata{
		Name:             "phi-3-mini",
		Provider:         ProviderMicrosoft,
		Encoding:         "cl100k_base",
		ContextWindow:    128000,
		InputPricePer1M:  0.0,
		OutputPricePer1M: 0.0,
	}
	modelRegistry["phi-3-small"] = ModelMetadata{
		Name:             "phi-3-small",
		Provider:         ProviderMicrosoft,
		Encoding:         "cl100k_base",
		ContextWindow:    128000,
		InputPricePer1M:  0.0,
		OutputPricePer1M: 0.0,
	}
	modelRegistry["phi-3-medium"] = ModelMetadata{
		Name:             "phi-3-medium",
		Provider:         ProviderMicrosoft,
		Encoding:         "cl100k_base",
		ContextWindow:    128000,
		InputPricePer1M:  0.0,
		OutputPricePer1M: 0.0,
	}
}
