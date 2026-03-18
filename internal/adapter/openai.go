package adapter

import (
	"errors"

	"github.com/pranavnallari/glide/internal/models"
)

const (
	DefaultMaxRetries = 3
	DefaultTimeout    = 30
	DefaultMaxTokens  = 2048
)

type OpenAIRequestMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type OpenAIRequest struct {
	Model               string                 `json:"model"`
	Messages            []OpenAIRequestMessage `json:"messages"`
	MaxCompletionTokens int                    `json:"max_completion_tokens"`
	Stream              bool                   `json:"stream"`
	StreamOptions       *StreamOptions         `json:"stream_options,omitempty"`
	Temperature         float64                `json:"temperature"`
}

type OpenAIResponseMessage struct {
	Content string `json:"content"`
	Refusal string `json:"refusal"`
	Role    string `json:"role"`
}

type Choice struct {
	Message OpenAIResponseMessage `json:"message"`
}

type OpenAITokenCount struct {
	CompletionTokens int `json:"completion_tokens"`
	PromptTokens     int `json:"prompt_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type OpenAIResponse struct {
	Id      string           `json:"id"`
	Choices []Choice         `json:"choices"`
	Model   string           `json:"model"`
	Usage   OpenAITokenCount `json:"usage"`
}

func ToOpenAI(req *models.UnifiedRequest) (*OpenAIRequest, error) {
	if len(req.Messages) == 0 {
		return nil, errors.New("messages cannot be empty")
	}
	if req.Model == "" {
		return nil, errors.New("model cannot be empty")
	}
	if req.Temperature < 0 || req.Temperature > 2.0 {
		return nil, errors.New("temperature must be between 0 and 2.0")
	}
	if req.MaxTokens < 0 {
		return nil, errors.New("max tokens cannot be negative")
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = DefaultMaxTokens
	}
	if req.MaxRetries < 0 {
		req.MaxRetries = DefaultMaxRetries
	}
	if req.Timeout < 0 {
		req.Timeout = DefaultTimeout
	}

	var r OpenAIRequest
	r.MaxCompletionTokens = req.MaxTokens
	r.Model = req.Model
	r.Stream = req.StreamOK
	if req.StreamOK {
		r.StreamOptions = &StreamOptions{IncludeUsage: true}
	}

	r.Temperature = req.Temperature

	messages := make([]OpenAIRequestMessage, len(req.Messages))

	for i, v := range req.Messages {
		messages[i] = OpenAIRequestMessage{
			Role:    v.Role,
			Content: v.Content,
		}
	}

	r.Messages = messages

	return &r, nil
}

func FromOpenAI(res *OpenAIResponse) (*models.UnifiedResponse, error) {
	var r models.UnifiedResponse
	r.Id = res.Id
	r.Provider = "openai"
	r.Model = res.Model
	r.Usage = models.TokenCount{
		CompletionTokens: res.Usage.CompletionTokens,
		PromptTokens:     res.Usage.PromptTokens,
		TotalTokens:      res.Usage.TotalTokens,
	}

	if len(res.Choices) == 0 {
		return nil, errors.New("openai returned no choices")
	}

	r.Content = res.Choices[0].Message.Content
	return &r, nil
}
