package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/components/model"
	"github.com/gitslim/go-ragger/internal/config"
)

func NewOllamaChatModel(cfg *config.ServerConfig) (model.ToolCallingChatModel, error) {
	ctx := context.Background()

	config := &ollama.ChatModelConfig{
		BaseURL: strings.Replace(cfg.OpenAIBaseURL, "/v1", "", 1), // TODO: remove hardcoding
		Model:   cfg.ChatModel,
		Timeout: time.Second * 30,
	}

	cm, err := ollama.NewChatModel(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create ollama chat model: %w", err)
	}
	return cm, nil
}
