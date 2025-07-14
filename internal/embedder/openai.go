package embedder

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/gitslim/go-ragger/internal/config"
)

// NewOpenAIEmbedder creates a new OpenAI embedder.
func NewOpenAIEmbedder(cfg *config.ServerConfig) (*openai.Embedder, error) {
	ctx := context.Background()
	// defaultDim := 33792 // 22528 //1536

	emb, err := openai.NewEmbedder(ctx, &openai.EmbeddingConfig{
		BaseURL: cfg.OpenAIBaseURL,
		APIKey:  cfg.OpenAIAPIKey,
		Model:   cfg.EmbeddingModel,
		// Dimensions: &defaultDim,
		Timeout: 60 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create embedder: %w", err)
	}

	return emb, nil
}
