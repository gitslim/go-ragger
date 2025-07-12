package model

import (
	"github.com/cloudwego/eino/components/model"
	"go.uber.org/fx"
)

// ModuleModel is the fx module for llm model
var ModuleModel = fx.Module("model",
	fx.Provide(
		fx.Annotate(
			NewQwenChatModel,
			fx.As(new(model.ToolCallingChatModel))),
	),
)
