package vectordb

import (
	"github.com/gitslim/go-ragger/internal/vectordb/milvus"
	"go.uber.org/fx"
)

// ModuleMilvus is fx module for Milvus vector database
var ModuleMilvus = fx.Module("milvus",
	fx.Provide(
		milvus.NewMilvusClient,
		milvus.NewMilvusIndexerFactory,
		milvus.NewMilvusRetrieverFactory,
	),
)
