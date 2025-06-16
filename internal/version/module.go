package version

import (
	"go.uber.org/fx"
)

var Module = fx.Module("version",
	fx.Provide(NewVersion),
	fx.Invoke(RegisterVersionHooks),
)
