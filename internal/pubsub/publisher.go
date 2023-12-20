package pubsub

import (
	"github.com/puzpuzpuz/xsync/v3"
)

// Publisher allows you to publish events to multiple subscribers.
type Publisher[T any] struct {
	subscribers *xsync.MapOf[*ConcurrentQueue[T], chan struct{}]
}

// NewPublisher creates a new publisher.
func NewPublisher[T any]() *Publisher[T] {
	return &Publisher[T]{
		subscribers: xsync.NewMapOf[*ConcurrentQueue[T], chan struct{}](),
	}
}

// Subscribe subscribes to the publisher.
func (p *Publisher[T]) Subscribe(cq *ConcurrentQueue[T]) {
	p.subscribers.Store(cq, make(chan struct{}))
}

// Unsubscribe unsubscribes from the publisher.
// Any pending sends to the channel will be cancelled.
func (p *Publisher[T]) Unsubscribe(cq *ConcurrentQueue[T]) {
	stopper, ok := p.subscribers.LoadAndDelete(cq)
	if ok {
		close(stopper)
	}
}

// Subscribers returns a list of all subscribers.
func (p *Publisher[T]) Subscribers() []*ConcurrentQueue[T] {
	var subscribers []*ConcurrentQueue[T]
	p.subscribers.Range(func(cq *ConcurrentQueue[T], stop chan struct{}) bool {
		subscribers = append(subscribers, cq)
		return true
	})
	return subscribers
}

// Publish publishes an event to all subscribers.
// It will block until all subscribers have received the event or the context
// has been cancelled.
func (p *Publisher[T]) Publish(events ...T) {
	p.subscribers.Range(func(cq *ConcurrentQueue[T], stop chan struct{}) bool {
		for _, ev := range events {
			select {
			case <-stop:
				return false
			case cq.In() <- ev:
			}
		}
		return true
	})
}

// Send sends events to a channel.
func Send[T any](ch chan<- T, events ...T) {
	for _, event := range events {
		ch <- event
	}
}
