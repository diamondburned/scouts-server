package scouts

import (
	"fmt"
	"image"
	"strings"
)

var (
	errAlreadyPlacedBoulder = fmt.Errorf("already placed boulder")
)

const BoulderMoveType MoveType = "boulder"

// BoulderMove represents a move that places a boulder on the board.
// A boulder is a 2x2 piece that cannot be moved.
type BoulderMove struct {
	// TopLeft is the top left position of the boulder.
	TopLeft image.Point `json:"top_left"`
}

var _ Move = (*BoulderMove)(nil)

// Type returns the move type.
func (m *BoulderMove) Type() MoveType {
	return BoulderMoveType
}

func (m *BoulderMove) String() string {
	v, _ := m.MarshalText()
	return string(v)
}

func (m *BoulderMove) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf(
		"%s %d,%d",
		m.Type(),
		m.TopLeft.X, m.TopLeft.Y,
	)), nil
}

func (m *BoulderMove) UnmarshalText(text []byte) error {
	parts := strings.Split(string(text), " ")
	if len(parts) != 2 {
		return fmt.Errorf("expected 2 parts, got %d", len(parts))
	}

	if parts[0] != string(BoulderMoveType) {
		return fmt.Errorf("expected %q move, got %q", BoulderMoveType, parts[0])
	}

	p, err := parsePoint(parts[1])
	if err != nil {
		return err
	}
	m.TopLeft = p

	return nil
}

func (m *BoulderMove) apply(game *Game) error {
	if game.currentState != gameStatePlay {
		return errStillPlacingScouts
	}

	if !game.currentTurn.hasEnoughPlays(1) {
		return errNotEnoughPlays
	}

	position := [4]image.Point{
		m.TopLeft,
		m.TopLeft.Add(image.Point{1, 0}),
		m.TopLeft.Add(image.Point{0, 1}),
		m.TopLeft.Add(image.Point{1, 1}),
	}
	for _, p := range position {
		if !game.board.PointIsPiece(p, NoPieceKind) {
			return fmt.Errorf("boulder cannot be placed over piece at %v", UnexpectedPieceError{
				Position: p,
				Expected: NoPieceKind,
				Actual:   game.board.PieceKindAt(p),
			})
		}
	}

	placedBoulderIx := int(game.currentTurn.Player) - 1
	if game.placedBoulders[placedBoulderIx] {
		return errAlreadyPlacedBoulder
	}

	boulderPiece := &BoulderPiece{
		player:   game.currentTurn.Player,
		position: position,
	}

	game.placedBoulders[placedBoulderIx] = true
	game.board.updatePiece(boulderPiece)
	game.addMove(m, 1)

	return nil
}
