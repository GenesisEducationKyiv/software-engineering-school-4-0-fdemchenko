package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/VictoriaMetrics/metrics"
	mailerconfig "github.com/fdemchenko/exchanger/cmd/mailer/internal/config"
	"github.com/fdemchenko/exchanger/cmd/mailer/internal/messaging"
	"github.com/fdemchenko/exchanger/cmd/mailer/internal/services"
	"github.com/fdemchenko/exchanger/internal/communication"
	"github.com/fdemchenko/exchanger/internal/communication/mailer"
	"github.com/fdemchenko/exchanger/internal/communication/rabbitmq"
	"github.com/fdemchenko/exchanger/internal/config"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/robfig/cron"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config.MustLoad[mailerconfig.Config](os.Getenv("MAILER_CONFIG_PATH"))
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

	mailerService := services.NewMailerService(cfg.SMTPServer)
	mailerService.StartWorkers(cfg.SMTPServer.ConnectionPoolSize)

	producer := rabbitmq.NewGenericProducer(emailsTriggersChannel)
	c := cron.New()
	err = c.AddFunc(cfg.SchedulerCRON, func() {
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
	mux := http.NewServeMux()
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics.WritePrometheus(w, false)
	})
	s := http.Server{
		ReadHeaderTimeout: cfg.HTTPServer.Timeout,
		Addr:              cfg.HTTPServer.Addr,
		Handler:           mux,
	}
	go func() {
		err := s.ListenAndServe()
		if err != nil {
			log.Fatal().Err(err).Send()
		}
	}()

	log.Info().Msg("Mialer service started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if err := rabbitMQConn.Close(); err != nil {
		log.Error().Err(err).Msg("Cannot close RabbitMQ connection")
	}
	c.Stop()
}
