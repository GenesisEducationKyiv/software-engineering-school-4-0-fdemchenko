package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/fdemchenko/exchanger/internal/communication/mailer"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type RateService interface {
	GetRate(context.Context, string) (float32, error)
}

type EmailService interface {
	Create(email string) error
	GetAll() ([]string, error)
}

type RabbitMQEmailScheduler struct {
	emailService EmailService
	rateService  RateService
	channel      *amqp.Channel
	queueName    string
	stopChan     chan bool
}

func NewEmailScheduler(
	emailService EmailService,
	rateService RateService,
	channel *amqp.Channel,
	queueName string,
) *RabbitMQEmailScheduler {
	return &RabbitMQEmailScheduler{
		rateService:  rateService,
		emailService: emailService,
		stopChan:     make(chan bool),
		channel:      channel,
		queueName:    queueName,
	}
}

func (es *RabbitMQEmailScheduler) StartBackhroundTask(updateInterval time.Duration) {
	ticker := time.NewTicker(updateInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				err := es.sendMessages()
				if err != nil {
					log.Error().Err(err).Send()
				}
			case <-es.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func (es *RabbitMQEmailScheduler) StopBackgroundTask() {
	es.stopChan <- true
}

func (es *RabbitMQEmailScheduler) sendMessages() error {
	rate, err := es.rateService.GetRate(context.Background(), "usd")
	if err != nil {
		return err
	}
	rateUpdateMessage := mailer.Message[mailer.ExchangeRateUpdatedEvent]{
		MessageHeader: mailer.MessageHeader{Type: mailer.ExchangeRateUpdated, Timestamp: time.Now()},
		Payload:       mailer.ExchangeRateUpdatedEvent{Rate: rate},
	}
	bytes, err := json.Marshal(rateUpdateMessage)
	if err != nil {
		return err
	}
	err = es.channel.Publish("", es.queueName, false, false,
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
		sendEmailMessage := mailer.Message[mailer.SendEmailNotificationCommand]{
			MessageHeader: mailer.MessageHeader{Type: mailer.SendEmailNotification, Timestamp: time.Now()},
			Payload:       mailer.SendEmailNotificationCommand{Email: email},
		}
		bytes, err := json.Marshal(sendEmailMessage)
		if err != nil {
			continue
		}
		err = es.channel.Publish("", es.queueName, false, false,
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
