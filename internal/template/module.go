package template

import (
	"github.com/cloudwego/eino/components/prompt"
	"go.uber.org/fx"
)

var ModuleChatTemplate = fx.Module("chat-template",
	fx.Provide(
		fx.Annotate(
			NewRAGChatTemplate,
			fx.As(new(prompt.ChatTemplate))),
	),
)
