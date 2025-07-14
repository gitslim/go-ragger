package chunker

import (
	"go.uber.org/fx"
)

// Module is the fx module for the chunker package.
var Module = fx.Module("chunker",
	fx.Provide(
		NewChunkrAIClient,
	),
	fx.Invoke(RunChunker),
)
