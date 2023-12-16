package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	*chi.Mux
}

func NewHandler() *Handler {
	h := &Handler{}

	h.Mux = chi.NewRouter()
	h.Route("/game/{id}", func(r chi.Router) {
		r.Use(h.authorizeGame)
		r.Get("/events", h.gameEvents)
		r.Post("/move", h.makeMove)
	})
}

func (h *Handler) authorizeGame(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "missing game id", http.StatusBadRequest)
			return
		}

		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

	})
}

func (h *Handler) gameEvents(w http.ResponseWriter, r *http.Request) {
}

func (h *Handler) makeMove(w http.ResponseWriter, r *http.Request) {
}
