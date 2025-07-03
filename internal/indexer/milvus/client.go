package milvus

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gitslim/go-ragger/internal/config"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"go.uber.org/fx"
)

func NewMilvusClient(cfg *config.ServerConfig, logger *slog.Logger) (*client.Client, error) {
	ctx := context.Background()

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

func RunMilvusClient(lc fx.Lifecycle, logger *slog.Logger, cli *client.Client) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return nil
		},
		OnStop: func(ctx context.Context) error {
			err := (*cli).Close()
			if err != nil {
				return fmt.Errorf("failed to stop milvus client: %w", err)
			}
			return nil
		},
	})
}
