package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fdemchenko/exchanger/cmd/mailer/internal/config"
	"github.com/fdemchenko/exchanger/cmd/mailer/internal/messaging"
	"github.com/fdemchenko/exchanger/cmd/mailer/internal/services"
	"github.com/fdemchenko/exchanger/internal/communication"
	"github.com/fdemchenko/exchanger/internal/communication/mailer"
	"github.com/fdemchenko/exchanger/internal/communication/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/robfig/cron"
	"github.com/rs/zerolog/log"
)

const (
	DefaultMailerConnectionPoolSize = 3
	EveryDayAt10AMCRON              = "0 0 10 * * *"
)

func main() {
	cfg := config.LoadConfig()
	rabbitMQConn, err := amqp.Dial(cfg.RabbitMQConnString)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	log.Info().Msg("Coonected to RabbitMQ successfully")

	rateEmailsChannel, err := rabbitmq.OpenWithQueueName(rabbitMQConn, mailer.RateEmailsQueue)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	emailsTriggersChannel, err := rabbitmq.OpenWithQueueName(rabbitMQConn, mailer.TriggerEmailsSendingQueue)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	mailerService := services.NewMailerService(cfg.SMTP)
	mailerService.StartWorkers(cfg.SMTP.ConnectionPoolSize)

	producer := rabbitmq.NewGenericProducer(emailsTriggersChannel)
	c := cron.New()
	err = c.AddFunc(EveryDayAt10AMCRON, func() {
		msg := communication.Message[struct{}]{
			MessageHeader: communication.MessageHeader{Type: mailer.StartEmailSending, Timestamp: time.Now()},
		}
		err := producer.SendMessage(msg, mailer.TriggerEmailsSendingQueue)
		if err != nil {
			log.Error().Err(err).Send()
		}
	})
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	c.Start()

	consumer := messaging.NewRateEmailsConsumer(rateEmailsChannel, mailerService)
	err = consumer.StartListening()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	log.Info().Msg("Mialer service started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if err := rabbitMQConn.Close(); err != nil {
		log.Error().Err(err).Msg("Cannot close RabbitMQ connection")
	}
	c.Stop()
}
