package app

import (
	"encoding/gob"
	"os"

	"github.com/gitslim/go-ragger/internal/agent"
	"github.com/gitslim/go-ragger/internal/chunker"
	"github.com/gitslim/go-ragger/internal/config"
	"github.com/gitslim/go-ragger/internal/db"
	"github.com/gitslim/go-ragger/internal/embedder"
	"github.com/gitslim/go-ragger/internal/indexator"
	"github.com/gitslim/go-ragger/internal/logger"
	"github.com/gitslim/go-ragger/internal/mem"
	"github.com/gitslim/go-ragger/internal/model"
	"github.com/gitslim/go-ragger/internal/template"
	"github.com/gitslim/go-ragger/internal/vectordb"
	"github.com/gitslim/go-ragger/internal/version"
	"github.com/gitslim/go-ragger/internal/web"
	"github.com/google/uuid"
	"github.com/spf13/pflag"
	"go.uber.org/fx"
)

func init() {
	// Register UUID type to be able to use it in gob
	gob.Register(uuid.UUID{})
}

// CreateServerApp creates a fx app for the server
func CreateServerApp() fx.Option {
	return fx.Options(
		fx.NopLogger,
		logger.Module,
		config.Module,
		version.Module,
		web.Module,
		db.Module,
		chunker.Module,
		embedder.ModuleOpenAIEmbedder,
		vectordb.ModuleMilvus,
		indexator.ModuleIndexator,
		agent.ModuleAgentFactory,
		mem.ModuleMemory,
		template.ModuleChatTemplate,
		model.ModuleModel,
	)
}

// RunServerApp runs the server app.
func RunServerApp() {
	checkVersionFlag()
	fx.New(CreateServerApp()).Run()
}

// checkVersionFlag checks if the version flag is set and print app version information
func checkVersionFlag() {
	pflag.Bool("version", false, "Print version information")
	pflag.Parse()

	if pflag.Lookup("version").Changed {
		version.PrintVersion()
		os.Exit(0)
	}
}
