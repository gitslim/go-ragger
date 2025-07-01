package chunkr

import (
	"github.com/gitslim/chunkr-ai-sdk/sdk/go/chunkrai/client"
	"github.com/gitslim/chunkr-ai-sdk/sdk/go/chunkrai/option"
	"github.com/gitslim/go-ragger/internal/config"
)

type Client struct {
	client.Client
}

func NewClient(config *config.ServerConfig) *Client {
	client := client.NewClient(
		option.WithBaseURL(config.ChunkrURL),
		option.WithApiKey(config.ChunkrAPIKey),
		option.WithMaxAttempts(3))

	return &Client{Client: *client}
}
