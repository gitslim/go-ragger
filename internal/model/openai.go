package model

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/gitslim/go-ragger/internal/config"
)

func NewOpenAIChatModel(cfg *config.ServerConfig) (model.ToolCallingChatModel, error) {
	ctx := context.Background()

	config := &openai.ChatModelConfig{
		BaseURL: cfg.OpenAIBaseURL,
		Model:   cfg.ChatModel,
		APIKey:  cfg.OpenAIAPIKey,
	}

	cm, err := openai.NewChatModel(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create openai chat model: %w", err)
	}
	return cm, nil
}
