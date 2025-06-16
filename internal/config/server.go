package config

import (
	"fmt"
	"log/slog"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	ServerAddress string `mapstructure:"server_address"`
	DSN           string `mapstructure:"dsn"`
}

func NewServerConfig() (*ServerConfig, error) {
	var cfg ServerConfig

	viper.SetConfigName("server")
	viper.AddConfigPath("./configs/")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("viper.ReadInConfig: %s", err)
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("viper.Unmarshal: %s", err)
	}

	slog.Info("Server config loaded", "config", cfg)

	return &cfg, nil
}
