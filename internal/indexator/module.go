package indexator

import (
	"go.uber.org/fx"
)

// ModuleIndexator is a module for indexator package
var ModuleIndexator = fx.Module("indexator",
	fx.Invoke(RunIndexator),
)
