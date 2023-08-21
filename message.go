package qstash

import (
	"net/http"
)

// Message published to or received from a qstash queue
type Message struct {
	ID             string
	Headers        http.Header
	Body           []byte
	Retried        int
	w              http.ResponseWriter
	isAcknowledged bool
}

// Ack acknowledges the message.
// If ack is not called, the message will be retried.
func (m *Message) Ack() {
	m.isAcknowledged = true
	m.w.WriteHeader(http.StatusOK)
}
