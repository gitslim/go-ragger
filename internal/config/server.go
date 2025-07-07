package config

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
)

type ServerConfig struct {
	ServerAddress  string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	DSN            string `env:"DSN" envDefault:"postgres://postgres:postgres@localhost:5432/ragger?sslmode=disable"`
	Debug          bool   `env:"DEBUG" envDefault:"false"`
	ChunkrURL      string `env:"CHUNKR_URL" envDefault:"localhost:8888"`
	ChunkrAPIKey   string `env:"CHUNKR_API_KEY" envDefault:""`
	OpenAIBaseURL  string `env:"OPENAI_BASE_URL" envDefault:"http://localhost:11434/v1"`
	OpenAIAPIKey   string `env:"OPENAI_API_KEY" envDefault:"dummy-key"`
	EmbeddingModel string `env:"EMBEDDING_MODEL" envDefault:"myaniu/qwen3-embedding:0.6b"`
	ChatModel      string `env:"CHAT_MODEL" envDefault:"deepseek-r1:8b"`
	MilvusAddress  string `env:"MILVUS_ADDRESS" envDefault:"localhost:19530"`
	MilvusUsername string `env:"MILVUS_USERNAME" envDefault:"root"`
	MilvusPassword string `env:"MILVUS_PASSWORD" envDefault:"Milvus"`
}

func NewServerConfig() (*ServerConfig, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	var cfg ServerConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("env.Parse: %w", err)
	}

	pflag.String("server_address", cfg.ServerAddress, "Address to listen on")
	pflag.String("dsn", cfg.DSN, "Database DSN")
	pflag.Bool("debug", cfg.Debug, "Show debug logs")
	pflag.String("chunkr_url", cfg.ChunkrURL, "Chunkr api server url")
	pflag.String("chunkr_api_key", cfg.ChunkrAPIKey, "Chunkr api key")
	pflag.String("openai_base_url", cfg.OpenAIBaseURL, "OpenAI compatible api base url")
	pflag.String("openai_api_key", cfg.OpenAIAPIKey, "OpenAI api key")
	pflag.String("embedding_model", cfg.EmbeddingModel, "Text embedding LLM model")
	pflag.String("chat_model", cfg.ChatModel, "Chat LLM model")
	pflag.String("milvus_address", cfg.MilvusAddress, "Milvus server address")
	pflag.String("milvus_username", cfg.MilvusUsername, "Milvus user name")
	pflag.String("milvus_password", cfg.MilvusPassword, "Milvus user password")

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
	if val, err := pflag.CommandLine.GetString("openai_base_url"); err == nil && val != "" {
		cfg.OpenAIBaseURL = val
	}
	if val, err := pflag.CommandLine.GetString("openai_api_key"); err == nil && val != "" {
		cfg.OpenAIAPIKey = val
	}
	if val, err := pflag.CommandLine.GetString("embedding_model"); err == nil && val != "" {
		cfg.EmbeddingModel = val
	}
	if val, err := pflag.CommandLine.GetString("chat_model"); err == nil && val != "" {
		cfg.ChatModel = val
	}
	if val, err := pflag.CommandLine.GetString("milvus_address"); err == nil && val != "" {
		cfg.MilvusAddress = val
	}
	if val, err := pflag.CommandLine.GetString("milvus_username"); err == nil && val != "" {
		cfg.MilvusUsername = val
	}
	if val, err := pflag.CommandLine.GetString("milvus_password"); err == nil && val != "" {
		cfg.MilvusPassword = val
	}

	return &cfg, nil
}
