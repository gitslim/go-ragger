package web

import "go.uber.org/fx"

var Module = fx.Module("web",
	fx.Invoke(RegisterHTTPServerHooks),
)
