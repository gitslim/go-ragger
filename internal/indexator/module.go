package indexator

import (
	"go.uber.org/fx"
)

var ModuleIndexator = fx.Module("indexator",
	fx.Invoke(RunIndexator),
)
