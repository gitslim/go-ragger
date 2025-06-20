package logger

import (
	"log/slog"
	"os"

	"github.com/gitslim/go-ragger/internal/config"
)

func NewLogger(cfg *config.ServerConfig) *slog.Logger {
	lvl := new(slog.LevelVar)
	if cfg.Debug {
		lvl.Set(slog.LevelDebug)
	} else {
		lvl.Set(slog.LevelInfo)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: lvl,
	}))

	slog.SetDefault(logger)

	return logger
}
