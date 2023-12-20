package gameserver

import (
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"libdb.so/scouts-server/scouts"
)

func TestRealGameTimer(t *testing.T) {
	start := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)
	sleep := func(d time.Duration) {
		start = start.Add(d)
	}
	hasTime := true

	timer := newRealGameTimer(start, Duration(10*time.Second), Duration(500*time.Millisecond))

	sleep(5 * time.Second)
	hasTime = timer.Subtract(start, scouts.Player1)
	assert.Equal(t, true, hasTime)
	assert.Equal(t, [2]Duration{
		Duration(5*time.Second) + Duration(500*time.Millisecond),
		Duration(10 * time.Second),
	}, timer.remaining)

	sleep(5 * time.Second)
	hasTime = timer.Subtract(start, scouts.Player2)
	assert.Equal(t, true, hasTime)
	assert.Equal(t, [2]Duration{
		Duration(5*time.Second) + Duration(500*time.Millisecond),
		Duration(5*time.Second) + Duration(500*time.Millisecond),
	}, timer.remaining)

	sleep(6 * time.Second)
	hasTime = timer.Subtract(start, scouts.Player1)
	assert.Equal(t, false, hasTime)
	assert.Equal(t, [2]Duration{
		Duration(0),
		Duration(5*time.Second) + Duration(500*time.Millisecond),
	}, timer.remaining)
}
