package gameserver

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"libdb.so/hrt"
	"libdb.so/scouts-server/api/user"
	"libdb.so/scouts-server/internal/pubsub"
	"libdb.so/scouts-server/scouts"
)

// GameSnapshot is a snapshot of a game, which can be used to restore it.

type gameInstance struct {
	// constant fields
	game   *scouts.Game
	logger *slog.Logger
	events *pubsub.Publisher[GameEvent]
	clock  customClock

	// mutable, mutex-guarded fields
	mu     sync.Mutex
	state  GameState
	timer  gameTimer
	stopCh chan struct{}
	waitg  sync.WaitGroup
}

func newGameInstance(opts CreateGameOptions, logger *slog.Logger, clock customClock) *gameInstance {
	return &gameInstance{
		game:   scouts.NewGame(),
		logger: logger.With("component", "api/gameserver/gamemanager.gameInstance"),
		events: pubsub.NewPublisher[GameEvent](),
		clock:  clock,
		state: GameState{
			CreatedAt: clock.Now(),
			Metadata:  opts,
		},
	}
}

func (g *gameInstance) sendEvent(evs ...GameEvent) {
	for _, ev := range evs {
		g.logger.Debug(
			"sending game event",
			"event_type", ev.Type(),
			"event", ev,
			"player_a", user.OptionalAuthorizedUserString(g.state.PlayerA),
			"player_b", user.OptionalAuthorizedUserString(g.state.PlayerB),
			"moves", len(g.state.Moves))
	}
	g.events.Publish(evs...)
}

// KillIfInactive kills the game if it has been inactive for the given TTL.
// True is returned if the game was killed.
func (g *gameInstance) KillIfInactive(ttl time.Duration) bool {
	g.mu.Lock()

	lastActiveAt := g.state.CreatedAt
	if g.state.BeganAt != nil {
		lastActiveAt = *g.state.BeganAt
	}
	if len(g.state.Moves) > 0 {
		lastActiveAt = g.state.Moves[len(g.state.Moves)-1].Time
	}

	kill := time.Since(lastActiveAt) > ttl

	g.logger.Debug(
		"checking game for inactivity",
		"last_active_at", lastActiveAt,
		"ttl", ttl,
		"kill", kill)

	g.mu.Unlock()

	if kill {
		g.Stop()
		return true
	}

	return false
}

func (g *gameInstance) Stop() {
	g.mu.Lock()
	if g.stopCh != nil {
		close(g.stopCh)
		g.stopCh = nil
		g.logger.Debug("game has stopped")
	} else {
		g.logger.Debug("game has already stopped")
	}
	g.mu.Unlock()

	g.waitg.Wait()
	g.logger.Debug("game has stopped and goroutines have finished")
}

func (g *gameInstance) startIfReady() {
	if !g.state.hasBothPlayers() || g.stopCh != nil {
		g.logger.Debug(
			"game is not ready to start",
			"player_a", user.OptionalAuthorizedUserString(g.state.PlayerA),
			"player_b", user.OptionalAuthorizedUserString(g.state.PlayerB))
		return
	}

	g.logger.Debug(
		"game is ready to start or resume",
		"player_a", user.OptionalAuthorizedUserString(g.state.PlayerA),
		"player_b", user.OptionalAuthorizedUserString(g.state.PlayerB),
		"started", g.state.BeganAt != nil)

	if g.state.BeganAt == nil {
		// Actually start the game here.
		now := g.clock.Now()
		g.state.BeganAt = &now

		// Reset the timer as well.
		g.timer = newGameTimer(now,
			g.state.Metadata.TimeLimit,
			g.state.Metadata.Increment)
	}

	events := playbackGameEvents(g.state)
	g.sendEvent(events...)

	g.stopCh = make(chan struct{})
	g.waitg.Add(1)
	go func(stop <-chan struct{}) {
		defer g.waitg.Done()

		endTimer := time.NewTimer(1 * time.Second)
		defer endTimer.Stop()

	timerLoop:
		for {
			select {
			case <-stop:
				g.logger.Debug("game stop signal received, going away")
				break timerLoop

			case now := <-endTimer.C:
				g.mu.Lock()

				turn := g.game.CurrentTurn()
				if !g.timer.Subtract(now, turn.Player) {
					g.mu.Unlock()

					g.logger.Debug(
						"player ran out of time",
						"player", turn.Player)
					break timerLoop
				}

				if minRemaining(g.timer) < Duration(5*time.Second) {
					// Ramp up the timer to be more precise once we get close to the
					// end.
					endTimer.Stop()
					endTimer.Reset(250 * time.Millisecond)
				}

				g.mu.Unlock()
			}
		}

		g.mu.Lock()
		defer g.mu.Unlock()

		now := g.clock.Now()
		turn := g.game.CurrentTurn()

		if !g.timer.Subtract(now, turn.Player) {
			// Someone ran out of time.
			g.sendEvent(GameEndEvent{
				Winner:        turn.Player.Opponent(),
				TimeRemaining: g.timer.Remaining(),
			})
		}

		// TODO(diamondburned): figure out how to do this properly.
		g.sendEvent(GoingAwayEvent{})

		for _, sub := range g.events.Subscribers() {
			sub.Close()
			g.events.Unsubscribe(sub)
			g.logger.Debug(
				"closed and unsubscribed game event subscriber")
		}
	}(g.stopCh)
}

