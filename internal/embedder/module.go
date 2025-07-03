package embedder

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/cloudwego/eino/components/embedding"
	"go.uber.org/fx"
)

var ModuleOpenAIEmbedder = fx.Module("openai-embedder",
	fx.Provide(
		fx.Annotate(
			NewOpenAIEmbedder,
			fx.As(new(embedding.Embedder)),
		),
	),
	fx.Invoke(CheckEmbedder),
)

func CheckEmbedder(lc fx.Lifecycle, logger *slog.Logger, emb embedding.Embedder) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			testStrings := []string{"foo", "bar"}

			_, err := emb.EmbedStrings(ctx, testStrings)
			if err != nil {
				logger.Error("embedder check failed", "error", err)
				return fmt.Errorf("embedder test failed: %w", err)
			}
			logger.Info("embedder check success")

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}
