package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"libdb.so/hrt"
	"libdb.so/scouts-server/api/gameserver"
	"libdb.so/scouts-server/api/user"
	"libdb.so/scouts-server/internal/context"
	"libdb.so/scouts-server/internal/unmarshal"
	"libdb.so/scouts-server/scouts"
)

type gameServices struct {
	*gameserver.GameManager
}

type gameHandler struct {
	service gameServices
}

func mountGameHandler(r *chi.Mux, service gameServices) {
	h := &gameHandler{service: service}
	r.Route("/game", func(r chi.Router) {
		r.Post("/", hrt.Wrap(h.createGame))
	})
	r.Route("/game/{id}", func(r chi.Router) {
		r.Use(h.authorizeGame)
		r.Get("/", hrt.Wrap(h.gameInfo))
		r.Post("/join", hrt.Wrap(h.joinGame))
		r.Post("/move", hrt.Wrap(h.makeMove))
		r.Get("/subscribe", h.subscribeGame)
	})
}

func (h *gameHandler) authorizeGame(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idstr := chi.URLParam(r, "id")
		if idstr == "" {
			err := errors.New("missing game ID")
			errorWriter.WriteError(w, hrt.WrapHTTPError(http.StatusBadRequest, err))
			return
		}

		id, err := unmarshal.Text[*gameserver.GameID](idstr)
		if err != nil {
			errorWriter.WriteError(w, hrt.WrapHTTPError(http.StatusBadRequest, err))
			return
		}

		ctx := context.With(r.Context(), *id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type createGameRequest struct {
	TimeLimit gameserver.Duration `json:"time_limit"`
	Increment gameserver.Duration `json:"increment"`
}

type createGameResponse struct {
	GameID gameserver.GameID `json:"game_id"`
}

func (h *gameHandler) createGame(ctx context.Context, req createGameRequest) (createGameResponse, error) {
	authorization := context.From[user.Authorization](ctx)

	id, err := h.service.CreateGame(authorization, gameserver.CreateGameOptions{
		TimeLimit: req.TimeLimit,
		Increment: req.Increment,
	})
	if err != nil {
		return createGameResponse{}, err
	}

	return createGameResponse{GameID: id}, nil
}

func (h *gameHandler) gameInfo(ctx context.Context, _ hrt.None) (gameserver.GameState, error) {
	gameID := context.From[gameserver.GameID](ctx)
	return h.service.QueryGame(gameID)
}

func (h *gameHandler) joinGame(ctx context.Context, _ hrt.None) (hrt.None, error) {
	gameID := context.From[gameserver.GameID](ctx)
	authorization := context.From[user.Authorization](ctx)
	return hrt.Empty, h.service.JoinGame(authorization, gameID)
}

type makeMoveRequest struct {
	Move string `json:"move"`
}

func (h *gameHandler) makeMove(ctx context.Context, req makeMoveRequest) (hrt.None, error) {
	gameID := context.From[gameserver.GameID](ctx)
	authorization := context.From[user.Authorization](ctx)

	move, err := scouts.ParseMove(req.Move)
	if err != nil {
		return hrt.Empty, err
	}

	return hrt.Empty, h.service.MakeMove(authorization, gameID, move)
}

var errNoFlusher = hrt.NewHTTPError(400, "client does not support Server-Sent Events")

func (h *gameHandler) subscribeGame(w http.ResponseWriter, r *http.Request) {
	if _, ok := w.(http.Flusher); !ok {
		errorWriter.WriteError(w, errNoFlusher)
		return
	}

	gameID := context.From[gameserver.GameID](r.Context())
	authorization := context.From[user.Authorization](r.Context())

	events, stop, err := h.service.SubscribeGame(authorization, gameID)
	if err != nil {
		errorWriter.WriteError(w, err)
		return
	}
	defer stop()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	for event := range events {
		writeGameEvent(w, event)
	}
}

// writeGameEvent writes a game event as a Server-Sent Event.
func writeGameEvent(w http.ResponseWriter, event gameserver.GameEvent) {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(w, "event: %s\n", event.Type())
	fmt.Fprintf(w, "data: %s\n\n", string(eventJSON))

	flusher, ok := w.(http.Flusher)
	if !ok {
		panic("http.ResponseWriter does not implement http.Flusher")
	}
	flusher.Flush()
}
