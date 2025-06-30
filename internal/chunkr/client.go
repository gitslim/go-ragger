package chunkr

import (
	"context"
	"fmt"

	"github.com/gitslim/go-ragger/internal/chunkr/api"
	"github.com/gitslim/go-ragger/internal/config"
)

type Client struct {
	api.Client
}

type sec struct {
	apiKey string
}

func (s *sec) APIKey(ctx context.Context, operationName api.OperationName) (api.APIKey, error) {
	if s.apiKey == "" {
		return api.APIKey{}, fmt.Errorf("API key is empty")
	}
	return api.APIKey{
		APIKey: s.apiKey,
	}, nil
}

func NewClient(config *config.ServerConfig) (*Client, error) {
	securitySource := &sec{apiKey: config.ChunkrURL}

	client, err := api.NewClient(config.ChunkrURL, securitySource)
	if err != nil {
		return &Client{}, err
	}

	return &Client{Client: *client}, nil

}
