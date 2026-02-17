package tokenizer

// CountResult represents the result of token counting.
type CountResult struct {
	FilePath    string         `json:"file_path"`
	IsDirectory bool           `json:"is_directory,omitempty"`
	FileCount   int            `json:"file_count,omitempty"`
	FileSize    int            `json:"file_size"`
	Characters  int            `json:"characters"`
	Words       int            `json:"words"`
	Lines       int            `json:"lines"`
	Methods     []MethodResult `json:"methods"`
	Costs       []CostEstimate `json:"costs,omitempty"`
}

// MethodResult represents token count for a specific method.
type MethodResult struct {
	Name          string `json:"name"`
	DisplayName   string `json:"display_name"`
	Tokens        int    `json:"tokens"`
	IsExact       bool   `json:"is_exact"`
	ContextWindow int    `json:"context_window,omitempty"`
}

// CostEstimate represents cost estimation for a model.
type CostEstimate struct {
	Model     string  `json:"model"`
	Tokens    int     `json:"tokens"`
	Cost      float64 `json:"cost"`
	RatePer1M float64 `json:"rate_per_1m"`
}

// CounterOptions configures the counter.
type CounterOptions struct {
	CharsPerToken float64
	WordsPerToken float64
	VocabFile     string
	Provider      string
}
