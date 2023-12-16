package scouts

import "fmt"

// Player is a type that represents a player.
type Player int

const (
	_ Player = iota
	PlayerA
	PlayerB
)

// ParsePlayer parses a player from a string.
// To convert a player to a string, use the String method.
func ParsePlayer(str string) (Player, error) {
	switch str {
	case "A":
		return PlayerA, nil
	case "B":
		return PlayerB, nil
	default:
		return 0, fmt.Errorf("invalid player: %q", str)
	}
}

// Opponent returns the opponent of the player.
func (p Player) Opponent() Player {
	switch p {
	case PlayerA:
		return PlayerB
	case PlayerB:
		return PlayerA
	default:
		panic(fmt.Errorf("invalid player: %v", p))
	}
}

// String returns the string representation of the player.
func (p Player) String() string {
	switch p {
	case PlayerA:
		return "A"
	case PlayerB:
		return "B"
	default:
		return fmt.Sprintf("Player(%d)", int(p))
	}
}

// Validate validates the player.
func (p Player) Validate() error {
	switch p {
	case PlayerA, PlayerB:
		return nil
	default:
		return fmt.Errorf("invalid player: %v", p)
	}
}
