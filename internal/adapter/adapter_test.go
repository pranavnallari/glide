package adapter

import (
	"testing"

	"github.com/pranavnallari/glide/internal/models"
)

func TestToOpenAIFailsOnEmptyMessages(t *testing.T) {
	req := &models.UnifiedRequest{Model: "gpt-4o"}
	_, err := ToOpenAI(req)
	if err == nil {
		t.Fatal("expected error for empty messages")
	}
}

func TestToOpenAISetsDefaultMaxTokens(t *testing.T) {
	req := &models.UnifiedRequest{
		Model:    "gpt-4o",
		Messages: []models.Message{{Role: "user", Content: "hello"}},
	}
	r, err := ToOpenAI(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.MaxCompletionTokens != DefaultMaxTokens {
		t.Fatalf("expected %d, got %d", DefaultMaxTokens, r.MaxCompletionTokens)
	}
}

func TestToAnthropicSeparatesSystemMessages(t *testing.T) {
	req := &models.UnifiedRequest{
		Model: "claude-sonnet-4-6",
		Messages: []models.Message{
			{Role: "system", Content: "you are helpful"},
			{Role: "user", Content: "hello"},
		},
	}
	r, err := ToAnthropic(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.System != "you are helpful" {
		t.Fatalf("expected system message to be extracted, got: %s", r.System)
	}
	if len(r.Messages) != 1 {
		t.Fatalf("expected 1 non-system message, got %d", len(r.Messages))
	}
}

func TestToAnthropicFailsOnInvalidTemperature(t *testing.T) {
	req := &models.UnifiedRequest{
		Model:       "claude-sonnet-4-6",
		Messages:    []models.Message{{Role: "user", Content: "hello"}},
		Temperature: 1.5, // invalid for anthropic, max is 1.0
	}
	_, err := ToAnthropic(req)
	if err == nil {
		t.Fatal("expected error for temperature > 1.0")
	}
}

func TestFromOpenAIFailsOnEmptyChoices(t *testing.T) {
	res := &OpenAIResponse{Choices: []Choice{}}
	_, err := FromOpenAI(res)
	if err == nil {
		t.Fatal("expected error for empty choices")
	}
}

func TestFromAnthropicMapsTokensCorrectly(t *testing.T) {
	res := &AnthropicResponse{
		Id:    "test-id",
		Model: "claude-sonnet-4-6",
		Content: []AnthropicResponseContent{
			{Text: "hello"},
		},
		Usage: AnthropicTokenCount{
			InputTokens:  10,
			OutputTokens: 20,
		},
	}
	unified, err := FromAnthropic(res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if unified.Usage.TotalTokens != 30 {
		t.Fatalf("expected total tokens 30, got %d", unified.Usage.TotalTokens)
	}
	if unified.Usage.PromptTokens != 10 {
		t.Fatalf("expected prompt tokens 10, got %d", unified.Usage.PromptTokens)
	}
}
