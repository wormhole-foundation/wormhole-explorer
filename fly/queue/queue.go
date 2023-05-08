package queue

import "context"

// Message represents a message from a queue.
type Message interface {
	Data() []byte
	Done(context.Context)
	Failed()
	IsExpired() bool
}
