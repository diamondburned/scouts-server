package scouts

// PlaceScoutTurns is the number of turns that each player has to place their
// scouts. After this number of turns, other moves are allowed.
const PlaceScoutTurns = 5

// StartingPlaysPerTurn is the number of plays that each player has at the
// beginning of each turn. The start is always when the player must place a
// scout.
const StartingPlaysPerTurn = 1

// PlaysPerTurn is the number of plays that each player has at the beginning of
// each turn after the start.
const PlaysPerTurn = 2

// PastTurn is a type that represents a turn in the game. A game turn instance
// must represent a valid turn, meaning it must contian no invalid moves.
type PastTurn struct {
	// Moves is the list of past valid moves that the player has made.
	Moves []Move
	// Player is the player whose turn it is.
	Player Player
}

// CurrentTurn is a type that represents the current turn in the game. A current
// turn instance will not contain any invalid moves, but it may not be a
// complete turn.
type CurrentTurn struct {
	// Moves is the list of past valid moves that the player has made.
	Moves []Move
	// Plays is the number of plays that the player has left.
	Plays int
	// Player is the player whose turn it is.
	Player Player
}

func (t *CurrentTurn) hasEnoughPlays(cost int) bool {
	return t.Plays >= cost
}
