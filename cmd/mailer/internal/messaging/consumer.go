package messaging

import (
	"encoding/json"
	"errors"

	"github.com/fdemchenko/exchanger/cmd/mailer/internal/services"
	"github.com/fdemchenko/exchanger/internal/communication"
	"github.com/fdemchenko/exchanger/internal/communication/mailer"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type rateEmailsConsumer struct {
	channel       *amqp.Channel
	mailerService *services.MailerService
}

func NewRateEmailsConsumer(
	channel *amqp.Channel,
	mailerService *services.MailerService,
) *rateEmailsConsumer {
	return &rateEmailsConsumer{
		channel:       channel,
		mailerService: mailerService,
	}
}

func (rec *rateEmailsConsumer) handleDelivery(delivery amqp.Delivery) error {
	message := communication.Message[json.RawMessage]{}
	err := json.Unmarshal(delivery.Body, &message)
	if err != nil {
		return err
	}
	switch message.Type {
	case mailer.ExchangeRateUpdated:
		updateEvent := mailer.ExchangeRateUpdatedEvent{}
		err := json.Unmarshal(message.Payload, &updateEvent)
		if err != nil {
			return err
		}
		err = rec.mailerService.UpdateCurrencyRateTemplates(updateEvent.Rate)
		if err != nil {
			return err
		}
	case mailer.SendEmailNotification:
		sendCommand := mailer.SendEmailNotificationCommand{}
		err := json.Unmarshal(message.Payload, &sendCommand)
		if err != nil {
			return err
		}
		rec.mailerService.SendEmail(sendCommand.Email)
	default:
		return errors.New("unknown message type")
	}
	return nil
}

func (rec *rateEmailsConsumer) StartListening() error {
	var startingError error
	go func() {
		deliveries, err := rec.channel.Consume(
			mailer.RateEmailsQueue,
			"",
			false, false, false, false, nil,
		)
		if err != nil {
			startingError = err
			return
		}

		for delivery := range deliveries {
			if err := rec.handleDelivery(delivery); err != nil {
				log.Error().Err(err).Send()
			}
			if err := delivery.Ack(false); err != nil {
				log.Error().Err(err).Send()
			}
		}
	}()

	return startingError
}
