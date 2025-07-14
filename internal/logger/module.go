package logger

import (
	"go.uber.org/fx"
)

// Module for logger.
var Module = fx.Module("logger",
	fx.Provide(
		NewLogger,
	),
)
