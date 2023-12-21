package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"libdb.so/hrt"
	"libdb.so/scouts-server/api/gameserver"
	"libdb.so/scouts-server/api/user"
	"libdb.so/scouts-server/internal/context"
	"libdb.so/scouts-server/internal/unmarshal"
)

// Services defines the services that the API handler needs to function.
type Services struct {
	*gameserver.GameManager
	user.SessionStorage
}

var errorWriter = hrt.JSONErrorWriter("error")

// Handler is the main API handler.
type Handler struct {
	*chi.Mux
	service Services
}

// NewHandler creates a new API handler.
func NewHandler(service Services) *Handler {
	h := &Handler{service: service}

	h.Mux = chi.NewRouter()
	h.With(
		middleware.CleanPath,
		middleware.RealIP,
		middleware.Recoverer,
		h.authorize,
		hrt.Use(hrt.Opts{
			Encoder:     hrt.DefaultEncoder,
			ErrorWriter: hrt.JSONErrorWriter("error"),
		}),
	)

	mountGameHandler(h.Mux, gameServices{
		GameManager: service.GameManager,
	})

	return h
}

func (h *Handler) authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token user.SessionToken
		if cookie, err := r.Cookie("session"); err == nil {
			t, err := unmarshal.Text[*user.SessionToken](cookie.Value)
			if err != nil {
				errorWriter.WriteError(w, err)
				return
			}
			token = *t
		} else {
			t, err := h.service.CreateSession()
			if err != nil {
				errorWriter.WriteError(w, err)
				return
			}
			http.SetCookie(w, &http.Cookie{
				Name:     "session",
				Value:    t.String(),
				MaxAge:   int(user.SessionTTL / time.Second),
				SameSite: http.SameSiteStrictMode,
			})
			token = t
		}

		userID, err := h.service.QuerySession(token)
		if err != nil {
			errorWriter.WriteError(w, err)
			return
		}

		var authorization user.Authorization
		if userID != nil {
			authorization = user.NewAuthorized(token, *userID)
		} else {
			authorization = user.NewAnonymous(token)
		}

		ctx := context.With(r.Context(), authorization)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
