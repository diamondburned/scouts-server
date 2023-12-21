package context

import (
	"context"
	"time"
)

// Context is an alias for context.Context.
type Context = context.Context

type key[T any] struct{}

// With returns a copy of parent that contains the given value which can be
// retrieved by calling From with the resulting context.
func With[T any](ctx context.Context, v T) context.Context {
	return context.WithValue(ctx, key[T]{}, v)
}

// From returns the value associated with the wanted type.
func From[T any](ctx context.Context) T {
	v, _ := ctx.Value(key[T]{}).(T)
	return v
}

// WithCancel returns a copy of parent with a new Done channel.
func WithCancel(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithCancel(parent)
}

// WithDeadline returns a copy of parent with the deadline adjusted to be no later than d.
func WithDeadline(parent context.Context, d time.Time) (context.Context, context.CancelFunc) {
	return context.WithDeadline(parent, d)
}

// WithTimeout returns WithDeadline(parent, time.Now().Add(timeout)).
func WithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, timeout)
}
