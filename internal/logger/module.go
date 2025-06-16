package logger

import (
	"log/slog"

	"github.com/robbert229/fxslog"
	"go.uber.org/fx"
)

var Module = fx.Module("logger",
	// slog for uber.fx
	fxslog.WithLogger(),

	fx.Provide(
		// Дефолтный slog логгер.
		func() *slog.Logger {
			return slog.Default()
		},
	))
