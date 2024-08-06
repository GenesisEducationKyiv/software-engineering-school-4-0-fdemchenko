package mailer

import (
	"time"
)

type MessageType string

const (
	ExchangeRateUpdated   MessageType = "ExchangeRateUpdated"
	SendEmailNotification MessageType = "SendEmailNotification"
)

type MessageHeader struct {
	Type      MessageType `json:"messageType"`
	Timestamp time.Time   `json:"timestamp"`
}

type Message[T any] struct {
	MessageHeader
	Payload T `json:"payload"`
}

type ExchangeRateUpdatedEvent struct {
	Rate float32 `json:"rate"`
}

type SendEmailNotificationCommand struct {
	Email string `json:"email"`
}
