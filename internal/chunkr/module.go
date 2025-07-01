package chunkr

import (
	"go.uber.org/fx"
)

var Module = fx.Module("chunkr",
	fx.Provide(
		NewClient,
	),
	fx.Invoke(RunDocumentProcessor),
)
