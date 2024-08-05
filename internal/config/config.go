package config

import (
	"errors"
	"flag"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

var (
	ErrConfigFileNotExist = errors.New("config file path does not exist")
)

func Load[T any](defaultConfigPath string) (*T, error) {
	configPath, err := fetchConfigPath(defaultConfigPath)
	if err != nil {
		return nil, err
	}

	var cfg T
	err = cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func MustLoad[T any](defaultConfigPath string) *T {
	cfg, err := Load[T](defaultConfigPath)
	if err != nil {
		panic(err)
	}
	return cfg
}

func fetchConfigPath(defaultConfigPath string) (string, error) {
	var configPath string
	flag.StringVar(&configPath, "config-path", defaultConfigPath, "Config file path")
	flag.Parse()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", ErrConfigFileNotExist
	}

	return configPath, nil
}
