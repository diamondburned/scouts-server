package gameserver

import (
	"bytes"
	"crypto/rand"
	"strconv"
	"time"

	"github.com/godruoyi/go-snowflake"
	"github.com/oklog/ulid/v2"
)

// UserID is a type that represents a user ID.
type UserID uint64

// GenerateUserID generates a new user ID.
func GenerateUserID() UserID {
	return UserID(snowflake.ID())
}

// CreatedAt returns the time that the user ID was generated.
func (u UserID) CreatedAt() time.Time {
	return ulid.Time(snowflake.ParseID(uint64(u)).Timestamp)
}

// String returns the string representation of the user ID.
func (u UserID) String() string {
	return strconv.FormatUint(uint64(u), 16)
}

// MarshalText marshals the user ID into text.
func (u UserID) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

// UnmarshalText unmarshals the user ID from text.
func (u *UserID) UnmarshalText(text []byte) error {
	id, err := strconv.ParseUint(string(text), 16, 64)
	if err != nil {
		return err
	}
	*u = UserID(id)
	return nil
}

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

// Duration is equivalent to time.Duration, but it marshals to and from JSON
// as a float64 representing seconds. A negative duration is considered infinite.
type Duration time.Duration

const infinitySymbol = "∞"

// InfiniteDuration is a duration that represents infinity.
const InfiniteDuration = Duration(-1)

// InfiniteDurationPair is a pair of infinite durations.
var InfiniteDurationPair = [2]Duration{
	InfiniteDuration,
	InfiniteDuration,
}

// ToDuration converts the duration to a time.Duration.
func (d Duration) ToDuration() time.Duration {
	return time.Duration(d)
}

// String returns the string representation of the duration.
func (d Duration) String() string {
	if d < 0 {
		return infinitySymbol
	}
	return strconv.FormatFloat(time.Duration(d).Seconds(), 'f', -1, 64) + "s"
}

// MarshalText marshals the duration into text.
func (d Duration) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

// UnmarshalText unmarshals the duration from text.
func (d *Duration) UnmarshalText(text []byte) error {
	if string(text) == infinitySymbol {
		*d = InfiniteDuration
		return nil
	}

	text = bytes.TrimSuffix(text, []byte("s"))
	seconds, err := strconv.ParseFloat(string(text), 64)
	if err != nil {
		return err
	}

	*d = Duration(time.Duration(seconds * float64(time.Second)))
	return nil
}
