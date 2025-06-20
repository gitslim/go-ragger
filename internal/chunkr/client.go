package chunkr

import (
	"context"
	"fmt"

	"github.com/gitslim/go-ragger/internal/chunkr/api"
	"github.com/gitslim/go-ragger/internal/config"
)

type Sec struct {
	apiKey string
}

func (s *Sec) APIKey(ctx context.Context, operationName api.OperationName) (api.APIKey, error) {
	if s.apiKey == "" {
		return api.APIKey{}, fmt.Errorf("API key is empty")
	}
	return api.APIKey{
		APIKey: s.apiKey,
	}, nil
}

func NewClient(config *config.ServerConfig) (api.Client, error) {
	securitySource := &Sec{apiKey: config.ChunkrURL}

	client, err := api.NewClient(config.ChunkrURL, securitySource)
	if err != nil {
		return api.Client{}, err
	}

	return *client, nil
}
