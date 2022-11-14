package queue

type Message struct {
	Data []byte
	Ack  func()
}
