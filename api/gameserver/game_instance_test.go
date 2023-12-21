package gameserver

import (
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/neilotoole/slogt"
	"libdb.so/scouts-server/api/user"
	"libdb.so/scouts-server/scouts"
)

func TestGameInstance(t *testing.T) {
	tests := []struct {
		name   string
		opts   CreateGameOptions
		replay func(*testing.T, *testingGameInstance)
	}{{
		name: "initial state testing",
		replay: func(t *testing.T, game *testingGameInstance) {
			killed := game.KillIfInactive(time.Hour)
			assert.False(t, killed, "game not expired yet has already been killed")

			// This should not do anything.
			game.Stop()

			state := game.StateSnapshot()
			assert.Equal(t, state.PlayerA, nil, "player A should be nil")
			assert.Equal(t, state.PlayerB, nil, "player B should be nil")
			assert.Equal(t, len(state.Moves), 0, "snapshot should have no moves")

			var err error
			err = game.MakeMove(game.User1, mustMove("place_scout 0,0"))
			assert.Error(t, err, "player was able to make move before game was ready")
			err = game.MakeMove(game.User2, mustMove("place_scout 0,0"))
			assert.Error(t, err, "player was able to make move before game was ready")
		},
	}, {
		name: "interrupted after both players join",
		replay: func(t *testing.T, game *testingGameInstance) {
			ev1, _ := game.join(t, game.User1)
			ev2, _ := game.join(t, game.User2)

			events := []GameEvent{
				PlayerJoinedEvent{scouts.Player1, ptr[user.UserID](1)},
				PlayerJoinedEvent{scouts.Player2, ptr[user.UserID](2)},
				TurnBeginEvent{
					PlayerSide:     scouts.Player1,
					PlaysRemaining: 1,
					TimeRemaining:  InfiniteDurationPair,
				},
			}

			expectEvents(t, ev1, events)
			expectEvents(t, ev2, events)
		},
	}, {
		name: "normal game up to placing scouts",
		replay: func(t *testing.T, game *testingGameInstance) {
			ev1, _ := game.join(t, game.User1)
			ev2, _ := game.join(t, game.User2)

			game.move(t, game.User1, mustMove("place_scout 0,9"))
			game.move(t, game.User2, mustMove("place_scout 0,0"))

			events := []GameEvent{
				PlayerJoinedEvent{scouts.Player1, ptr[user.UserID](1)},
				PlayerJoinedEvent{scouts.Player2, ptr[user.UserID](2)},
				TurnBeginEvent{
					PlayerSide:     scouts.Player1,
					PlaysRemaining: 1,
					TimeRemaining:  InfiniteDurationPair,
				},
				MoveMadeEvent{
					Move:           mustMove("place_scout 0,9"),
					PlayerSide:     scouts.Player1,
					PlaysRemaining: 0,
					TimeRemaining:  InfiniteDurationPair,
				},
				TurnBeginEvent{
					PlayerSide:     scouts.Player2,
					PlaysRemaining: 1,
					TimeRemaining:  InfiniteDurationPair,
				},
				MoveMadeEvent{
					Move:           mustMove("place_scout 0,0"),
					PlayerSide:     scouts.Player2,
					PlaysRemaining: 0,
					TimeRemaining:  InfiniteDurationPair,
				},
			}

			expectEvents(t, ev1, events)
			expectEvents(t, ev2, events)
		},
	}, {
		name: "normal game but illegal move",
		replay: func(t *testing.T, game *testingGameInstance) {
			ev1, _ := game.join(t, game.User1)
			ev2, _ := game.join(t, game.User2)

			err := game.MakeMove(game.User1, mustMove("jump 0,0 0,9"))
			assert.Error(t, err, "player was able to make illegal move")

			// you should still be able to make a legal move afterwards
			game.move(t, game.User1, mustMove("place_scout 0,9"))
			game.move(t, game.User2, mustMove("place_scout 0,0"))

			events := []GameEvent{
				PlayerJoinedEvent{scouts.Player1, ptr[user.UserID](1)},
				PlayerJoinedEvent{scouts.Player2, ptr[user.UserID](2)},
				TurnBeginEvent{
					PlayerSide:     scouts.Player1,
					PlaysRemaining: 1,
					TimeRemaining:  InfiniteDurationPair,
				},
				MoveMadeEvent{
					Move:           mustMove("place_scout 0,9"),
					PlayerSide:     scouts.Player1,
					PlaysRemaining: 0,
					TimeRemaining:  InfiniteDurationPair,
				},
				TurnBeginEvent{
					PlayerSide:     scouts.Player2,
					PlaysRemaining: 1,
					TimeRemaining:  InfiniteDurationPair,
				},
				MoveMadeEvent{
					Move:           mustMove("place_scout 0,0"),
					PlayerSide:     scouts.Player2,
					PlaysRemaining: 0,
					TimeRemaining:  InfiniteDurationPair,
				},
			}

			expectEvents(t, ev1, events)
			expectEvents(t, ev2, events)
		},
	}, {
		name: "normal game but player leaves",
		replay: func(t *testing.T, game *testingGameInstance) {
			_, stop1 := game.join(t, game.User1)
			ev2, _ := game.join(t, game.User2)

			game.move(t, game.User1, mustMove("place_scout 0,9"))
			game.move(t, game.User2, mustMove("place_scout 0,0"))

			events := []GameEvent{
				PlayerJoinedEvent{scouts.Player1, ptr[user.UserID](1)},
				PlayerJoinedEvent{scouts.Player2, ptr[user.UserID](2)},
				TurnBeginEvent{
					PlayerSide:     scouts.Player1,
					PlaysRemaining: 1,
					TimeRemaining:  InfiniteDurationPair,
				},
				MoveMadeEvent{
					Move:           mustMove("place_scout 0,9"),
					PlayerSide:     scouts.Player1,
					PlaysRemaining: 0,
					TimeRemaining:  InfiniteDurationPair,
				},
				TurnBeginEvent{
					PlayerSide:     scouts.Player2,
					PlaysRemaining: 1,
					TimeRemaining:  InfiniteDurationPair,
				},
				MoveMadeEvent{
					Move:           mustMove("place_scout 0,0"),
					PlayerSide:     scouts.Player2,
					PlaysRemaining: 0,
					TimeRemaining:  InfiniteDurationPair,
				},
				TurnBeginEvent{
					PlayerSide:     scouts.Player1,
					PlaysRemaining: 1,
					TimeRemaining:  InfiniteDurationPair,
				},
				PlayerLeftEvent{PlayerSide: scouts.Player1, UserID: ptr[user.UserID](1)},
				GoingAwayEvent{},
			}

			stop1()
			expectEvents(t, ev2, events)
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			game := newTestingGameInstance(t, test.opts)
			test.replay(t, game)
		})
	}
}

