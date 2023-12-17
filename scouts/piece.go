package scouts

import (
	"encoding/json"
)

// Piece is a type that represents a piece on the board.
// A piece must be able to be marshaled to JSON. It must not be unmarshaled.
type Piece interface {
	json.Marshaler

	// Kind returns the kind of the piece.
	Kind() PieceKind
	// Player returns the player that owns the piece.
	Player() Player
	// Position returns the position of the piece, which may be multiple points.
	Position() []Point
}

// ScoutPiece is a type that represents a scout piece on the board.
type ScoutPiece struct {
	player    Player
	position  Point
	returning bool
}

// Kind returns the kind of the piece.
func (p *ScoutPiece) Kind() PieceKind {
	return ScoutPieceKind
}

// Player returns the player that owns the piece.
func (p *ScoutPiece) Player() Player {
	return p.player
}

// Position returns the position of the piece, which may be multiple points.
func (p *ScoutPiece) Position() []Point {
	return []Point{p.position}
}

// Returning returns whether the scout is returning to its base.
func (p *ScoutPiece) Returning() bool {
	return p.returning
}

// ScoutsPieceModel is the JSON model for a scout piece.
type ScoutsPieceModel struct {
	Kind      PieceKind `json:"kind"`
	Player    Player    `json:"player"`
	Position  Point     `json:"position"`
	Returning bool      `json:"returning"`
}

// MarshalJSON marshals the scout piece to JSON.
func (p *ScoutPiece) MarshalJSON() ([]byte, error) {
	return json.Marshal(ScoutsPieceModel{
		Kind:      p.Kind(),
		Player:    p.player,
		Position:  p.position,
		Returning: p.returning,
	})
}

func (p *ScoutPiece) winsGame() bool {
	return p.returning && IsPlayerBase(p.player, p.position)
}

// BoulderPiece is a type that represents a boulder piece on the board.
type BoulderPiece struct {
	position [4]Point
	player   Player
}

func boulderPiecePosition(topLeft Point) [4]Point {
	return [4]Point{
		topLeft,
		topLeft.Add(Point{1, 0}),
		topLeft.Add(Point{0, 1}),
		topLeft.Add(Point{1, 1}),
	}
}

// Kind returns the kind of the piece.
func (p *BoulderPiece) Kind() PieceKind {
	return BoulderPieceKind
}

// Player returns the player that owns the piece.
func (p *BoulderPiece) Player() Player {
	return p.player
}

// Position returns the position of the piece, which may be multiple points.
func (p *BoulderPiece) Position() []Point {
	return p.position[:]
}

// BoulderPieceModel is the JSON model for a boulder piece.
type BoulderPieceModel struct {
	Kind     PieceKind `json:"kind"`
	Player   Player    `json:"player"`
	Position [4]Point  `json:"position"`
}

// MarshalJSON marshals the boulder piece to JSON.
func (p *BoulderPiece) MarshalJSON() ([]byte, error) {
	return json.Marshal(BoulderPieceModel{
		Kind:     p.Kind(),
		Player:   p.player,
		Position: p.position,
	})
}

// PieceKind is a type that can either be ScoutPiece or BoulderPiece.
type PieceKind string

const (
	NoPieceKind      PieceKind = ""
	ScoutPieceKind   PieceKind = "scout"
	BoulderPieceKind PieceKind = "boulder"
)

func (k PieceKind) String() string {
	if k == NoPieceKind {
		return "no piece"
	}
	return string(k)
}
