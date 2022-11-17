package queue

type Message struct {
	Data      []byte
	Ack       func()
	IsExpired func() bool
}
