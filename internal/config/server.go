package config

import (
	"fmt"

	"github.com/caarlos0/env/v6"
	"github.com/spf13/pflag"
)

type ServerConfig struct {
	ServerAddress string `env:"SERVER_ADDRESS" envDefault:"127.0.0.1:8080"`
	DSN           string `env:"DSN" envDefault:"postgres://postgres:postgres@localhost:5432/ragger?sslmode=disable"`
	Debug         bool   `env:"DEBUG" envDefault:"false"`
}

func NewServerConfig() (*ServerConfig, error) {
	var cfg ServerConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("env.Parse: %w", err)
	}

	pflag.String("server_address", cfg.ServerAddress, "Address to listen on")
	pflag.String("dsn", cfg.DSN, "Database DSN")
	pflag.Bool("debug", cfg.Debug, "Show debug logs")

	pflag.Parse()

	if val, err := pflag.CommandLine.GetString("server_address"); err == nil && val != "" {
		cfg.ServerAddress = val
	}
	if val, err := pflag.CommandLine.GetString("dsn"); err == nil && val != "" {
		cfg.DSN = val
	}
	if val, err := pflag.CommandLine.GetBool("debug"); err == nil {
		cfg.Debug = val
	}

	return &cfg, nil
}
