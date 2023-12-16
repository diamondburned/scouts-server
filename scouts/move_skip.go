package scouts

import "fmt"

const SkipMoveType MoveType = "skip"

// SkipMove represents a skip move.
type SkipMove struct{}

var _ Move = (*SkipMove)(nil)

// Type returns the move type.
func (m *SkipMove) Type() MoveType {
	return SkipMoveType
}

func (m *SkipMove) String() string {
	v, _ := m.MarshalText()
	return string(v)
}

func (m *SkipMove) MarshalText() ([]byte, error) {
	return []byte(m.Type()), nil
}

func (m *SkipMove) UnmarshalText(text []byte) error {
	if string(text) != string(SkipMoveType) {
		return fmt.Errorf("expected %q move, got %q", SkipMoveType, text)
	}
	return nil
}

func (m *SkipMove) validate(game *Game) error {
	if !game.currentTurn.hasEnoughPlays(1) {
		return errNotEnoughPlays
	}

	if game.currentState != gameStatePlay {
		return errStillPlacingScouts
	}

	return nil
}

func (m *SkipMove) apply(game *Game) {
	game.addMove(m, 1)
}
