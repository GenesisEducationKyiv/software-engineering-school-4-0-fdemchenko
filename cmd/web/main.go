package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"time"

	"github.com/fdemchenko/exchanger/internal/database"
	"github.com/fdemchenko/exchanger/internal/repositories"
	"github.com/fdemchenko/exchanger/internal/services"
	"github.com/fdemchenko/exchanger/internal/services/mailer"
	"github.com/fdemchenko/exchanger/internal/services/rate"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type config struct {
	addr string
	db   struct {
		dsn            string
		maxConnections int
	}
	mailer mailer.MailerConfig
}

type RateService interface {
	GetRate(context.Context, string) (float32, error)
}

type EmailService interface {
	Create(email string) error
	GetAll() ([]string, error)
}

type application struct {
	cfg          config
	rateService  RateService
	emailService EmailService
}

const (
	DefaultSMTPPort                 = 25
	ServerTimeout                   = 10 * time.Second
	DefaultMaxDBConnections         = 25
	DefaultMailerInterval           = 24 * time.Hour
	RateCachingDuration             = 15 * time.Minute
	DefaultMailerConnectionPoolSize = 3
)

func main() {
	cfg := initConfig()

	zerolog.TimeFieldFormat = time.RFC3339

	db, err := openDB(cfg)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	log.Info().Msg("Coonected to DB successfully")

	err = database.AutoMigrate(db, false)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	emailRepository := &repositories.PostgresEmailRepository{DB: db}
	emailService := services.NewEmailService(emailRepository)
	rateService := rate.NewRateService(
		rate.WithFetchers(
			rate.NewNBURateFetcher("nbu fetcher"),
			rate.NewFawazRateFetcher("fawaz fetcher"),
			rate.NewPrivatRateFetcher("privat fetcher"),
		),
		rate.WithUpdateInterval(RateCachingDuration),
	)

	mailerService := mailer.NewMailerService(cfg.mailer, emailService, rateService)
	mailerService.StartEmailSending(cfg.mailer.UpdateInterval)

	app := application{
		cfg:          cfg,
		rateService:  rateService,
		emailService: emailService,
	}

	log.Info().Str("address", app.cfg.addr).Msg("Web server started")
	err = app.serveHTTP()
	if err != nil {
		log.Fatal().Err(err).Send()
	}
}

func initConfig() config {
	var cfg config
	cfg.mailer.UpdateInterval = DefaultMailerInterval
	flag.StringVar(&cfg.addr, "addr", ":8080", "http listen address")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("EXCHANGER_DSN"), "Data source name")
	flag.IntVar(&cfg.db.maxConnections, "db-max-conn", DefaultMaxDBConnections, "Database max connection")

	flag.StringVar(&cfg.mailer.Host, "smtp-host", os.Getenv("EXCHANGER_SMPT_HOST"), "Smpt host")
	flag.IntVar(&cfg.mailer.Port, "smtp-port", DefaultSMTPPort, "Smpt port")
	flag.StringVar(&cfg.mailer.Username, "smtp-username", os.Getenv("EXCHANGER_SMPT_USERNAME"), "Smpt username")
	flag.StringVar(&cfg.mailer.Password, "smtp-password", os.Getenv("EXCHANGER_SMPT_PASSWORD"), "Smpt password")
	flag.StringVar(&cfg.mailer.Sender, "smtp-sender", os.Getenv("EXCHANGER_SMPT_SENDER"), "Smpt sender")
	flag.IntVar(&cfg.mailer.ConnectionPoolSize,
		"mailer-connections",
		DefaultMailerConnectionPoolSize,
		"Mailer connection pool size",
	)
	flag.Func("mailer-interval", "Email update interval (E.g. 24h, 1h30m)", func(s string) error {
		if s == "" {
			cfg.mailer.UpdateInterval = DefaultMailerInterval
			return nil
		}
		duration, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		cfg.mailer.UpdateInterval = duration
		return nil
	})
	flag.Parse()
	return cfg
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.db.maxConnections)
	return db, nil
}
