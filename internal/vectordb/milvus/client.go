package milvus

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gitslim/go-ragger/internal/config"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"go.uber.org/fx"
)

// NewMilvusClient creates a new Milvus client
func NewMilvusClient(lc fx.Lifecycle, cfg *config.ServerConfig, logger *slog.Logger) (*client.Client, error) {
	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			cancel()
			return nil
		},
	})

	logger.Info("creating milvus client", "address", cfg.MilvusAddress)

	cli, err := client.NewClient(ctx, client.Config{
		Address:  cfg.MilvusAddress,
		Username: cfg.MilvusUsername,
		Password: cfg.MilvusPassword,
	})

	if err != nil {
		return nil, fmt.Errorf("Failed to create milvus client: %w", err)
	}

	return &cli, nil
}
