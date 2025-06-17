package app

import (
	"github.com/gitslim/go-ragger/internal/config"
	"github.com/gitslim/go-ragger/internal/logger"
	"github.com/gitslim/go-ragger/internal/version"
	"github.com/gitslim/go-ragger/internal/web"
	"go.uber.org/fx"
)

func CreateServerApp() fx.Option {
	return fx.Options(
		fx.NopLogger,
		version.Module,
		logger.Module,
		config.Module,
		web.Module,
	)
}

// RunServerApp запускает приложение сервера
func RunServerApp() {
	fx.New(CreateServerApp()).Run()
}
