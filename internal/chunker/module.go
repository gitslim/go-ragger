package chunker

import (
	"go.uber.org/fx"
)

var Module = fx.Module("chunker",
	fx.Provide(
		NewChunkrAIClient,
	),
	fx.Invoke(RunChunker),
)
