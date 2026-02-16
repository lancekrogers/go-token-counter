package tokens

import (
	"testing"
)

func TestCalculateCosts(t *testing.T) {
	methods := []MethodResult{
		{
			Name:    "tiktoken_gpt_4o",
			Tokens:  1000,
			IsExact: true,
		},
	}

	costs := CalculateCosts(methods)
	if len(costs) == 0 {
		t.Fatal("Expected cost estimates, got none")
	}

	// Verify main models are included
	mainModels := map[string]bool{
		"gpt-5":             false,
		"gpt-4o":            false,
		"claude-4-sonnet":   false,
		"claude-4.5-sonnet": false,
	}

	for _, cost := range costs {
		if _, ok := mainModels[cost.Model]; ok {
			mainModels[cost.Model] = true
		}

		if cost.Tokens != 1000 {
			t.Errorf("Cost for %s: tokens = %d, want 1000", cost.Model, cost.Tokens)
		}
		if cost.RatePer1M <= 0 {
			t.Errorf("Cost for %s: rate per 1M should be positive, got %f", cost.Model, cost.RatePer1M)
		}
		if cost.Cost <= 0 {
			t.Errorf("Cost for %s: cost should be positive, got %f", cost.Model, cost.Cost)
		}

		// Verify per-1M calculation: cost = tokens * rate / 1_000_000
		expectedCost := float64(cost.Tokens) * cost.RatePer1M / 1_000_000.0
		if cost.Cost != expectedCost {
			t.Errorf("Cost for %s: got %f, want %f", cost.Model, cost.Cost, expectedCost)
		}
	}

	for model, found := range mainModels {
		if !found {
			t.Errorf("Main model %s not found in cost output", model)
		}
	}
}

func TestCalculateCosts_EmptyMethods(t *testing.T) {
	costs := CalculateCosts([]MethodResult{})
	if len(costs) != 0 {
		t.Errorf("Expected no costs for empty methods, got %d", len(costs))
	}
}

func TestCalculateCosts_ZeroTokens(t *testing.T) {
	methods := []MethodResult{
		{Name: "test", Tokens: 0, IsExact: true},
	}

	costs := CalculateCosts(methods)
	if len(costs) != 0 {
		t.Errorf("Expected no costs for zero tokens, got %d", len(costs))
	}
}

func TestGetTokenCount(t *testing.T) {
	tests := []struct {
		name     string
		methods  []MethodResult
		expected int
	}{
		{
			name: "Prefers exact GPT count",
			methods: []MethodResult{
				{Name: "character_based", Tokens: 100, IsExact: false},
				{Name: "tiktoken_gpt_4o", Tokens: 80, IsExact: true},
			},
			expected: 80,
		},
		{
			name: "Falls back to character-based",
			methods: []MethodResult{
				{Name: "character_based_div4", DisplayName: "Character-based (÷4.0)", Tokens: 100, IsExact: false},
				{Name: "word_based", Tokens: 90, IsExact: false},
			},
			expected: 100,
		},
		{
			name: "Falls back to first method",
			methods: []MethodResult{
				{Name: "word_based", Tokens: 90, IsExact: false},
				{Name: "whitespace", Tokens: 85, IsExact: false},
			},
			expected: 90,
		},
		{
			name:     "Returns zero for empty",
			methods:  []MethodResult{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTokenCount(tt.methods)
			if result != tt.expected {
				t.Errorf("getTokenCount() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestIsMainModel(t *testing.T) {
	mainModels := []string{"gpt-5", "gpt-4o", "claude-4-sonnet", "claude-4.5-sonnet"}
	for _, model := range mainModels {
		if !isMainModel(model) {
			t.Errorf("isMainModel(%q) = false, want true", model)
		}
	}

	nonMainModels := []string{"gpt-4", "gpt-3.5-turbo", "claude-3-haiku", "llama-3.1-8b"}
	for _, model := range nonMainModels {
		if isMainModel(model) {
			t.Errorf("isMainModel(%q) = true, want false", model)
		}
	}
}

func TestGetPricingForModel(t *testing.T) {
	tests := []struct {
		name      string
		model     string
		expectNil bool
		checkRate bool
		rate      float64
	}{
		{"Exact match", "gpt-5", false, true, 1.25},
		{"Case insensitive", "GPT-5", false, true, 1.25},
		{"Partial match", "gpt-4o", false, true, 2.50},
		{"Claude model", "claude-4-opus", false, true, 15.00},
		{"Unknown model", "nonexistent-model-xyz", true, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pricing := GetPricingForModel(tt.model)
			if tt.expectNil {
				if pricing != nil {
					t.Errorf("Expected nil pricing for %q, got %+v", tt.model, pricing)
				}
				return
			}
			if pricing == nil {
				t.Fatalf("Expected pricing for %q, got nil", tt.model)
			}
			if tt.checkRate && pricing.InputPer1M != tt.rate {
				t.Errorf("Pricing for %q: input rate = %f, want %f", tt.model, pricing.InputPer1M, tt.rate)
			}
		})
	}
}

func TestCalculateCosts_PerMillionFormat(t *testing.T) {
	methods := []MethodResult{
		{Name: "tiktoken_gpt_5", Tokens: 1_000_000, IsExact: true},
	}

	costs := CalculateCosts(methods)

	for _, cost := range costs {
		if cost.Model == "gpt-5" {
			// At 1M tokens, cost should equal rate
			if cost.Cost != cost.RatePer1M {
				t.Errorf("At 1M tokens, cost (%f) should equal rate per 1M (%f)", cost.Cost, cost.RatePer1M)
			}
			return
		}
	}
	t.Error("gpt-5 not found in costs")
}
