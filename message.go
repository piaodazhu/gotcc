package gooc

type Message struct {
	Sender     uint32
	SenderName string
	// Finish     bool
	Value interface{}
}

type ErrorMessage struct {
	SenderName string
	Value      error
}
