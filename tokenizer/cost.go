package tokenizer

import (
	"strings"
)

// mainModels is the ordered set of models shown in default cost output.
var mainModels = []string{
	"gpt-5",
	"gpt-4o",
	"claude-4-sonnet",
	"claude-4.5-sonnet",
}

// characterBasedMethodPrefix identifies character-based approximation methods.
const characterBasedMethodPrefix = "character_based"

// CalculateCosts calculates cost estimates based on token counts.
// Pricing is sourced from the model registry (single source of truth).
// Only main models are included in the output.
func CalculateCosts(methods []MethodResult) []CostEstimate {
	costs := []CostEstimate{}

	tokenCount := getTokenCount(methods)
	if tokenCount == 0 {
		return costs
	}

	for _, modelName := range mainModels {
		meta := GetModelMetadata(modelName)
		if meta == nil || meta.InputPricePer1M == 0 {
			continue
		}
		costs = append(costs, CostEstimate{
			Model:     modelName,
			Tokens:    tokenCount,
			RatePer1M: meta.InputPricePer1M,
			Cost:      float64(tokenCount) * meta.InputPricePer1M / 1_000_000.0,
		})
	}

	return costs
}

// getTokenCount finds the best token count to use for cost calculation.
// Prefers exact BPE counts, then falls back to approximations.
func getTokenCount(methods []MethodResult) int {
	for _, method := range methods {
		if method.IsExact {
			return method.Tokens
		}
	}

	for _, method := range methods {
		if strings.HasPrefix(method.Name, characterBasedMethodPrefix) {
			return method.Tokens
		}
	}

	if len(methods) > 0 {
		return methods[0].Tokens
	}

	return 0
}

// GetPricingForModel returns pricing information for a specific model.
// Pricing is sourced from the model registry.
func GetPricingForModel(model string) *ModelMetadata {
	model = strings.ToLower(model)

	// Exact match
	if meta := GetModelMetadata(model); meta != nil {
		return meta
	}

	// Fuzzy match: check all models for substring containment
	for name, meta := range modelRegistry {
		if strings.Contains(strings.ToLower(name), model) ||
			strings.Contains(model, strings.ToLower(name)) {
			return &meta
		}
	}

	return nil
}
