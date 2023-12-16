package scouts

import (
	"fmt"
	"image"
	"strings"
)

const DashMoveType MoveType = "dash"

// DashMove represents a dash move. A player may dash a scout in any direction
// in 1 unit increments. The scout may not dash off the board.
type DashMove struct {
	ScoutPosition image.Point `json:"scout_position"`
	Destination   image.Point `json:"destination"`
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

	p, err := parsePoint(parts[1])
	if err != nil {
		return err
	}
	m.ScoutPosition = p

	p, err = parsePoint(parts[2])
	if err != nil {
		return err
	}
	m.Destination = p

	return nil
}

func (m *DashMove) apply(game *Game) error {
	if game.currentState != gameStatePlay {
		return errStillPlacingScouts
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

	scoutPiece := game.board.PieceAt(m.ScoutPosition).(*ScoutPiece)
	scoutPiece.position = m.Destination

	if !scoutPiece.returning && game.board.IsPlayerBase(game.currentTurn.Player.Opponent(), m.Destination) {
		scoutPiece.returning = true
	}

	game.board.updatePiece(scoutPiece)
	game.addMove(m, 1)

	return nil
}
