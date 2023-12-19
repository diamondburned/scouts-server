package scouts

import (
	"fmt"
	"image"
	"strings"

	"golang.org/x/exp/constraints"
)

const JumpMoveType MoveType = "jump"

// JumpMove represents a jump move.
// A jump move describes a piece jumping over another piece in one of the four
// cardinal directions.
// A jump move costs 0 plays, but later jumps must be made on the same scout.
type JumpMove struct {
	ScoutPosition Point `json:"scout_position"`
	Destination   Point `json:"destination"`
}

var _ Move = (*JumpMove)(nil)

// Type returns the move type.
func (m *JumpMove) Type() MoveType {
	return JumpMoveType
}

func (m *JumpMove) String() string {
	v, _ := m.MarshalText()
	return string(v)
}

func (m *JumpMove) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf(
		"%s %d,%d %d,%d",
		m.Type(),
		m.ScoutPosition.X, m.ScoutPosition.Y,
		m.Destination.X, m.Destination.Y,
	)), nil
}

func (m *JumpMove) UnmarshalText(text []byte) error {
	parts := strings.Split(string(text), " ")
	if len(parts) != 3 {
		return fmt.Errorf("expected 3 parts, got %d", len(parts))
	}

	if parts[0] != string(JumpMoveType) {
		return fmt.Errorf("expected %q move, got %q", JumpMoveType, parts[0])
	}

	if err := m.ScoutPosition.UnmarshalText([]byte(parts[1])); err != nil {
		return fmt.Errorf("failed to unmarshal scout position: %w", err)
	}

	if err := m.Destination.UnmarshalText([]byte(parts[2])); err != nil {
		return fmt.Errorf("failed to unmarshal destination: %w", err)
	}

	return nil
}

func (m *JumpMove) cost(game *Game) int {
	if len(game.currentTurn.Moves) > 0 {
		lastMove := game.currentTurn.Moves[len(game.currentTurn.Moves)-1]
		if jump, ok := lastMove.(*JumpMove); ok && jump.Destination == m.ScoutPosition {
			// This jump is free because we're jumping the same scout as the last
			// jump.
			return 0
		}
	}
	return 1
}

// Apply applies the move to the board.
func (m *JumpMove) validate(game *Game) error {
	if game.currentState != gameStatePlay {
		return errStillPlacingScouts
	}

	// Assert that the player has enough plays.
	if !game.currentTurn.hasEnoughPlays(m.cost(game)) {
		return errNotEnoughPlays
	}

	// Assert that the piece at the scout position is a scout.
	if !game.board.PointIsPiece(m.ScoutPosition, ScoutPieceKind) {
		return errNotYourScout
	}

	// Assert that the piece at the scout position is the player's piece.
	if !game.board.PointIsPlayer(m.ScoutPosition, game.currentTurn.Player) {
		return errNotYourScout
	}

	jumpingDistance := Point(image.Rectangle{
		Min: image.Point(m.ScoutPosition),
		Max: image.Point(m.Destination),
	}.Size())
	// Assert that the jump is in one of the four cardinal directions.
	if jumpingDistance.X != 0 && jumpingDistance.Y != 0 {
		return errInvalidJump
	}
	// Assert that the jump is exactly two spaces.
	if abs(jumpingDistance.X) != 2 && abs(jumpingDistance.Y) != 2 {
		return errInvalidJump
	}

	// Assert that we must jump over a scout.
	jumpingOverPosition := m.ScoutPosition.Add(jumpingDistance.Div(2))
	if !game.board.PointIsPiece(jumpingOverPosition, ScoutPieceKind) {
		return fmt.Errorf("cannot jump over piece: %w", UnexpectedPieceError{
			Position: jumpingOverPosition,
			Expected: ScoutPieceKind,
			Actual:   game.board.PieceKindAt(jumpingOverPosition),
		})
	}

	// Assert that the place that we're jumping to is empty.
	if !game.board.PointIsPiece(m.Destination, NoPieceKind) {
		return fmt.Errorf("cannot jump to occupied space: %w", UnexpectedPieceError{
			Position: m.Destination,
			Expected: NoPieceKind,
			Actual:   game.board.PieceKindAt(m.Destination),
		})
	}

	return nil
}

func (m *JumpMove) apply(game *Game) {
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

	game.addMove(m, m.cost(game))
}

func abs[T constraints.Integer](x T) T {
	if x < 0 {
		return -x
	}
	return x
}
