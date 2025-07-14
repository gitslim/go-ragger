package version

import (
	"go.uber.org/fx"
)

// Module is the fx module for the version package
var Module = fx.Module("version",
	fx.Provide(NewVersion),
	fx.Invoke(RegisterVersionHooks),
)
