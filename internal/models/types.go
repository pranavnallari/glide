package models

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type UnifiedRequest struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model"`
	Strategy    string    `json:"strategy"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
	StreamOK    bool      `json:"stream_ok"`
	Timeout     int       `json:"timeout"`
	MaxRetries  int       `json:"max_retries"`
}

type UnifiedResponse struct {
	Id       string     `json:"id"`
	Provider string     `json:"provider"`
	Content  string     `json:"content"`
	Model    string     `json:"model"`
	Usage    TokenCount `json:"usage"`
}

type TokenCount struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
