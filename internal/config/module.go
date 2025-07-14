package config

import "go.uber.org/fx"

// Module is the fx module for the config package.
var Module = fx.Module("config",
	fx.Provide(
		NewServerConfig,
	),
)
