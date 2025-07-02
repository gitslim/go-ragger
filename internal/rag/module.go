package rag

import (
	"go.uber.org/fx"
)

var Module = fx.Module("rag",
	fx.Provide(),
	fx.Invoke(RunRAGPipeline),
)
