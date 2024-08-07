package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/VictoriaMetrics/metrics"
	customersconfig "github.com/fdemchenko/exchanger/cmd/customers/internal/config"
	"github.com/fdemchenko/exchanger/cmd/customers/internal/data"
	"github.com/fdemchenko/exchanger/cmd/customers/internal/messaging"
	"github.com/fdemchenko/exchanger/internal/communication/customers"
	"github.com/fdemchenko/exchanger/internal/communication/rabbitmq"
	"github.com/fdemchenko/exchanger/internal/config"
	"github.com/fdemchenko/exchanger/internal/database"
	"github.com/fdemchenko/exchanger/migrations"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config.MustLoad[customersconfig.Config](os.Getenv("CUSTOMERS_CONFIG_PATH"))

	zerolog.TimeFieldFormat = time.RFC3339
	db, err := database.OpenDB(cfg.DB.DSN, database.Options{MaxOpenConnections: cfg.DB.MaxConnections})
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	log.Info().Msg("Coonected to DB successfully")

	err = database.AutoMigrate(db, migrations.CustomersMigrationsFS, "customers", "customers_service", false)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	log.Info().Msg("Migrations successfully applied")

	rabbitMQConn, err := amqp.Dial(cfg.RabbitMQConnString)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	log.Info().Msg("Coonected to RabbitMQ successfully")

	requestsChannel, err := rabbitmq.OpenWithQueueName(rabbitMQConn, customers.CreateCustomerRequestQueue)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	responcesChannel, err := rabbitmq.OpenWithQueueName(rabbitMQConn, customers.CreateCustomerResponseQueue)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	customersRepository := &data.CustomerPostgreSQLRepository{DB: db}
	producer := rabbitmq.NewGenericProducer(responcesChannel)
	consumer := messaging.NewCustomerCreationConsumer(requestsChannel, customersRepository, producer)

	log.Info().Msg("Mialer service started")
	err = consumer.StartListening()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics.WritePrometheus(w, false)
	})
	s := http.Server{
		Addr:              cfg.HTTPServer.Addr,
		Handler:           mux,
		ReadHeaderTimeout: cfg.HTTPServer.Timeout,
	}
	go func() {
		err := s.ListenAndServe()
		if err != nil {
			log.Fatal().Err(err).Send()
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if err := rabbitMQConn.Close(); err != nil {
		log.Error().Err(err).Msg("Cannot close RabbitMQ connection")
	}

	if err := db.Close(); err != nil {
		log.Error().Err(err).Msg("Cannot close DB connection")
	}
}
