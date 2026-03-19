package api

import (
	"encoding/json"
	"net/http"

	"github.com/pranavnallari/glide/internal/models"
	"github.com/pranavnallari/glide/internal/router"
)

type Handler struct {
	r *router.Router
}

func NewHandler(r *router.Router) *Handler {
	return &Handler{r: r}
}

func (h *Handler) Chat(w http.ResponseWriter, r *http.Request) {
	var req models.UnifiedRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "failed to parse request", http.StatusBadRequest)
		return
	}

	if len(req.Messages) == 0 {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}
	res, err := h.r.Route(r.Context(), &req)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if res == nil {
		http.Error(w, "no response from provider", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
