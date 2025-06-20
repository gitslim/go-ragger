package config

import (
	"fmt"

	"github.com/caarlos0/env/v6"
	"github.com/spf13/pflag"
)

type ServerConfig struct {
	ServerAddress string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	DSN           string `env:"DSN" envDefault:"postgres://postgres:postgres@localhost:5432/ragger?sslmode=disable"`
	Debug         bool   `env:"DEBUG" envDefault:"false"`
	ChunkrURL     string `env:"CHUNKR_URL" envDefault:"localhost:8888"`
	ChunkrAPIKey  string `env:"CHUNKR_API_KEY" envDefault:""`
}

func NewServerConfig() (*ServerConfig, error) {
	var cfg ServerConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("env.Parse: %w", err)
	}

	pflag.String("server_address", cfg.ServerAddress, "Address to listen on")
	pflag.String("dsn", cfg.DSN, "Database DSN")
	pflag.Bool("debug", cfg.Debug, "Show debug logs")
	pflag.String("chunkr_url", cfg.ChunkrURL, "Chunkr api server url")
	pflag.String("chunkr_api_key", cfg.ChunkrAPIKey, "Chunkr api key")

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
	if val, err := pflag.CommandLine.GetString("chunkr_url"); err == nil && val != "" {
		cfg.ChunkrURL = val
	}
	if val, err := pflag.CommandLine.GetString("chunkr_api_key"); err == nil && val != "" {
		cfg.ChunkrAPIKey = val
	}

	return &cfg, nil
}
