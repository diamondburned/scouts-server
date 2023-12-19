package scouts

import (
	"strings"
)

// PossibleMoves represents a list of possible moves.
type PossibleMoves struct {
	// Moves is a list of possible moves. It never contains BoulderMove, since
	// that's a special case that's determined through the CanPlaceBoulder
	// field.
	Moves Moves `json:"moves"`
	// CanPlaceBoulder is whether or not the player can place a boulder.
	CanPlaceBoulder bool `json:"can_place_boulder"`
}

func (m PossibleMoves) String() string {
	strs := make([]string, 0, len(m.Moves)+1)
	for _, move := range m.Moves {
		strs = append(strs, move.String())
	}
	if m.CanPlaceBoulder {
		strs = append(strs, "boulder")
	}
	return strings.Join(strs, " | ")
}

func calculatePossibleMoves(g *Game, p Player) PossibleMoves {
	if g.currentTurn.Player != p {
		return PossibleMoves{}
	}

	switch g.currentState {
	case gameStatePlaceScouts:
		// bruteforceable
		var y int
		switch p {
		case PlayerA:
			y = playerABaseY
		case PlayerB:
			y = playerBBaseY
		}

		allScoutMoves := make([]Move, 0, BoardBounds.Dx())
		for x := BoardBounds.Min.X; x < BoardBounds.Max.X; x++ {
			move := &PlaceScoutMove{ScoutPosition: Pt(x, y)}
			if move.validate(g) == nil {
				allScoutMoves = append(allScoutMoves, move)
			}
		}

		return PossibleMoves{Moves: allScoutMoves}

	case gameStatePlay:
		moves := PossibleMoves{Moves: make([]Move, 0, 4)}

		if !g.playerPlacedBoulder(p) {
			moves.CanPlaceBoulder = true
		}

		if (*SkipMove)(nil).validate(g) == nil {
			moves.Moves = append(moves.Moves, &SkipMove{})
		}

		for piece := range g.board.pieces {
			scout, ok := piece.(*ScoutPiece)
			if !ok || scout.player != p {
				continue
			}

			// Check for dash moves.
			for _, move := range generateAllDashMoves(scout.position) {
				if move.validate(g) == nil {
					moves.Moves = append(moves.Moves, move)
				}
			}

			// Check for jump moves.
			for _, move := range generateAllJumpMoves(scout.position) {
				if move.validate(g) == nil {
					moves.Moves = append(moves.Moves, move)
				}
			}
		}

		return moves

	default:
		return PossibleMoves{}
	}
}

func generateAllDashMoves(scoutPosition Point) []*DashMove {
	return []*DashMove{
		{ScoutPosition: scoutPosition, Destination: scoutPosition.Add(Pt(+1, 0))},
		{ScoutPosition: scoutPosition, Destination: scoutPosition.Add(Pt(-1, 0))},
		{ScoutPosition: scoutPosition, Destination: scoutPosition.Add(Pt(0, +1))},
		{ScoutPosition: scoutPosition, Destination: scoutPosition.Add(Pt(0, -1))},
		{ScoutPosition: scoutPosition, Destination: scoutPosition.Add(Pt(+1, +1))},
		{ScoutPosition: scoutPosition, Destination: scoutPosition.Add(Pt(+1, -1))},
		{ScoutPosition: scoutPosition, Destination: scoutPosition.Add(Pt(-1, +1))},
		{ScoutPosition: scoutPosition, Destination: scoutPosition.Add(Pt(-1, -1))},
	}
}

func generateAllJumpMoves(scoutPosition Point) []*JumpMove {
	return []*JumpMove{
		{ScoutPosition: scoutPosition, Destination: scoutPosition.Add(Pt(+2, 0))},
		{ScoutPosition: scoutPosition, Destination: scoutPosition.Add(Pt(-2, 0))},
		{ScoutPosition: scoutPosition, Destination: scoutPosition.Add(Pt(0, +2))},
		{ScoutPosition: scoutPosition, Destination: scoutPosition.Add(Pt(0, -2))},
	}
}
