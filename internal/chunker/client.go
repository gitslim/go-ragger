package chunker

import (
	"github.com/gitslim/chunkr-ai-sdk/sdk/go/chunkrai/client"
	"github.com/gitslim/chunkr-ai-sdk/sdk/go/chunkrai/option"
	"github.com/gitslim/go-ragger/internal/config"
)

// ChunkrAIClient is a client for the Chunkr AI API.
type ChunkrAIClient struct {
	client.Client
}

// NewChunkrAIClient creates a new ChunkrAIClient instance.
func NewChunkrAIClient(config *config.ServerConfig) *ChunkrAIClient {
	client := client.NewClient(
		option.WithBaseURL(config.ChunkrURL),
		option.WithApiKey(config.ChunkrAPIKey),
		option.WithMaxAttempts(3))

	return &ChunkrAIClient{Client: *client}
}
