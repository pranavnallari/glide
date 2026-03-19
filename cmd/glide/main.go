package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pranavnallari/glide/internal/adapter"
	"github.com/pranavnallari/glide/internal/api"
	"github.com/pranavnallari/glide/internal/config"
	"github.com/pranavnallari/glide/internal/router"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	defer stop()

	conf, err := config.Load("config.yaml")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	providers := make(map[string]adapter.Provider)
	for name, p := range conf.Providers {
		if p.Enabled {
			switch name {
			case "openai":
				providers[name] = adapter.NewOpenAIProvider(p)
			case "anthropic":
				providers[name] = adapter.NewAnthropicProvider(p)
			default:
				slog.Warn("invalid provider name", "received", name)
			}
		}
	}

	if len(providers) == 0 {
		slog.Error("no providers enabled")
		os.Exit(1)
	}

	llmRouter := router.NewRouter(conf.Routing, providers, conf.Providers)

	handler := api.NewHandler(llmRouter)

	httpRouter := api.NewRouter(handler)

	srv := &http.Server{Addr: ":" + conf.Server.Port, Handler: httpRouter}
	slog.Info("server starting", "port", conf.Server.Port)

	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	slog.Info("shutdown signal received")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	} else {
		slog.Info("server stopped cleanly")
	}

}
