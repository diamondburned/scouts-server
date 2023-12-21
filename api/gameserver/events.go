package gameserver

import (
	"libdb.so/scouts-server/api/user"
	"libdb.so/scouts-server/scouts"
)

// GameEvent is an interface that represents a game event.
// An event is emitted when the game state changes.
type GameEvent interface {
	// Type returns the type of the game event.
	Type() string
}

// PlayerJoinedEvent is an event that is emitted when a player joins the game.
type PlayerJoinedEvent struct {
	// PlayerSide is the side that the user joined.
	PlayerSide scouts.Player `json:"player_side"`
	// UserID is the ID of the user that joined the game.
	// If this is nil, then the user is anonymous.
	UserID *user.UserID `json:"user_id"`
}

// PlayerLeftEvent is an event that is emitted when a player leaves the game.
type PlayerLeftEvent struct {
	// PlayerSide is the side that the user left.
	PlayerSide scouts.Player `json:"player_side"`
	// UserID is the ID of the user that left the game.
	// If this is nil, then the user is anonymous.
	UserID *user.UserID `json:"user_id"`
}

// PlayerConnectedEvent is an event that is emitted when a player
// connects to the game. It can only be emitted after a PlayerJoinedEvent
// but before a PlayerLeftEvent.
type PlayerConnectedEvent struct {
	// PlayerSide is the side that the user connected.
	PlayerSide scouts.Player `json:"player_side"`
}

// PlayerDisconnectedEvent is an event that is emitted when a player
// disconnects from the game. It can only be emitted after a PlayerJoinedEvent
// but before a PlayerLeftEvent. A disconnect implies that the player might
// still choose to rejoin the game, after which a PlayerConnectedEvent will be
// emitted again.
type PlayerDisconnectedEvent struct {
	// PlayerSide is the side that the user disconnected.
	PlayerSide scouts.Player `json:"player_side"`
}

// TurnBeginEvent is an event that is emitted when a turn begins.
type TurnBeginEvent struct {
	// PlayerSide is the side that is about to make a move.
	PlayerSide scouts.Player `json:"player_side"`
	// PlaysRemaining is the number of plays remaining for the player.
	PlaysRemaining int `json:"plays_remaining"`
	// TimeRemaining is the time remaining for both sides.
	TimeRemaining [2]Duration `json:"time_remaining"`
}

// MoveMadeEvent is an event that is emitted when a move is made.
type MoveMadeEvent struct {
	// Move is the move that was made.
	Move scouts.Move `json:"move"`
	// PlayerSide is the side that made the move.
	PlayerSide scouts.Player `json:"player_side"`
	// PlaysRemaining is the number of plays remaining for the player.
	PlaysRemaining int `json:"plays_remaining"`
	// TimeRemaining is the time remaining for both sides.
	TimeRemaining [2]Duration `json:"time_remaining"`
}

// GameEndEvent is an event that is emitted when the game ends.
type GameEndEvent struct {
	// Winner is the side that won the game.
	Winner scouts.Player `json:"winner"`
	// TimeRemaining is the time remaining for both sides.
	TimeRemaining [2]Duration `json:"time_remaining"`
}

// GoingAwayEvent is an event that is emitted when the server is about to
// disconnect the client.
type GoingAwayEvent struct {
	// Reason string `json:"reason"`
}

// var (
// 	// GoingAwayGameEnd is an event that is emitted when the server is about to
// 	// disconnect the client because the game has ended.
// 	GoingAwayGameEnd = GoingAwayEvent{Reason: "game has ended"}
// 	// GoingAwayPlayerLeft is an event that is emitted when the server is about
// 	// to disconnect the client because the player left the game.
// 	GoingAwayPlayerLeft = GoingAwayEvent{Reason: "player left the game"}
// 	// GoingAwayGameExpired is an event that is emitted when the server is about
// 	// to disconnect the client because the game has expired.
// 	GoingAwayGameExpired = GoingAwayEvent{Reason: "game has expired"}
// )

func (PlayerJoinedEvent) Type() string       { return "player_joined" }
func (PlayerLeftEvent) Type() string         { return "player_left" }
func (PlayerConnectedEvent) Type() string    { return "player_connected" }
func (PlayerDisconnectedEvent) Type() string { return "player_disconnected" }
func (TurnBeginEvent) Type() string          { return "turn_begin" }
func (MoveMadeEvent) Type() string           { return "move_made" }
func (GameEndEvent) Type() string            { return "game_end" }
func (GoingAwayEvent) Type() string          { return "going_away" }
