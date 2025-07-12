package db

import (
	"go.uber.org/fx"
)

// Module is the fx module for the database.
var Module = fx.Module("db",
	fx.Provide(
		NewPgxPool,
		NewDb,
	),
	fx.Invoke(RegisterDBPoolHooks),
)
