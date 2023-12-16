package scouts

import (
	"encoding"
	"fmt"
	"image"
	"strings"
)

var (
	errStillPlacingScouts = fmt.Errorf("you are still placing scouts")
	errHasPlacedAllScouts = fmt.Errorf("you have placed all scouts")
	errHasNoPlays         = fmt.Errorf("you have no plays")
	errNotEnoughPlays     = fmt.Errorf("you do not have enough plays")
	errCanOnlyPlaceAtBase = fmt.Errorf("you can only place a piece at your base")
	errNotYourScout       = fmt.Errorf("you cannot move a scout that is not yours")
	errOutOfBounds        = fmt.Errorf("cannot go off the board")
	errInvalidJump        = fmt.Errorf("invalid jump")
)

// UnexpectedPieceError is an error that is returned when a piece is not the
// expected piece.
type UnexpectedPieceError struct {
	Position image.Point `json:"position"`
	Expected PieceKind   `json:"expected"`
	Actual   PieceKind   `json:"actual"`
}

func (e UnexpectedPieceError) Error() string {
	return fmt.Sprintf(
		"expected %s at %s, got %s",
		e.Expected, e.Position, e.Actual,
	)
}

// UnexpectedGameStateError is an error that is returned when the game state is
// not the expected game state.
type UnexpectedGameStateError struct {
	Expected gameState `json:"expected"`
	Actual   gameState `json:"actual"`
}

func (e UnexpectedGameStateError) Error() string {
	return fmt.Sprintf(
		"expected game state %s, got %s",
		e.Expected, e.Actual,
	)
}

func parsePoint(str string) (image.Point, error) {
	var p image.Point
	_, err := fmt.Sscanf(str, "%d,%d", &p.X, &p.Y)
	return p, err
}

// MoveType is a type that represents a move type.
type MoveType string

// Move is a type that represents a move.
type Move interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
	fmt.Stringer

	// Type returns the move type.
	Type() MoveType

	validate(*Game) error
	apply(*Game)
}

// ParseMove parses a move from a string.
func ParseMove(s string) (Move, error) {
	arg0, _, _ := strings.Cut(s, " ")
	var move Move
	switch MoveType(arg0) {
	case PlaceScoutMoveType:
		move = &PlaceScoutMove{}
	case JumpMoveType:
		move = &JumpMove{}
	case DashMoveType:
		move = &DashMove{}
	case SkipMoveType:
		move = &SkipMove{}
	case BoulderMoveType:
		move = &BoulderMove{}
	default:
		return nil, fmt.Errorf("unknown move type: %q", arg0)
	}
	if err := move.UnmarshalText([]byte(s)); err != nil {
		return nil, fmt.Errorf("invalid %s move: %v", arg0, err)
	}
	return move, nil
}

func moveIsEq(move1, move2 Move) bool {
	// Tell no one about this.
	return move1.String() == move2.String()
}

// Moves is a list of moves.
type Moves []Move

// ParseMoves parses a list of moves.
func ParseMoves(str string) (Moves, error) {
	var moves Moves
	err := moves.UnmarshalText([]byte(str))
	return moves, err
}

var (
	_ encoding.TextMarshaler   = (*Moves)(nil)
	_ encoding.TextUnmarshaler = (*Moves)(nil)
)

func (m *Moves) MarshalText() ([]byte, error) {
	strs := make([]string, len(*m))
	for i, move := range *m {
		b, err := move.MarshalText()
		if err != nil {
			return nil, err
		}
		strs[i] = string(b)
	}
	return []byte(strings.Join(strs, "; ")), nil
}

func (m *Moves) UnmarshalText(text []byte) error {
	parts := strings.Split(string(text), ";")
	moves := make([]Move, 0, len(parts))
	for _, s := range parts {
		s = strings.TrimSpace(s)
		move, err := ParseMove(s)
		if err != nil {
			return fmt.Errorf("cannot parse move %q: %v", s, err)
		}
		moves = append(moves, move)
	}
	*m = moves
	return nil
}
