package config

import (
	"time"
)

type Config struct {
	SMTPServer         SMTPServer `yaml:"smtp-server"`
	HTTPServer         HTTPServer `yaml:"http-server"`
	Env                string     `yaml:"env"`
	RabbitMQConnString string     `yaml:"rabbitmq-addr" env-required:"true"`
	SchedulerCRON      string     `yaml:"scheduler-cron" env-default:"0 0 10 * * *"`
}

type HTTPServer struct {
	Addr    string        `yaml:"addr"`
	Timeout time.Duration `yaml:"timeout" env-default:"5s"`
}

type SMTPServer struct {
	Host               string `yaml:"host" env-required:"true"`
	Username           string `yaml:"username" env-required:"true"`
	Port               int    `yaml:"port" env-default:"25"`
	Password           string `yaml:"password" env-requiered:"true"`
	Sender             string `yaml:"sender" env-required:"true"`
	ConnectionPoolSize int    `yaml:"connection-pool-size" env-default:"3"`
}

const (
	DefaultRabbitMQPort      = 5672
	DefaultSchedulerInterval = 24 * time.Hour
)
