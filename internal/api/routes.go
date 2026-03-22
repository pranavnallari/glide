package api

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /chat", h.Chat)
	mux.Handle("GET /metrics", promhttp.Handler())
	mux.HandleFunc("GET /health", h.Health)
	return mux
}
