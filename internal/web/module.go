package web

import "go.uber.org/fx"

// Module is the fx module for the web package
var Module = fx.Module("web",
	fx.Invoke(RegisterHTTPServerHooks),
)
