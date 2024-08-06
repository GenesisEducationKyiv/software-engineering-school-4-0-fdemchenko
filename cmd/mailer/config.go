package main

import (
	"flag"
	"os"
)

type Config struct {
	SMTP               SMTPConfig
	RabbitMQConnString string
}

type SMTPConfig struct {
	Host               string
	Username           string
	Port               int
	Password           string
	Sender             string
	ConnectionPoolSize int
}

var ServiceConfig Config

func init() {
	flag.StringVar(&ServiceConfig.SMTP.Host, "smtp-host", os.Getenv("EXCHANGER_SMTP_HOST"), "Smtp host")
	flag.IntVar(&ServiceConfig.SMTP.Port, "smtp-port", DefaultSMTPPort, "Smtp port")
	flag.IntVar(&ServiceConfig.SMTP.ConnectionPoolSize,
		"smtp-connections",
		DefaultMailerConnectionPoolSize,
		"Smtp connection pool size",
	)
	flag.StringVar(&ServiceConfig.SMTP.Username, "smtp-username", os.Getenv("EXCHANGER_SMTP_USERNAME"), "Smtp username")
	flag.StringVar(&ServiceConfig.SMTP.Password, "smtp-password", os.Getenv("EXCHANGER_SMTP_PASSWORD"), "Smtp password")
	flag.StringVar(&ServiceConfig.SMTP.Sender, "smtp-sender", os.Getenv("EXCHANGER_SMTP_SENDER"), "Smtp sender")

	flag.StringVar(&ServiceConfig.RabbitMQConnString,
		"rabbitmq-conn-string",
		os.Getenv("EXCHANGER_RABBITMQ_CONN_STRING"),
		"RabbitMQ connection string",
	)
	flag.Parse()
}
