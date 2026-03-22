package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pranavnallari/glide/internal/limiter"
	"github.com/pranavnallari/glide/internal/models"
	"github.com/pranavnallari/glide/internal/observability"
	"github.com/pranavnallari/glide/internal/router"
)

type Handler struct {
	r *router.Router
	l *limiter.Limiter
	m *observability.Metrics
}

func NewHandler(r *router.Router, l *limiter.Limiter, m *observability.Metrics) *Handler {
	return &Handler{r: r, l: l, m: m}
}

func (h *Handler) Chat(w http.ResponseWriter, r *http.Request) {
	keyName := r.Header.Get("Authorization")
	var req models.UnifiedRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "failed to parse request", http.StatusBadRequest)
		return
	}

	if len(req.Messages) == 0 {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	if !h.l.Allow("", keyName) {
		h.m.RequestsTotal.WithLabelValues("unknown", "rate_limited").Inc()
		http.Error(w, "rate_limited", http.StatusTooManyRequests)
		return
	}

	start := time.Now()

	res, err := h.r.Route(r.Context(), &req)
	if err != nil {
		h.m.RequestsTotal.WithLabelValues("unknown", "error").Inc()
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if res == nil {
		http.Error(w, "no response from provider", http.StatusInternalServerError)
		return
	}
	elapsed := time.Since(start)
	h.m.RequestsTotal.WithLabelValues(res.Provider, "success").Inc()
	h.m.RequestDuration.WithLabelValues(res.Provider).Observe(elapsed.Seconds())
	h.m.TokensTotal.WithLabelValues(res.Provider, "input").Add(float64(res.Usage.PromptTokens))
	h.m.TokensTotal.WithLabelValues(res.Provider, "output").Add(float64(res.Usage.CompletionTokens))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
