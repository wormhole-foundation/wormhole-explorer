package queue

// Message represents a message from a queue.
type Message struct {
	Data      []byte
	Ack       func()
	IsExpired func() bool
}
