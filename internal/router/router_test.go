package router

import (
	"context"
	"errors"
	"testing"

	"github.com/pranavnallari/glide/internal/adapter"
	"github.com/pranavnallari/glide/internal/config"
	"github.com/pranavnallari/glide/internal/models"
)

// mockProvider lets us control what a provider returns in tests
type mockProvider struct {
	name string
	resp *models.UnifiedResponse
	err  error
}

func (m *mockProvider) Call(_ context.Context, _ *models.UnifiedRequest) (*models.UnifiedResponse, error) {
	return m.resp, m.err
}

func (m *mockProvider) Name() string { return m.name }

func newTestRouter(strategy string, providers []config.RouteTarget, mocks map[string]*mockProvider) *Router {
	providerMap := make(map[string]adapter.Provider) // won't compile yet — see note below
	for k, v := range mocks {
		providerMap[k] = v
	}
	routes := map[string]config.RouteConfig{
		"default": {
			Strategy: strategy,
			Order:    providers,
		},
	}
	return NewRouter(routes, providerMap, nil, nil)
}

func TestFallbackUsesFirstProvider(t *testing.T) {
	mocks := map[string]*mockProvider{
		"openai": {name: "openai", resp: &models.UnifiedResponse{Provider: "openai", Content: "hello"}, err: nil},
	}
	order := []config.RouteTarget{{Provider: "openai", Model: "gpt-4o"}}
	r := newTestRouter("fallback", order, mocks)

	res, err := r.Route(context.Background(), &models.UnifiedRequest{
		Messages: []models.Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Provider != "openai" {
		t.Fatalf("expected openai, got %s", res.Provider)
	}
}

func TestFallbackMovesToNextOnError(t *testing.T) {
	mocks := map[string]*mockProvider{
		"openai":    {name: "openai", err: errors.New("openai down")},
		"anthropic": {name: "anthropic", resp: &models.UnifiedResponse{Provider: "anthropic", Content: "hello"}, err: nil},
	}
	order := []config.RouteTarget{
		{Provider: "openai", Model: "gpt-4o"},
		{Provider: "anthropic", Model: "claude-sonnet-4-6"},
	}
	r := newTestRouter("fallback", order, mocks)

	res, err := r.Route(context.Background(), &models.UnifiedRequest{
		Messages: []models.Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Provider != "anthropic" {
		t.Fatalf("expected fallback to anthropic, got %s", res.Provider)
	}
}

func TestFallbackReturnsErrorWhenAllFail(t *testing.T) {
	mocks := map[string]*mockProvider{
		"openai": {name: "openai", err: errors.New("openai down")},
	}
	order := []config.RouteTarget{{Provider: "openai", Model: "gpt-4o"}}
	r := newTestRouter("fallback", order, mocks)

	_, err := r.Route(context.Background(), &models.UnifiedRequest{
		Messages: []models.Message{{Role: "user", Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error when all providers fail")
	}
}

func TestLoadBalanceRoundRobins(t *testing.T) {
	mocks := map[string]*mockProvider{
		"openai":    {name: "openai", resp: &models.UnifiedResponse{Provider: "openai"}, err: nil},
		"anthropic": {name: "anthropic", resp: &models.UnifiedResponse{Provider: "anthropic"}, err: nil},
	}
	pool := []config.RouteTarget{
		{Provider: "openai", Model: "gpt-4o-mini"},
		{Provider: "anthropic", Model: "claude-sonnet-4-6"},
	}
	routes := map[string]config.RouteConfig{
		"default": {Strategy: "load_balance", Pool: pool},
	}
	providerMap := make(map[string]adapter.Provider)
	for k, v := range mocks {
		providerMap[k] = v
	}
	r := NewRouter(routes, providerMap, nil, nil)

	req := &models.UnifiedRequest{Messages: []models.Message{{Role: "user", Content: "hi"}}}
	res1, _ := r.Route(context.Background(), req)
	res2, _ := r.Route(context.Background(), req)

	if res1.Provider == res2.Provider {
		t.Fatal("expected round robin to alternate providers")
	}
}
