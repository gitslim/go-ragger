package app

import (
	"github.com/gitslim/go-ragger/internal/config"
	"github.com/gitslim/go-ragger/internal/logger"
	"github.com/gitslim/go-ragger/internal/version"
	"go.uber.org/fx"
)

func CreateServerApp() fx.Option {
	return fx.Options(
		fx.NopLogger,
		version.Module,
		logger.Module,
		config.Module,
	)
}

// RunServerApp запускает приложение сервера
func RunServerApp() {
	fx.New(CreateServerApp()).Run()
}
