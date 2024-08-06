package mailer

import "github.com/fdemchenko/exchanger/internal/communication"

const RateEmailsQueue = "emails"
const TriggerEmailsSendingQueue = "email_trigger"

const (
	ExchangeRateUpdated   communication.MessageType = "ExchangeRateUpdated"
	SendEmailNotification communication.MessageType = "SendEmailNotification"
	StartEmailSending     communication.MessageType = "StartEmailSending"
)

type ExchangeRateUpdatedEvent struct {
	Rate float32 `json:"rate"`
}

type SendEmailNotificationCommand struct {
	Email string `json:"email"`
}
