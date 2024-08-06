package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fdemchenko/exchanger/cmd/customers/internal/data"
	"github.com/fdemchenko/exchanger/cmd/customers/internal/messaging"
	"github.com/fdemchenko/exchanger/internal/communication/customers"
	"github.com/fdemchenko/exchanger/internal/communication/rabbitmq"
	"github.com/fdemchenko/exchanger/internal/database"
	"github.com/fdemchenko/exchanger/migrations"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type config struct {
	db struct {
		dsn                string
		maxOpenConnections int
	}
	rabbitMQConnString string
}

const DefaultMaxDBConnections = 10

func main() {
	var cfg config
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("EXCHANGER_CUSTOMERS_DSN"), "Data source name")
	flag.IntVar(&cfg.db.maxOpenConnections, "db-max-conn", DefaultMaxDBConnections, "Database max connection")
	flag.StringVar(&cfg.rabbitMQConnString,
		"rabbitmq-conn-string",
		os.Getenv("EXCHANGER_RABBITMQ_CONN_STRING"),
		"RabbitMQ connection string",
	)

	zerolog.TimeFieldFormat = time.RFC3339
	db, err := database.OpenDB(cfg.db.dsn, database.Options{MaxOpenConnections: cfg.db.maxOpenConnections})
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	log.Info().Msg("Coonected to DB successfully")

	err = database.AutoMigrate(db, migrations.CustomersMigrationsFS, "customers", "customers_service", false)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	rabbitMQConn, err := amqp.Dial(cfg.rabbitMQConnString)
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
