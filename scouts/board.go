package scouts

import (
	"image"
	"slices"
)

// BoardBounds is the size of the board.
var BoardBounds = image.Rect(0, 0, 8, 10)

var playerABaseY = BoardBounds.Max.Y - 1
var playerBBaseY = BoardBounds.Min.Y

// IsPlayerBase returns whether the given point is on the base of the given
// player.
func IsPlayerBase(player Player, pt Point) bool {
	switch player {
	case PlayerA:
		return pt.Y == BoardBounds.Max.Y-1
	case PlayerB:
		return pt.Y == BoardBounds.Min.Y
	default:
		panic("invalid player")
	}
}

// Board is a type that represents the board. Player A is at the bottom of the
// board and player B is at the top of the board.
// It exposes no methods for modifying the board, only for reading it.
// To modify the board, use the Apply method on a Move.
type Board struct {
	positions map[Point]Piece
	pieces    map[Piece][]Point
}

// NewBoard returns a new board.
func NewBoard() *Board {
	return &Board{
		positions: make(map[Point]Piece),
		pieces:    make(map[Piece][]Point),
	}
}

// Bounds returns the bounds of the board.
func (b *Board) Bounds() image.Rectangle {
	return BoardBounds
}

// Pieces returns the pieces on the board.
func (b *Board) Pieces() []Piece {
	pieces := make([]Piece, 0, len(b.positions))
	for piece := range b.pieces {
		pieces = append(pieces, piece)
	}
	return pieces
}

// PieceAt returns the piece at the given point, or nil if there is no piece at
// the given point.
func (b *Board) PieceAt(p Point) Piece {
	return b.positions[p]
}

// HasPieceAt returns whether there is a piece at the given point.
func (b *Board) HasPieceAt(p Point) bool {
	return b.PieceAt(p) != nil
}

func (b *Board) PieceKindAt(p Point) PieceKind {
	piece := b.PieceAt(p)
	if piece == nil {
		return NoPieceKind
	}
	return piece.Kind()
}

func (b *Board) PointIsPlayer(p Point, player Player) bool {
	if !p.In(b.Bounds()) {
		return false
	}
	piece := b.PieceAt(p)
	return piece != nil && piece.Player() == player
}

func (b *Board) PointIsPiece(p Point, kind PieceKind) bool {
	if !p.In(b.Bounds()) {
		return false
	}
	piece := b.PieceAt(p)
	if kind == NoPieceKind {
		return piece == nil
	}
	return piece != nil && piece.Kind() == kind
}

func (b *Board) updatePiece(p Piece) {
	oldPosition, ok := b.pieces[p]
	if !ok {
		panic("piece not on board")
	}

	for _, pt := range oldPosition {
		delete(b.positions, pt)
	}

	b.addPiece(p)
}

func (b *Board) addPiece(p Piece) {
	position := slices.Clone(p.Position())
	b.pieces[p] = position

	for _, pt := range position {
		if !pt.In(b.Bounds()) {
			panic("piece out of bounds")
		}
		b.positions[pt] = p
	}
}

// FormattedBoard is a type that represents a board in a human-readable format.
// This is mostly used for debugging.
type FormattedBoard string

const (
	FormattedPlayerAScoutPiece   = 'A'
	FormattedPlayerBScoutPiece   = 'B'
	FormattedPlayerABoulderPiece = 'a'
	FormattedPlayerBBoulderPiece = 'b'
	FormattedBlankPiece          = '.'
)

// FormatBoard returns a human-readable representation of the board.
func FormatBoard(board *Board) FormattedBoard {
	stride := BoardBounds.Dx() + 1
	buf := make([]byte, stride*BoardBounds.Dy())
	for i := range buf {
		buf[i] = FormattedBlankPiece
	}
	for i := stride - 1; i < len(buf); i += stride {
		buf[i] = '\n'
	}

	for pt, piece := range board.positions {
		var playerAPiece, playerBPiece byte
		switch piece.Kind() {
		case ScoutPieceKind:
			playerAPiece = FormattedPlayerAScoutPiece
			playerBPiece = FormattedPlayerBScoutPiece
		case BoulderPieceKind:
			playerAPiece = FormattedPlayerABoulderPiece
			playerBPiece = FormattedPlayerBBoulderPiece
		}

		var playerPiece byte
		switch piece.Player() {
		case PlayerA:
			playerPiece = playerAPiece
		case PlayerB:
			playerPiece = playerBPiece
		}

		o := pixOffset(pt, stride)
		buf[o] = playerPiece
	}

	return FormattedBoard(buf)
}

func pixOffset(pt Point, stride int) int {
	return pt.Y*stride + pt.X
}
