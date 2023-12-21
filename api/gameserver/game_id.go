package gameserver

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

// GameID is a type that represents a game ID.
type GameID ulid.ULID

// GenerateGameID generates a new game ID.
func GenerateGameID() GameID {
	return GameID(ulid.MustNew(ulid.Now(), rand.Reader))
}

// CreatedAt returns the time that the game ID was generated.
func (g GameID) CreatedAt() time.Time {
	return ulid.Time((ulid.ULID)(g).Time())
}

// String returns the string representation of the game ID.
func (g GameID) String() string {
	return ulid.ULID(g).String()
}

// MarshalText marshals the game ID into text.
func (g GameID) MarshalText() ([]byte, error) {
	return []byte(g.String()), nil
}

// UnmarshalText unmarshals the game ID from text.
func (g *GameID) UnmarshalText(text []byte) error {
	id, err := ulid.Parse(string(text))
	if err != nil {
		return err
	}
	*g = GameID(id)
	return nil
}
