package juggler

import "github.com/PuerkitoBio/exp/juggler/msg"

// PubSubConn defines the methods to manage subscriptions to events
// for a connection.
type PubSubConn interface {
	// Subscribe subscribes the connection to channel, which is treated
	// as a pattern if pattern is true.
	Subscribe(channel string, pattern bool) error

	// Unsubscribe unsubscribes the connection from the channel, which
	// is treated as a pattern if pattern is true.
	Unsubscribe(channel string, pattern bool) error

	// Events returns a stream of event payloads from events published
	// on channels that the connection is subscribed to.
	// The returned channel is closed when the connection is closed,
	// or when an error is received. Callers can call EventsErr on the
	// PubSubConn to check the error that caused the channel to be closed,
	// if any.
	//
	// Only the first call to Events starts the goroutine that listens to
	// events. Subsequent calls return the same channel, so that many
	// consumers can process events.
	Events() <-chan *msg.EvntPayload

	// EventsErr returns the error that caused the channel returned from
	// Events to be closed. Is only non-nil once the channel is closed.
	EventsErr() error
}
