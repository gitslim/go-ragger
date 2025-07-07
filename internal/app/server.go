package app

import (
	"encoding/gob"

	"github.com/gitslim/go-ragger/internal/agent"
	"github.com/gitslim/go-ragger/internal/chunker"
	"github.com/gitslim/go-ragger/internal/config"
	"github.com/gitslim/go-ragger/internal/db"
	"github.com/gitslim/go-ragger/internal/embedder"
	"github.com/gitslim/go-ragger/internal/indexer"
	"github.com/gitslim/go-ragger/internal/logger"
	"github.com/gitslim/go-ragger/internal/mem"
	"github.com/gitslim/go-ragger/internal/model"
	"github.com/gitslim/go-ragger/internal/rag"
	"github.com/gitslim/go-ragger/internal/template"
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
		// fx.NopLogger,
		version.Module,
		logger.Module,
		config.Module,
		web.Module,
		db.Module,
		chunker.Module,
		embedder.ModuleOpenAIEmbedder,
		indexer.ModuleMilvusIndexer,
		rag.Module,
		agent.ModuleAgentFactory,
		mem.ModuleMemory,
		template.ModuleChatTemplate,
		model.ModuleModel,
	)
}

// RunServerApp запускает приложение сервера
func RunServerApp() {
	fx.New(CreateServerApp()).Run()
}