func (g *gameInstance) MakeMove(user user.Authorized, move scouts.Move) error {
	now := g.clock.Now()

	g.mu.Lock()
	defer g.mu.Unlock()

	var player scouts.Player
	switch {
	case g.state.PlayerA != nil && g.state.PlayerA.Eq(user):
		player = scouts.PlayerA
	case g.state.PlayerB != nil && g.state.PlayerB.Eq(user):
		player = scouts.PlayerB
	default:
		return fmt.Errorf("%w: invalid session token", ErrInvalidMove)
	}

	turn := g.game.CurrentTurn()
	if turn.Player != player {
		return fmt.Errorf("%w: not your turn", ErrInvalidMove)
	}

	if !g.timer.Subtract(now, player) {
		return fmt.Errorf("%w: out of time", ErrInvalidMove)
	}

	events, err := makeMoveForEvents(g.game, player, move, g.timer)
	if err != nil {
		return err
	}

	g.state.Moves = append(g.state.Moves, MoveSnapshot{
		Player: player,
		Move:   move,
		Time:   now,
	})

	g.sendEvent(events...)
	return nil
}

// StateSnapshot returns a snapshot of the current state of the game.
func (g *gameInstance) StateSnapshot() GameState {
	g.mu.Lock()
	defer g.mu.Unlock()

	s := g.state
	s.SnapshotAt = g.clock.Now()

	return s
}

func (g *gameInstance) PlayerJoinNext(user user.Authorized) (<-chan GameEvent, func(), error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	var player scouts.Player

	switch {
	case g.state.PlayerA == nil || g.state.PlayerA.Eq(user):
		g.state.PlayerA = &user
		player = scouts.PlayerA
	case g.state.PlayerB == nil || g.state.PlayerB.Eq(user):
		g.state.PlayerB = &user
		player = scouts.PlayerB
	default:
		return nil, nil, ErrGameFull
	}

	g.sendEvent(PlayerJoinedEvent{
		PlayerSide: player,
		UserID:     user.User,
	})

	g.startIfReady()

	queue := pubsub.NewConcurrentQueue[GameEvent]()
	queue.Start()

	in := queue.In()

	pubsub.Send(in, playbackPlayerJoinEvents(g.state)...)
	pubsub.Send(in, playbackGameEvents(g.state)...)
	g.events.Subscribe(queue)

	return queue.Out(), func() {
		defer queue.Stop()
		defer g.events.Unsubscribe(queue)

		g.sendEvent(PlayerLeftEvent{
			PlayerSide: player,
			UserID:     user.User,
		})

		g.mu.Lock()
		switch player {
		case scouts.PlayerA:
			g.state.PlayerA = nil
		case scouts.PlayerB:
			g.state.PlayerB = nil
		}
		g.mu.Unlock()

		g.logger.Debug(
			"player left and mutex unlocked, stopping game",
			"player", player)

		g.Stop()
	}, nil
}

func turnBeginEvent(game *scouts.Game, timer gameTimer) TurnBeginEvent {
	turn := game.CurrentTurn()
	return TurnBeginEvent{
		PlayerSide:     turn.Player,
		PlaysRemaining: turn.Plays,
		TimeRemaining:  timer.Remaining(),
	}
}

func makeMoveForEvents(game *scouts.Game, player scouts.Player, move scouts.Move, timer gameTimer) ([]GameEvent, error) {
	events := make([]GameEvent, 0, 2)
	last := game.CurrentTurn()

	if err := game.MakeMove(player, move); err != nil {
		return nil, hrt.WrapHTTPError(http.StatusBadRequest, err)
	}

	turn := game.CurrentTurn()

	playsRemaining := turn.Plays
	if turn.Player != last.Player {
		playsRemaining = 0
	}

	events = append(events, MoveMadeEvent{
		Move:           move,
		PlayerSide:     player,
		PlaysRemaining: playsRemaining,
		TimeRemaining:  timer.Remaining(),
	})

	if winner, ended := game.Ended(); ended {
		events = append(events, GameEndEvent{
			Winner:        winner,
			TimeRemaining: timer.Remaining(),
		})
		return events, nil
	}

	if turn.Player != last.Player {
		events = append(events, TurnBeginEvent{
			PlayerSide:     turn.Player,
			PlaysRemaining: turn.Plays,
			TimeRemaining:  timer.Remaining(),
		})
		return events, nil
	}

	return events, nil
}

func playbackPlayerJoinEvents(state GameState) []GameEvent {
	var events []GameEvent
	if state.PlayerA != nil {
		events = append(events, PlayerJoinedEvent{
			PlayerSide: scouts.PlayerA,
			UserID:     state.PlayerA.User,
		})
	}
	if state.PlayerB != nil {
		events = append(events, PlayerJoinedEvent{
			PlayerSide: scouts.PlayerB,
			UserID:     state.PlayerB.User,
		})
	}
	return events
}

func playbackGameEvents(state GameState) []GameEvent {
	if state.BeganAt == nil {
		return nil
	}

	game := scouts.NewGame()
	timer := newGameTimer(*state.BeganAt, state.Metadata.TimeLimit, state.Metadata.Increment)

	events := []GameEvent{turnBeginEvent(game, timer)}
	for _, move := range state.Moves {
		timer.Subtract(move.Time, move.Player)
		moveEvents, _ := makeMoveForEvents(game, move.Player, move.Move, timer)
		events = append(events, moveEvents...)
	}

	return events
}
