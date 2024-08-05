package main

import (
	"context"
	"os"
	"time"

	webconfig "github.com/fdemchenko/exchanger/cmd/web/internal/config"
	"github.com/fdemchenko/exchanger/cmd/web/internal/messaging"
	"github.com/fdemchenko/exchanger/cmd/web/internal/repositories"
	"github.com/fdemchenko/exchanger/cmd/web/internal/services"
	"github.com/fdemchenko/exchanger/cmd/web/internal/services/rate"
	"github.com/fdemchenko/exchanger/internal/communication/customers"
	"github.com/fdemchenko/exchanger/internal/communication/mailer"
	"github.com/fdemchenko/exchanger/internal/communication/rabbitmq"
	"github.com/fdemchenko/exchanger/internal/config"
	"github.com/fdemchenko/exchanger/internal/database"
	"github.com/fdemchenko/exchanger/migrations"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

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
	cfg              *webconfig.Config
	rateService      RateService
	emailService     EmailService
	customerProducer *rabbitmq.GenericProducer
}

func main() {
	cfg := config.MustLoad[webconfig.Config](os.Getenv("WEB_CONFIG_PATH"))
	zerolog.TimeFieldFormat = time.RFC3339
	db, err := database.OpenDB(cfg.DB.DSN, database.Options{MaxOpenConnections: cfg.DB.MaxConnections})
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	log.Info().Msg("Coonected to DB successfully")

	err = database.AutoMigrate(db, migrations.RatesMigrationsFS, "rates", "exchanger", false)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	log.Info().Msg("Migrations successfully applied")

	rabbitMQConn, err := amqp.Dial(cfg.RabbitMQConnString)
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
		rate.WithUpdateInterval(cfg.RateCacheTTL),
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

	log.Info().Str("address", cfg.HTTPServer.Addr).Msg("Web server started")
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
