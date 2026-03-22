package router

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/pranavnallari/glide/internal/adapter"
	"github.com/pranavnallari/glide/internal/config"
	"github.com/pranavnallari/glide/internal/limiter"
	"github.com/pranavnallari/glide/internal/models"
)

type Router struct {
	sync.Mutex
	routes         map[string]config.RouteConfig
	providers      map[string]adapter.Provider
	providerConfig map[string]config.ProviderConfig
	ind            map[string]int
	limiter        *limiter.Limiter
}

func NewRouter(routes map[string]config.RouteConfig, providers map[string]adapter.Provider, provCfg map[string]config.ProviderConfig, l *limiter.Limiter) *Router {
	return &Router{
		routes:         routes,
		providers:      providers,
		providerConfig: provCfg,
		ind:            make(map[string]int),
		limiter:        l,
	}
}

func (r *Router) Route(ctx context.Context, req *models.UnifiedRequest) (*models.UnifiedResponse, error) {
	routeCfg, ok := r.routes[req.Strategy]
	if !ok {
		routeCfg = r.routes["default"]
	}

	strategy := routeCfg.Strategy
	switch strategy {
	case "fallback":
		return r.routeFallback(ctx, req, routeCfg)
	case "priority":
		return r.routePriority(ctx, req, routeCfg)
	case "load_balance":
		return r.routeLoadBalance(ctx, strategy, req, routeCfg)
	default:
		return nil, errors.New("unknown strategy")
	}
}

func (r *Router) routeFallback(ctx context.Context, req *models.UnifiedRequest, route config.RouteConfig) (*models.UnifiedResponse, error) {
	for i := range len(route.Order) {
		req.Model = route.Order[i].Model
		provider := r.providers[route.Order[i].Provider]

		if r.limiter != nil && !r.limiter.Allow(provider.Name(), "") {
			slog.Warn("provider rate limited", "provider", provider.Name())
			continue
		}

		res, err := provider.Call(ctx, req)
		if err != nil {
			slog.Warn("provider failed", "provider", route.Order[i].Provider, "error", err)
			continue
		}

		return res, nil
	}

	return nil, errors.New("all providers failed")
}

func (r *Router) routePriority(ctx context.Context, req *models.UnifiedRequest, route config.RouteConfig) (*models.UnifiedResponse, error) {
	for i := range len(route.Order) {
		provider := r.providers[route.Order[i].Provider]

		if !r.providerConfig[provider.Name()].Enabled {
			continue
		}
		req.Model = route.Order[i].Model

		if r.limiter != nil && !r.limiter.Allow(provider.Name(), "") {
			slog.Warn("provider rate limited", "provider", provider.Name())
			continue
		}

		res, err := provider.Call(ctx, req)
		if err != nil {
			slog.Warn("provider failed", "provider", route.Order[i].Provider, "error", err)
			continue
		}

		return res, nil
	}
	return nil, errors.New("no healthy providers")
}

func (r *Router) routeLoadBalance(ctx context.Context, routeName string, req *models.UnifiedRequest, route config.RouteConfig) (*models.UnifiedResponse, error) {
	if len(route.Pool) == 0 {
		return nil, errors.New("load balance pool is empty")
	}

	r.Lock()

	currInd := r.ind[routeName] % len(route.Pool)
	req.Model = route.Pool[currInd].Model
	provider := r.providers[route.Pool[currInd].Provider]
	r.ind[routeName] = currInd + 1

	r.Unlock()
	if r.limiter != nil && !r.limiter.Allow(provider.Name(), "") {
		slog.Warn("provider rate limited", "provider", provider.Name())
		return nil, errors.New("provider rate limited")
	}
	return provider.Call(ctx, req)

}
