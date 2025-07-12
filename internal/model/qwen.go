package model

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/qwen"

	"github.com/cloudwego/eino/components/model"
	"github.com/gitslim/go-ragger/internal/config"
	"github.com/gitslim/go-ragger/internal/util"
)

func NewQwenChatModel(cfg *config.ServerConfig) (model.ToolCallingChatModel, error) {
	ctx := context.Background()
	enableThinking := true

	config := &qwen.ChatModelConfig{
		BaseURL:        cfg.OpenAIBaseURL,
		Model:          cfg.ChatModel,
		APIKey:         cfg.OpenAIAPIKey,
		Timeout:        0,
		EnableThinking: &enableThinking,
		// MaxTokens:      util.Of(2048),
		Temperature: util.Of(float32(0.7)),
		TopP:        util.Of(float32(0.7)),
	}

	cm, err := qwen.NewChatModel(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create openai chat model: %w", err)
	}
	return cm, nil
}
