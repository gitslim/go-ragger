package template

import (
	"github.com/cloudwego/eino/components/prompt"
	"go.uber.org/fx"
)

// ModuleChatTemplate is a fx module for chat template
var ModuleChatTemplate = fx.Module("chat-template",
	fx.Provide(
		fx.Annotate(
			NewRAGChatTemplate,
			fx.As(new(prompt.ChatTemplate))),
	),
)
