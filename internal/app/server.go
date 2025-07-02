package app

import (
	"encoding/gob"

	"github.com/gitslim/go-ragger/internal/chunker"
	"github.com/gitslim/go-ragger/internal/config"
	"github.com/gitslim/go-ragger/internal/db"
	"github.com/gitslim/go-ragger/internal/logger"
	"github.com/gitslim/go-ragger/internal/rag"
	"github.com/gitslim/go-ragger/internal/version"
	"github.com/gitslim/go-ragger/internal/web"
	"github.com/google/uuid"
	"go.uber.org/fx"
)

func init() {
	// Регистрируем uuid.UUID для работы с gob
	gob.Register(uuid.UUID{})
}

func CreateServerApp() fx.Option {
	return fx.Options(
		fx.NopLogger,
		version.Module,
		logger.Module,
		config.Module,
		web.Module,
		db.Module,
		chunker.Module,
		rag.Module,
	)
}

// RunServerApp запускает приложение сервера
func RunServerApp() {
	fx.New(CreateServerApp()).Run()
}
