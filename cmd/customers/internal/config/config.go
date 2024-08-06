package config

import "time"

type Config struct {
	HTTPServer         HTTPServer `yaml:"http-server"`
	DB                 DataStore  `yaml:"data-store"`
	RabbitMQConnString string     `yaml:"rabbitmq-addr" env:"CUSTOMERS_RABBITMQ_CONN_STR" env-required:"true"`
	Env                string     `yaml:"env" env-default:"local"`
}

type HTTPServer struct {
	Addr    string        `yaml:"addr"`
	Timeout time.Duration `yaml:"timeout" env-default:"10s"`
}

type DataStore struct {
	DSN            string `yaml:"dsn" env:"CUSTOMERS_DSN" env-required:"true"`
	MaxConnections int    `yaml:"max-connections" env-default:"25"`
}
