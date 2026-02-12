package tokens

import (
	"strings"
)

// ModelPricing represents pricing for a model.
type ModelPricing struct {
	Model       string
	InputPer1K  float64
	OutputPer1K float64
}

// Common model pricing (as of 2024).
var modelPricing = []ModelPricing{
	// OpenAI GPT-4
	{Model: "gpt-4", InputPer1K: 0.01, OutputPer1K: 0.03},
	{Model: "gpt-4-turbo", InputPer1K: 0.01, OutputPer1K: 0.03},
	{Model: "gpt-4o", InputPer1K: 0.005, OutputPer1K: 0.015},

	// OpenAI GPT-3.5
	{Model: "gpt-3.5-turbo", InputPer1K: 0.0005, OutputPer1K: 0.0015},
	{Model: "gpt-3.5-turbo-16k", InputPer1K: 0.003, OutputPer1K: 0.004},

	// Anthropic Claude
	{Model: "claude-3-opus", InputPer1K: 0.015, OutputPer1K: 0.075},
	{Model: "claude-3-sonnet", InputPer1K: 0.003, OutputPer1K: 0.015},
	{Model: "claude-3-haiku", InputPer1K: 0.00025, OutputPer1K: 0.00125},
	{Model: "claude-2.1", InputPer1K: 0.008, OutputPer1K: 0.024},
	{Model: "claude-2", InputPer1K: 0.008, OutputPer1K: 0.024},
}

// CalculateCosts calculates cost estimates based on token counts.
func CalculateCosts(methods []MethodResult) []CostEstimate {
	costs := []CostEstimate{}

	tokenCount := getTokenCount(methods)
	if tokenCount == 0 {
		return costs
	}

	for _, pricing := range modelPricing {
		if isMainModel(pricing.Model) {
			cost := CostEstimate{
				Model:     pricing.Model,
				Tokens:    tokenCount,
				RatePer1K: pricing.InputPer1K,
				Cost:      float64(tokenCount) * pricing.InputPer1K / 1000.0,
			}
			costs = append(costs, cost)
		}
	}

	return costs
}

// getTokenCount finds the best token count to use for cost calculation.
func getTokenCount(methods []MethodResult) int {
	for _, method := range methods {
		if method.IsExact && strings.Contains(strings.ToLower(method.Name), "gpt") {
			return method.Tokens
		}
	}

	for _, method := range methods {
		if strings.Contains(method.Name, "Character-based") {
			return method.Tokens
		}
	}

	if len(methods) > 0 {
		return methods[0].Tokens
	}

	return 0
}

// isMainModel checks if a model should be shown in default cost output.
func isMainModel(model string) bool {
	mainModels := []string{
		"gpt-4",
		"gpt-3.5-turbo",
		"claude-3-opus",
		"claude-3-sonnet",
	}

	for _, main := range mainModels {
		if model == main {
			return true
		}
	}

	return false
}

// GetPricingForModel returns pricing information for a specific model.
func GetPricingForModel(model string) *ModelPricing {
	model = strings.ToLower(model)

	for _, pricing := range modelPricing {
		if strings.ToLower(pricing.Model) == model {
			return &pricing
		}
	}

	for _, pricing := range modelPricing {
		if strings.Contains(strings.ToLower(pricing.Model), model) ||
			strings.Contains(model, strings.ToLower(pricing.Model)) {
			return &pricing
		}
	}

	return nil
}
