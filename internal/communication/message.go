package communication

import (
	"time"
)

type MessageType string

type MessageHeader struct {
	Type      MessageType `json:"messageType"`
	Timestamp time.Time   `json:"timestamp"`
}

type Message[T any] struct {
	MessageHeader
	Payload T `json:"payload"`
}
