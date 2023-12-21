package gameserver

import (
	"bytes"
	"strconv"
	"time"
)

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
