package model

import (
	"github.com/cloudwego/eino/components/model"
	"go.uber.org/fx"
)

var ModuleModel = fx.Module("model",
	fx.Provide(
		fx.Annotate(
			NewQwenChatModel,
			fx.As(new(model.ToolCallingChatModel))),
	),
)
