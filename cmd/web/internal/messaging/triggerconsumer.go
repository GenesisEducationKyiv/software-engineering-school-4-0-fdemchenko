package messaging

import (
	"encoding/json"

	"github.com/fdemchenko/exchanger/internal/communication"
	"github.com/fdemchenko/exchanger/internal/communication/mailer"
	"github.com/fdemchenko/exchanger/internal/services"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type emailTriggerConsumer struct {
	channel             *amqp.Channel
	rabbitMQEmailSender *services.RabbitMQEmailSender
}

func NewEmailTriggerConsumer(
	channel *amqp.Channel,
	rabbitMQEmailSender *services.RabbitMQEmailSender,
) *emailTriggerConsumer {
	return &emailTriggerConsumer{
		channel:             channel,
		rabbitMQEmailSender: rabbitMQEmailSender,
	}
}

func (etc *emailTriggerConsumer) StartListening() error {
	var startingError error
	go func() {
		deliveries, err := etc.channel.Consume(
			mailer.TriggerEmailsSendingQueue,
			"",
			false, false, false, false, nil,
		)
		if err != nil {
			startingError = err
			return
		}

		for delivery := range deliveries {
			if err := etc.handleDelivery(delivery); err != nil {
				log.Error().Err(err).Send()
			}
			if err := delivery.Ack(false); err != nil {
				log.Error().Err(err).Send()
			}
		}
	}()

	return startingError
}

func (etc *emailTriggerConsumer) handleDelivery(delivery amqp.Delivery) error {
	msg := communication.Message[json.RawMessage]{}
	err := json.Unmarshal(delivery.Body, &msg)
	if err != nil {
		return err
	}
	switch msg.Type {
	case mailer.StartEmailSending:
		err := etc.rabbitMQEmailSender.SendMessages()
		if err != nil {
			log.Error().Err(err).Msg("Sending emails to message borker failed")
		}
	default:
		log.Error().Msg("Invalid message type emails sending trigger")
	}
	return nil
}
