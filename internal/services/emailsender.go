package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/fdemchenko/exchanger/internal/communication"
	"github.com/fdemchenko/exchanger/internal/communication/mailer"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type RateService interface {
	GetRate(context.Context, string) (float32, error)
}

type EmailService interface {
	Create(email string) (int, error)
	GetAll() ([]string, error)
}

type RabbitMQEmailSender struct {
	emailService EmailService
	rateService  RateService
	channel      *amqp.Channel
}

func NewRabbitMQEmailSender(
	emailService EmailService,
	rateService RateService,
	channel *amqp.Channel,
) *RabbitMQEmailSender {
	return &RabbitMQEmailSender{
		rateService:  rateService,
		emailService: emailService,
		channel:      channel,
	}
}

func (es *RabbitMQEmailSender) SendMessages() error {
	rate, err := es.rateService.GetRate(context.Background(), "usd")
	if err != nil {
		return err
	}
	rateUpdateMessage := communication.Message[mailer.ExchangeRateUpdatedEvent]{
		MessageHeader: communication.MessageHeader{Type: mailer.ExchangeRateUpdated, Timestamp: time.Now()},
		Payload:       mailer.ExchangeRateUpdatedEvent{Rate: rate},
	}
	bytes, err := json.Marshal(rateUpdateMessage)
	if err != nil {
		return err
	}
	err = es.channel.Publish("", mailer.RateEmailsQueue, false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        bytes,
		},
	)
	if err != nil {
		return err
	}

	emails, err := es.emailService.GetAll()
	if err != nil {
		return err
	}

	for _, email := range emails {
		sendEmailMessage := communication.Message[mailer.SendEmailNotificationCommand]{
			MessageHeader: communication.MessageHeader{Type: mailer.SendEmailNotification, Timestamp: time.Now()},
			Payload:       mailer.SendEmailNotificationCommand{Email: email},
		}
		bytes, err := json.Marshal(sendEmailMessage)
		if err != nil {
			continue
		}
		err = es.channel.Publish("", mailer.RateEmailsQueue, false, false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        bytes,
			},
		)
		if err != nil {
			log.Error().Err(err).Send()
		}
	}
	return nil
}
