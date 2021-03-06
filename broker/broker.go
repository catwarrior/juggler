// Package broker defines the generic interfaces that a broker must
// implement in order to act as a juggler broker. The redisbroker
// package implements those interfaces against a redis backend.
package broker

import (
	"time"

	"github.com/PuerkitoBio/juggler/message"
	"github.com/pborman/uuid"
)

// DefaultCallTimeout is the default timeout to use for a call
// request to expire. If no result is available before this delay,
// no result will ever be sent. Callers can set a message-specific
// timeout, this value is only used if no timeout was specified
// on the message. It should not be set to less than 1ms.
var DefaultCallTimeout = time.Minute

// CallerBroker defines the methods for a broker in the caller role.
type CallerBroker interface {
	// NewResultsConn returns a new ResultsConn that can be used
	// to process results from calls for the specified connection UUID.
	NewResultsConn(uuid.UUID) (ResultsConn, error)

	// Call registers a call request in the broker.
	Call(cp *message.CallPayload, timeout time.Duration) error
}

// CalleeBroker defines the methods for a broker in the callee role.
type CalleeBroker interface {
	// NewCallsConn returns a new CallsConn that can be used to
	// process call requests for the specified URIs. For use in
	// a redis cluster, all URIs must belong to the same
	// cluster slot.
	NewCallsConn(uris ...string) (CallsConn, error)

	// Result registers a call result in the broker.
	Result(rp *message.ResPayload, timeout time.Duration) error
}

// PubSubBroker defines the methods for a broker in the pub-sub role.
type PubSubBroker interface {
	// NewPubSubConn returns a new PubSubConn that can be used to
	// manage subscriptions to pub-sub channels, and to process
	// events sent on subscribed channels.
	NewPubSubConn() (PubSubConn, error)

	// Publish publishes an event on the specified channel.
	Publish(channel string, pp *message.PubPayload) error
}

// ResultsConn defines the methods to list the results from calls
// made on the ResultsConn connection UUID.
type ResultsConn interface {
	// Results returns a stream of call results for the connection UUID used
	// to create the ResultsConn. The returned channel is closed when the
	// connection is closed, or when an error occurs. Callers can call
	// ResultsErr to check the error that caused the channel to be closed.
	//
	// Only the first call to Results starts the goroutine that checks
	// for results. Subsequent calls return the same channel, so that many
	// consumers can process results.
	Results() <-chan *message.ResPayload

	// ResultsErr returns the error that caused the channel returned from
	// Results to be closed. Is only non-nil once the channel is closed.
	ResultsErr() error

	// Close closes the connection.
	Close() error
}

// CallsConn defines the methods to list the call requests using this
// connection.
type CallsConn interface {
	// Calls returns a stream of call requests for the URI used to
	// create the CallConn. The returned channel is closed when the
	// connection is closed, or when an error occurs. Callers can call
	// CallsErr to check the error that caused the channel to be closed.
	//
	// Only the first call to Calls starts the goroutine that checks
	// for requests. Subsequent calls return the same channel, so that many
	// consumers can process calls.
	Calls() <-chan *message.CallPayload

	// CallsErr returns the error that caused the channel returned from
	// Calls to be closed. Is only non-nil once the channel is closed.
	CallsErr() error

	// Close closes the connection.
	Close() error
}

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
	// PubSubConn to check the error that caused the channel to be closed.
	//
	// Only the first call to Events starts the goroutine that listens to
	// events. Subsequent calls return the same channel, so that many
	// consumers can process events.
	Events() <-chan *message.EvntPayload

	// EventsErr returns the error that caused the channel returned from
	// Events to be closed. Is only non-nil once the channel is closed.
	EventsErr() error

	// Close closes the connection.
	Close() error
}
