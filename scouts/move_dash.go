package scouts

import (
	"errors"
	"fmt"
	"image"
	"strings"
)

var errDashTooFar = errors.New("cannot dash more than 1 unit at a time")

const DashMoveType MoveType = "dash"

// DashMove represents a dash move. A player may dash a scout in any direction
// in 1 unit increments. The scout may not dash off the board.
type DashMove struct {
	ScoutPosition Point `json:"scout_position"`
	Destination   Point `json:"destination"`
}

var _ Move = (*DashMove)(nil)

// Type returns the move type.
func (m *DashMove) Type() MoveType {
	return DashMoveType
}

func (m *DashMove) String() string {
	v, _ := m.MarshalText()
	return string(v)
}

func (m *DashMove) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf(
		"%s %d,%d %d,%d",
		m.Type(),
		m.ScoutPosition.X, m.ScoutPosition.Y,
		m.Destination.X, m.Destination.Y,
	)), nil
}

func (m *DashMove) UnmarshalText(text []byte) error {
	parts := strings.Split(string(text), " ")
	if len(parts) != 3 {
		return fmt.Errorf("expected 3 parts, got %d", len(parts))
	}

	if parts[0] != string(DashMoveType) {
		return fmt.Errorf("expected %q move, got %q", DashMoveType, parts[0])
	}

	if err := m.ScoutPosition.UnmarshalText([]byte(parts[1])); err != nil {
		return fmt.Errorf("failed to unmarshal scout position: %w", err)
	}

	if err := m.Destination.UnmarshalText([]byte(parts[2])); err != nil {
		return fmt.Errorf("failed to unmarshal destination: %w", err)
	}

	return nil
}

func (m *DashMove) validate(game *Game) error {
	if game.currentState != gameStatePlay {
		return fmt.Errorf("cannot dash piece: %w", UnexpectedGameStateError{
			Expected: gameStatePlay,
			Actual:   game.currentState,
		})
	}

	if !game.currentTurn.hasEnoughPlays(1) {
		return errNotEnoughPlays
	}

	if !game.board.PointIsPiece(m.ScoutPosition, ScoutPieceKind) {
		return errNotYourScout
	}

	if !game.board.PointIsPlayer(m.ScoutPosition, game.currentTurn.Player) {
		return errNotYourScout
	}

	if !game.board.PointIsPiece(m.Destination, NoPieceKind) {
		return fmt.Errorf("cannot dash piece: %w", UnexpectedPieceError{
			Position: m.Destination,
			Expected: NoPieceKind,
			Actual:   game.board.PieceKindAt(m.Destination),
		})
	}

	dashingDistance := Point(image.Rectangle{
		Min: image.Point(m.ScoutPosition),
		Max: image.Point(m.Destination),
	}.Size())
	if abs(dashingDistance.X) > 1 || abs(dashingDistance.Y) > 1 {
		return errDashTooFar
	}

	return nil
}

func (m *DashMove) apply(game *Game) {
	scoutPiece := game.board.PieceAt(m.ScoutPosition).(*ScoutPiece)
	scoutPiece.position = m.Destination

	if !scoutPiece.returning && IsPlayerBase(game.currentTurn.Player.Opponent(), m.Destination) {
		scoutPiece.returning = true
	}

	game.board.updatePiece(scoutPiece)

	if scoutPiece.winsGame() {
		switch game.currentTurn.Player {
		case PlayerA:
			game.currentState = gameStateEndP1Won
		case PlayerB:
			game.currentState = gameStateEndP2Won
		}
	}

	game.addMove(m, 1)
}
