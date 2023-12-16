package scouts

import (
	"encoding/json"
	"image"
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
	Position() []image.Point
}

// ScoutPiece is a type that represents a scout piece on the board.
type ScoutPiece struct {
	player    Player
	position  image.Point
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
func (p *ScoutPiece) Position() []image.Point {
	return []image.Point{p.position}
}

// Returning returns whether the scout is returning to its base.
func (p *ScoutPiece) Returning() bool {
	return p.returning
}

// MarshalJSON marshals the scout piece to JSON.
func (p *ScoutPiece) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind      PieceKind   `json:"kind"`
		Player    Player      `json:"player"`
		Position  image.Point `json:"position"`
		Returning bool        `json:"returning"`
	}{
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
	position [4]image.Point
	player   Player
}

func boulderPiecePosition(topLeft image.Point) [4]image.Point {
	return [4]image.Point{
		topLeft,
		topLeft.Add(image.Point{1, 0}),
		topLeft.Add(image.Point{0, 1}),
		topLeft.Add(image.Point{1, 1}),
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
func (p *BoulderPiece) Position() []image.Point {
	return p.position[:]
}

// MarshalJSON marshals the boulder piece to JSON.
func (p *BoulderPiece) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind     PieceKind      `json:"kind"`
		Player   Player         `json:"player"`
		Position [4]image.Point `json:"position"`
	}{
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
