package trace

import "encoding/json"

// OpenAIRequest is a minimal model of an OpenAI-compatible chat completion request.
type OpenAIRequest struct {
	Model       string                   `json:"model"`
	Messages    []map[string]interface{} `json:"messages"`
	Temperature float64                  `json:"temperature,omitempty"`
	MaxTokens   int                      `json:"max_tokens,omitempty"`
	Stream      bool                     `json:"stream,omitempty"`
}

// OpenAIResponse is a minimal model of an OpenAI-compatible chat completion response.
type OpenAIResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// ParseRequest unmarshals body into an OpenAIRequest.
// Returns a zero-value struct (not an error) on empty or non-JSON body,
// so callers can always proceed.
func ParseRequest(body []byte) (*OpenAIRequest, error) {
	var req OpenAIRequest
	if len(body) == 0 {
		return &req, nil
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

// ParseResponse unmarshals body into an OpenAIResponse.
func ParseResponse(body []byte) (*OpenAIResponse, error) {
	var resp OpenAIResponse
	if len(body) == 0 {
		return &resp, nil
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CalculateCost returns the estimated USD cost for a call based on
// 2024 published per-token pricing. Unknown models fall back to gpt-4o rates.
func CalculateCost(model string, promptTokens, completionTokens int) float64 {
	type pricing struct{ input, output float64 } // USD per 1M tokens
	table := map[string]pricing{
		"gpt-4o":            {2.50, 10.00},
		"gpt-4o-mini":       {0.15, 0.60},
		"gpt-4-turbo":       {10.00, 30.00},
		"gpt-3.5-turbo":     {0.50, 1.50},
		"claude-3-5-sonnet": {3.00, 15.00},
		"claude-3-opus":     {15.00, 75.00},
		"qwen-turbo":        {0.056, 0.14},
		"qwen-plus":         {0.14, 0.42},
		"qwen-max":          {0.56, 1.68},
		"deepseek-chat":     {0.14, 0.28},
	}
	p, ok := table[model]
	if !ok {
		p = table["gpt-4o"]
	}
	return float64(promptTokens)*p.input/1_000_000 +
		float64(completionTokens)*p.output/1_000_000
}