type testingGameInstance struct {
	*gameInstance
	User1 user.Authorized
	User2 user.Authorized
}

func newTestingGameInstance(t *testing.T, opts CreateGameOptions) *testingGameInstance {
	token1 := user.GenerateSessionToken()
	token2 := user.GenerateSessionToken()

	user1 := user.NewAnonymous(token1)
	user2 := user.NewAnonymous(token2)

	logger := slogt.New(t)
	game := newGameInstance(opts, logger, nil)

	return &testingGameInstance{
		gameInstance: game,
		User1:        user1,
		User2:        user2,
	}
}

func (g *testingGameInstance) join(t *testing.T, player user.Authorized) (<-chan GameEvent, func()) {
	ev, stop, err := g.PlayerJoinNext(player)
	assert.NoError(t, err, "player should be able to join")
	t.Cleanup(func() { stop() })
	return ev, stop
}

func (g *testingGameInstance) move(t *testing.T, user user.Authorized, move scouts.Move) {
	err := g.MakeMove(user, move)
	assert.NoError(t, err, "player should be able to make move")
}

func mustMove(move string) scouts.Move {
	m, err := scouts.ParseMove(move)
	if err != nil {
		panic(err)
	}
	return m
}

func expectEvents(t *testing.T, ch <-chan GameEvent, events []GameEvent) {
	for _, ev := range events {
		select {
		case actual, ok := <-ch:
			if !ok {
				t.Fatal("channel closed prematurely")
			}
			assert.Equal(t, ev, actual, "expected event")
			t.Logf("received event %T", actual)
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for event")
		}
	}
}

func assertChClosed[T any](t *testing.T, ch <-chan T, msg ...any) {
	if msg == nil {
		msg = []any{"channel should be closed"}
	}
	timeout := time.After(1 * time.Second)
drainLoop:
	for {
		select {
		case _, ok := <-ch:
			if ok {
				break drainLoop
			}
		case <-timeout:
			t.Fatal("timed out waiting for channel to close")
		}
	}
}

func ptr[T any](v T) *T {
	return &v
}
