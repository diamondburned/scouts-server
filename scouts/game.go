package scouts

import (
	"fmt"
	"slices"
)

type gameState int

const (
	gameStatePlaceScouts gameState = iota
	gameStatePlay
	gameStateEnd
)

// Game is a game instance.
type Game struct {
	board          *Board
	turns          []PastTurn
	currentTurn    CurrentTurn
	currentState   gameState
	placedBoulders [2]bool
}

// NewGame returns a new game instance.
func NewGame() *Game {
	return &Game{
		board: NewBoard(),
		turns: make([]PastTurn, 0, 12),
		currentTurn: CurrentTurn{
			Moves:  make([]Move, 0, StartingPlaysPerTurn),
			Plays:  StartingPlaysPerTurn,
			Player: PlayerA,
		},
		currentState: gameStatePlaceScouts,
	}
}

// NewGameFromPastTurns returns a new game instance from the given past turns.
func NewGameFromPastTurns(turns []PastTurn) (*Game, error) {
	g := NewGame()
	for i, turn := range turns {
		for _, move := range turn.Moves {
			if err := g.Apply(turn.Player, move); err != nil {
				return nil, fmt.Errorf(
					"failed to apply turn %d, move %q for player %v: %v",
					i+1, move, turn.Player, err)
			}
		}
	}
	return g, nil
}

// Apply applies the given move to the game.
func (g *Game) Apply(p Player, move Move) error {
	if g.currentTurn.Player != p {
		return fmt.Errorf("it is not %v's turn", p)
	}
	return move.apply(g)
}

// Board returns the board.
func (g *Game) Board() *Board {
	return g.board
}

// PastTurns returns the past turns. It only contains valid turns.
// It does not contain the current turn.
func (g *Game) PastTurns() []PastTurn {
	return slices.Clone(g.turns)
}

// PlayerPastTurns returns the past turns for the given player. It only contains
// valid turns. It does not contain the current turn.
func (g *Game) PlayerPastTurns(p Player) []PastTurn {
	turns := make([]PastTurn, 0, len(g.turns)/2)
	for _, turn := range g.turns {
		if turn.Player == p {
			turns = append(turns, turn)
		}
	}
	return turns
}

// CurrentTurn returns the current turn.
func (g *Game) CurrentTurn() CurrentTurn {
	return g.currentTurn
}

// addMove adds the given move to the current turn. If the current turn is
// complete, it will add the current turn to the past turns and start a new
// current turn and return true. Otherwise, it will return false.
func (g *Game) addMove(move Move, cost int) (ended bool) {
	g.currentTurn.Moves = append(g.currentTurn.Moves, move)
	g.currentTurn.Plays -= cost
	if g.currentTurn.Plays > 0 {
		return false
	}

	g.turns = append(g.turns, PastTurn{
		Moves:  g.currentTurn.Moves,
		Player: g.currentTurn.Player,
	})

	g.currentTurn = CurrentTurn{
		Moves:  make([]Move, 0, PlaysPerTurn),
		Plays:  PlaysPerTurn,
		Player: g.currentTurn.Player.Opponent(),
	}

	return true
}

// addEndMove calls addMove and asserts that the current turn is complete.
func (g *Game) addEndMove(move Move, cost int) {
	if !g.addMove(move, cost) {
		panic("current turn is not complete")
	}
}
