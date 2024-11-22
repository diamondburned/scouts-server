package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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

	h.Use(
		middleware.Logger,
		cors.Handler(cors.Options{
			AllowedOrigins:   []string{"https://*", "http://*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
		}),
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
			// session cookie is set
			t, err := unmarshal.Text[*user.SessionToken](cookie.Value)
			if err != nil {
				errorWriter.WriteError(w, err)
				return
			}
			token = *t
		} else {
			// session cookie is not set, so we create one
			t, err := h.service.CreateSession()
			if err != nil {
				errorWriter.WriteError(w, err)
				return
			}
			tokenBytes, err := t.MarshalText()
			if err != nil {
				errorWriter.WriteError(w, err)
				return
			}
			http.SetCookie(w, &http.Cookie{
				Name:     "session",
				Value:    string(tokenBytes),
				MaxAge:   int(user.SessionTTL / time.Second),
				SameSite: http.SameSiteNoneMode,
				Secure:   true,
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
