package gameserver

import (
	"context"
	"log/slog"
	"time"

	"github.com/puzpuzpuz/xsync/v3"
	"libdb.so/hrt"
	"libdb.so/scouts-server/api/user"
	"libdb.so/scouts-server/scouts"
)

// ErrNotFound is an error that is returned when an item is not found.
var ErrNotFound = hrt.NewHTTPError(404, "not found")

// ErrGameFull is an error that is returned when a game is full.
var ErrGameFull = hrt.NewHTTPError(400, "game already has two players")

// ErrInvalidMove is an error that is returned when a move is invalid.
var ErrInvalidMove = hrt.NewHTTPError(400, "invalid move")

// CreateGameOptions is a struct that contains options for creating a game.
// All fields are optional.
type CreateGameOptions struct {
	// TimeLimit is the time limit per side.
	// If this is zero, then there is no time limit.
	TimeLimit Duration
	// Increment is the time increment per move.
	Increment Duration
}

// GameState is a struct that contains metadata about a game.
type GameState struct {
	// GameID is the GameID of the game.
	GameID GameID `json:"game_id"`
	// BeganAt is the time that the game began.
	// If nil, then the game has not begun yet.
	BeganAt *time.Time `json:"began_at"`
	// PlayerA is the first player.
	// If nil, then the player has not joined yet.
	PlayerA *user.Authorization `json:"player_a"`
	// PlayerB is the second player.
	// If nil, then the player has not joined yet.
	PlayerB *user.Authorization `json:"player_b"`
	// Moves is the list of moves that have been made in the game.
	Moves []MoveSnapshot `json:"moves"`
	// Metadata is the metadata of the game.
	Metadata CreateGameOptions `json:"metadata"`
	// CreatedAt is the time that the game was created.
	CreatedAt time.Time `json:"created_at"`
	// SnapshotAt is the time that the snapshot was taken.
	SnapshotAt time.Time `json:"snapshot_at"`
}

func (s GameState) hasBothPlayers() bool {
	return s.PlayerA != nil && s.PlayerB != nil
}

// MoveSnapshot contains a single move that a player made and the time that the
// move was made.
type MoveSnapshot struct {
	Player scouts.Player
	Move   scouts.Move
	Time   time.Time
}

const (
	gameTTL = 2 * time.Hour
	gameGC  = 1 * time.Hour
)

// GameManager is in charge of managing games. Games managed here may or may not
// be persisted in a database.
type GameManager struct {
	games  *xsync.MapOf[GameID, *gameInstance]
	logger *slog.Logger
}

// NewGameManager creates a new game manager.
func NewGameManager(logger *slog.Logger) *GameManager {
	return &GameManager{
		games:  xsync.NewMapOf[GameID, *gameInstance](),
		logger: logger.With("component", "api/gameserver/gamemanager"),
	}
}

// BeginGC starts a background goroutine that will periodically garbage collect
// games that have been inactive for a certain amount of time.
func (m *GameManager) BeginGC() (stop func()) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(gameGC)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.gc()
			}
		}
	}()
	return cancel
}

func (m *GameManager) gc() {
	m.games.Range(func(id GameID, game *gameInstance) bool {
		if game.KillIfInactive(gameTTL) {
			m.games.Delete(id)
			m.logger.Info(
				"game has been garbage collected",
				"game_id", id,
				"game_began_at", game.state.BeganAt,
				"game_moves", len(game.state.Moves))
		}
		return true
	})
}

// QueryGame queries the game with the given game ID.
func (m *GameManager) QueryGame(id GameID) (GameState, error) {
	game, ok := m.games.Load(id)
	if !ok {
		return GameState{}, ErrNotFound
	}
	return game.state, nil
}

// CreateGame creates a game with the given game ID and game metadata.
func (m *GameManager) CreateGame(user user.Authorization, metadata CreateGameOptions) (GameID, error) {
	game := newGameInstance(metadata, m.logger, nil)
	for {
		game.state.GameID = GenerateGameID()
		_, exists := m.games.LoadOrStore(game.state.GameID, game)
		if !exists {
			break
		}
	}
	game.logger = game.logger.With("game_id", game.state.GameID)
	return game.state.GameID, nil
}

// JoinGame joins the game with the given game ID.
// If the game is full, then this returns an error, otherwise the player
// automatically takes the next available side.
func (m *GameManager) JoinGame(user user.Authorization, id GameID) (<-chan GameEvent, func(), error) {
	game, ok := m.games.Load(id)
	if !ok {
		return nil, nil, ErrNotFound
	}

	ch, stop, err := game.PlayerJoinNext(user)
	if err != nil {
		return nil, nil, err
	}

	return ch, stop, nil
}

// MakeMove makes a move in the game with the given game ID.
func (m *GameManager) MakeMove(user user.Authorization, id GameID, move scouts.Move) error {
	game, ok := m.games.Load(id)
	if !ok {
		return ErrNotFound
	}
	return game.MakeMove(user, move)
}
