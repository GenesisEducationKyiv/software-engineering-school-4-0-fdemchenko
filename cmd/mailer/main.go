package main

import (
	"encoding/json"

	"github.com/fdemchenko/exchanger/internal/communication/mailer"
	"github.com/fdemchenko/exchanger/internal/communication/rabbitmq"
	"github.com/rs/zerolog/log"
)

const (
	DefaultSMTPPort                 = 25
	DefaultRabbitMQPort             = 5672
	DefaultMailerConnectionPoolSize = 3
)

func main() {
	ch, err := rabbitmq.OpenWithQueueName(ServiceConfig.RabbitMQConnString, mailer.QueueName)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	deliveries, err := ch.Consume(mailer.QueueName, "", true, false, false, false, nil)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	log.Info().Msg("Mialer service started")

	mailerService := NewMailerService(ServiceConfig.SMTP)
	mailerService.StartWorkers(ServiceConfig.SMTP.ConnectionPoolSize)

	for delivery := range deliveries {
		message := mailer.Message[json.RawMessage]{}
		err := json.Unmarshal(delivery.Body, &message)
		if err != nil {
			log.Error().Err(err).Send()
			continue
		}

		log.Debug().Str("message type", string(message.Type)).
			Time("timestamp", message.Timestamp).
			Msg("New message received")

		err = mailerService.HandleMessage(message)
		if err != nil {
			log.Error().Err(err).Send()
		}
	}
}
