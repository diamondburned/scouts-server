package scouts

import (
	"fmt"
	"strings"
)

const PlaceScoutMoveType MoveType = "place_scout"

// PlaceScoutMove represents a place scout move.
// Each player must place 5 scouts on their base before the game begins.
type PlaceScoutMove struct {
	ScoutPosition Point `json:"scout_position"`
}

var _ Move = (*PlaceScoutMove)(nil)

// Type returns the move type.
func (m *PlaceScoutMove) Type() MoveType {
	return PlaceScoutMoveType
}

func (m *PlaceScoutMove) String() string {
	v, _ := m.MarshalText()
	return string(v)
}

func (m *PlaceScoutMove) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf(
		"%s %d,%d",
		m.Type(), m.ScoutPosition.X, m.ScoutPosition.Y,
	)), nil
}

func (m *PlaceScoutMove) UnmarshalText(text []byte) error {
	k, v, _ := strings.Cut(string(text), " ")
	if k != string(PlaceScoutMoveType) {
		return fmt.Errorf("expected %q move, got %q", PlaceScoutMoveType, k)
	}
	if err := m.ScoutPosition.UnmarshalText([]byte(v)); err != nil {
		return fmt.Errorf("failed to unmarshal scout position: %w", err)
	}
	return nil
}

func (m *PlaceScoutMove) validate(game *Game) error {
	if !game.currentTurn.hasEnoughPlays(1) {
		return errNotEnoughPlays
	}

	if game.currentState != gameStatePlaceScouts {
		return errHasPlacedAllScouts
	}

	if !IsPlayerBase(game.currentTurn.Player, m.ScoutPosition) {
		return errCanOnlyPlaceAtBase
	}

	if !game.board.PointIsPiece(m.ScoutPosition, NoPieceKind) {
		return fmt.Errorf("cannot place scout: %w", UnexpectedPieceError{
			Expected: NoPieceKind,
			Actual:   game.board.PieceKindAt(m.ScoutPosition),
			Position: m.ScoutPosition,
		})
	}

	return nil
}

func (m *PlaceScoutMove) apply(game *Game) {
	piece := &ScoutPiece{
		player:   game.currentTurn.Player,
		position: m.ScoutPosition,
	}

	game.board.addPiece(piece)
	game.addEndMove(m, 1)
	game.currentTurn.Plays = StartingPlaysPerTurn

	if len(game.turns) == 10 {
		game.currentState = gameStatePlay
	}
}
