package gameserver

import (
	"time"

	"libdb.so/scouts-server/scouts"
)

type gameTimer interface {
	// Subtract subtracts the elapsed time from the player's remaining time.
	// It returns whether the player has time remaining.
	Subtract(now time.Time, player scouts.Player) (keepGoing bool)
	// Remaining returns the remaining time for both players.
	Remaining() [2]Duration
}

func newGameTimer(now time.Time, timeLimit, increment Duration) gameTimer {
	if timeLimit == 0 {
		return &nilGameTimer{}
	}
	return newRealGameTimer(now, timeLimit, increment)
}

func minRemaining(g gameTimer) Duration {
	remaining := g.Remaining()
	return min(remaining[0], remaining[1])
}

// nilGameTimer is a gameTimer that does nothing. It is used when the game
// has no time limit.
type nilGameTimer struct{}

func (nilGameTimer) Subtract(now time.Time, player scouts.Player) bool {
	return true
}

func (nilGameTimer) Remaining() [2]Duration {
	return [2]Duration{
		InfiniteDuration,
		InfiniteDuration,
	}
}

// realGameTimer is a gameTimer that keeps track of the remaining time for
// each player.
type realGameTimer struct {
	lastTick  time.Time
	increment Duration
	remaining [2]Duration
}

func newRealGameTimer(now time.Time, timeLimit, increment Duration) *realGameTimer {
	return &realGameTimer{
		lastTick:  now,
		increment: increment,
		remaining: [2]Duration{
			timeLimit,
			timeLimit,
		},
	}
}

func (g *realGameTimer) Subtract(now time.Time, player scouts.Player) (keepGoing bool) {
	elapsed := now.Sub(g.lastTick)
	g.lastTick = now

	g.remaining[player-1] -= Duration(elapsed)
	if g.remaining[player-1] < 0 {
		g.remaining[player-1] = 0
		return false
	} else {
		g.remaining[player-1] += g.increment
		return true
	}
}

func (g *realGameTimer) Remaining() [2]Duration {
	return g.remaining
}

type customClock func() time.Time

func (n customClock) Now() time.Time {
	if n == nil {
		return time.Now()
	}
	return n()
}

type fakeClock struct {
	now time.Time
}

func (f *fakeClock) Clock() customClock {
	return func() time.Time { return f.now }
}
