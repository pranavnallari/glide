package adapter

import (
	"errors"
	"strings"

	"github.com/pranavnallari/glide/internal/models"
)

type AnthropicRequestMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AnthropicRequest struct {
	MaxTokens   int                       `json:"max_tokens"`
	Model       string                    `json:"model"`
	Stream      bool                      `json:"stream"`
	Temperature float64                   `json:"temperature"`
	Messages    []AnthropicRequestMessage `json:"messages"`
	System      string                    `json:"system,omitempty"`
}

type AnthropicTokenCount struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type AnthropicResponseContent struct {
	Text string `json:"text"`
}
type AnthropicResponse struct {
	Id      string                     `json:"id"`
	Content []AnthropicResponseContent `json:"content"`
	Model   string                     `json:"model"`
	Type    string                     `json:"type"`
	Usage   AnthropicTokenCount        `json:"usage"`
}

func ToAnthropic(req *models.UnifiedRequest) (*AnthropicRequest, error) {
	if len(req.Messages) == 0 {
		return nil, errors.New("messages cannot be empty")
	}
	if req.Model == "" {
		return nil, errors.New("model cannot be empty")
	}
	if req.Temperature < 0 || req.Temperature > 1.0 {
		return nil, errors.New("temperature must be between 0 and 1.0")
	}
	if req.MaxTokens < 0 {
		return nil, errors.New("max tokens cannot be negative")
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = DefaultMaxTokens
	}

	var r AnthropicRequest
	r.MaxTokens = req.MaxTokens
	r.Model = req.Model
	r.Stream = req.StreamOK

	r.Temperature = req.Temperature

	var sysMessages []string
	var nonSysMessages []AnthropicRequestMessage

	for _, v := range req.Messages {
		if v.Role == "system" {
			sysMessages = append(sysMessages, v.Content)
		} else {
			nonSysMessages = append(nonSysMessages, AnthropicRequestMessage{
				Role:    v.Role,
				Content: v.Content,
			})
		}
	}

	r.System = strings.Join(sysMessages, "\n")
	if len(nonSysMessages) == 0 {
		return nil, errors.New("no user messages to submit")
	}
	r.Messages = nonSysMessages

	return &r, nil
}

func FromAnthropic(res *AnthropicResponse) (*models.UnifiedResponse, error) {
	var r models.UnifiedResponse
	r.Id = res.Id
	r.Model = res.Model
	r.Provider = "anthropic"
	r.Usage = models.TokenCount{
		CompletionTokens: res.Usage.OutputTokens,
		PromptTokens:     res.Usage.InputTokens,
		TotalTokens:      res.Usage.InputTokens + res.Usage.OutputTokens,
	}

	if len(res.Content) == 0 {
		return nil, errors.New("anthropic returned no content")
	}

	r.Content = res.Content[0].Text
	return &r, nil
}
