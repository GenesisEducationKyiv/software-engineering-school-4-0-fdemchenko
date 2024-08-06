package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/fdemchenko/exchanger/cmd/web/internal/messaging"
	"github.com/fdemchenko/exchanger/internal/communication/customers"
	"github.com/fdemchenko/exchanger/internal/communication/mailer"
	"github.com/fdemchenko/exchanger/internal/communication/rabbitmq"
	"github.com/fdemchenko/exchanger/internal/database"
	"github.com/fdemchenko/exchanger/internal/repositories"
	"github.com/fdemchenko/exchanger/internal/services"
	"github.com/fdemchenko/exchanger/internal/services/rate"
	"github.com/fdemchenko/exchanger/migrations"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type config struct {
	addr string
	db   struct {
		dsn            string
		maxConnections int
	}
	mailerUpdateInterval time.Duration
	rabbitMQConnString   string
}

type RateService interface {
	GetRate(context.Context, string) (float32, error)
}

type EmailService interface {
	Create(email string) (int, error)
	GetAll() ([]string, error)
	DeleteByEmail(email string) error
	DeleteByID(id int) error
}

type application struct {
	cfg              config
	rateService      RateService
	emailService     EmailService
	customerProducer *rabbitmq.GenericProducer
}

const (
	ServerTimeout           = 10 * time.Second
	DefaultMaxDBConnections = 25
	DefaultMailerInterval   = 24 * time.Hour
	RateCachingDuration     = 15 * time.Minute
)

func main() {
	cfg := initConfig()

	zerolog.TimeFieldFormat = time.RFC3339

	db, err := database.OpenDB(cfg.db.dsn, database.Options{MaxOpenConnections: cfg.db.maxConnections})
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	log.Info().Msg("Coonected to DB successfully")

	err = database.AutoMigrate(db, migrations.RatesMigrationsFS, "rates", "exchanger", false)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	rabbitMQConn, err := amqp.Dial(cfg.rabbitMQConnString)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	log.Info().Msg("Coonected to RabbitMQ successfully")

	createCustomersChannel, err := rabbitmq.OpenWithQueueName(rabbitMQConn, customers.CreateCustomerRequestQueue)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	customersProducer := rabbitmq.NewGenericProducer(createCustomersChannel)
	subscriptionRepository := &repositories.PostgresSubscriptionRepository{DB: db}
	emailService := services.NewSubscriptionService(subscriptionRepository)
	rateService := rate.NewRateService(
		rate.WithFetchers(
			rate.NewNBURateFetcher("nbu fetcher"),
			rate.NewFawazRateFetcher("fawaz fetcher"),
			rate.NewPrivatRateFetcher("privat fetcher"),
		),
		rate.WithUpdateInterval(RateCachingDuration),
	)

	checkCustomersCreationChannel, err := rabbitmq.OpenWithQueueName(
		rabbitMQConn,
		customers.CreateCustomerResponseQueue,
	)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	customersSAGAConsumer := messaging.NewCustomerCreationSAGAConsumer(
		checkCustomersCreationChannel,
		subscriptionRepository,
	)

	err = customersSAGAConsumer.StartListening()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	rateEmailsChannel, err := rabbitmq.OpenWithQueueName(rabbitMQConn, mailer.RateEmailsQueue)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	emailsSender := services.NewRabbitMQEmailSender(emailService, rateService, rateEmailsChannel)
	triggerConsumer := messaging.NewEmailTriggerConsumer(rateEmailsChannel, emailsSender)
	err = triggerConsumer.StartListening()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	app := application{
		cfg:              cfg,
		rateService:      rateService,
		emailService:     emailService,
		customerProducer: customersProducer,
	}

	log.Info().Str("address", app.cfg.addr).Msg("Web server started")
	err = app.serveHTTP()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	if err := rabbitMQConn.Close(); err != nil {
		log.Error().Err(err).Msg("Cannot close RabbitMQ connection")
	}

	if err := db.Close(); err != nil {
		log.Error().Err(err).Msg("Cannot close DB connection")
	}
}

func initConfig() config {
	var cfg config
	cfg.mailerUpdateInterval = DefaultMailerInterval
	flag.StringVar(&cfg.addr, "addr", ":8080", "http listen address")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("EXCHANGER_DSN"), "Data source name")
	flag.IntVar(&cfg.db.maxConnections, "db-max-conn", DefaultMaxDBConnections, "Database max connection")
	flag.StringVar(&cfg.rabbitMQConnString,
		"rabbitmq-conn-string",
		os.Getenv("EXCHANGER_RABBITMQ_CONN_STRING"),
		"RabbitMQ connection string",
	)
	flag.Parse()
	return cfg
}
