package chunkr

import (
	"github.com/gitslim/go-ragger/internal/chunkr/api"
	"go.uber.org/fx"
)

var Module = fx.Module("chunkr",
	fx.Provide(
		api.NewClient,
	),
)
