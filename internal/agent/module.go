package agent

import (
	"go.uber.org/fx"
)

// ModuleAgentFactory is a fx module for the agent factory.
var ModuleAgentFactory = fx.Module("agent-factory",
	fx.Provide(
		NewRAGAgentFactory,
	),
)
