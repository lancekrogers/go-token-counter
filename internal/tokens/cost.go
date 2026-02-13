package tokens

import (
	"strings"
)

// ModelPricing represents pricing for a model.
type ModelPricing struct {
	Model       string
	InputPer1M  float64
	OutputPer1M float64
}

// Model pricing data last updated: 2026-02-13
// Sources:
// - OpenAI: https://openai.com/api/pricing/
// - Anthropic: https://platform.claude.com/docs/en/about-claude/pricing
// Pricing is stored as cost per 1M tokens (industry standard).
var modelPricing = []ModelPricing{
	// OpenAI GPT-5 series (2026)
	{Model: "gpt-5", InputPer1M: 1.25, OutputPer1M: 10.00},
	{Model: "gpt-5-mini", InputPer1M: 0.25, OutputPer1M: 2.00},

	// OpenAI GPT-4.1 series (2026)
	{Model: "gpt-4.1", InputPer1M: 2.00, OutputPer1M: 8.00},
	{Model: "gpt-4.1-mini", InputPer1M: 0.40, OutputPer1M: 1.60},
	{Model: "gpt-4.1-nano", InputPer1M: 0.10, OutputPer1M: 0.40},

	// OpenAI GPT-4o series
	{Model: "gpt-4o", InputPer1M: 2.50, OutputPer1M: 10.00},
	{Model: "gpt-4o-mini", InputPer1M: 0.15, OutputPer1M: 0.60},

	// OpenAI o-series reasoning models (2026)
	{Model: "o3", InputPer1M: 2.00, OutputPer1M: 8.00},
	{Model: "o3-mini", InputPer1M: 1.10, OutputPer1M: 4.40},
	{Model: "o4-mini", InputPer1M: 1.10, OutputPer1M: 4.40},

	// OpenAI Legacy
	{Model: "gpt-4", InputPer1M: 10.00, OutputPer1M: 30.00},
	{Model: "gpt-4-turbo", InputPer1M: 10.00, OutputPer1M: 30.00},
	{Model: "gpt-3.5-turbo", InputPer1M: 0.50, OutputPer1M: 1.50},

	// Anthropic Claude 4.5 series (2026)
	{Model: "claude-4.5-sonnet", InputPer1M: 3.00, OutputPer1M: 15.00},

	// Anthropic Claude 4 series (2026)
	{Model: "claude-4-opus", InputPer1M: 15.00, OutputPer1M: 75.00},
	{Model: "claude-4-sonnet", InputPer1M: 3.00, OutputPer1M: 15.00},

	// Anthropic Claude 3.x series
	{Model: "claude-3.7-sonnet", InputPer1M: 3.00, OutputPer1M: 15.00},
	{Model: "claude-3.5-sonnet", InputPer1M: 3.00, OutputPer1M: 15.00},
	{Model: "claude-3-opus", InputPer1M: 15.00, OutputPer1M: 75.00},
	{Model: "claude-3-sonnet", InputPer1M: 3.00, OutputPer1M: 15.00},
	{Model: "claude-3-haiku", InputPer1M: 0.25, OutputPer1M: 1.25},
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
				RatePer1M: pricing.InputPer1M,
				Cost:      float64(tokenCount) * pricing.InputPer1M / 1_000_000.0,
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
		"gpt-5",
		"gpt-4o",
		"claude-4-sonnet",
		"claude-4.5-sonnet",
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
