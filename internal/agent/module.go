package agent

import (
	"go.uber.org/fx"
)

var ModuleAgentFactory = fx.Module("agent-factory",
	fx.Provide(
		NewRAGAgentFactory,
	),
)
