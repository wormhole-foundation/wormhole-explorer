package queue

// Message represents a message from a queue.
type Message interface {
	Data() []byte
	Done()
	Failed()
	IsExpired() bool
}
