package vectordb

import (
	"github.com/gitslim/go-ragger/internal/vectordb/milvus"
	"go.uber.org/fx"
)

var ModuleMilvus = fx.Module("milvus",
	fx.Provide(
		milvus.NewMilvusClient,
		milvus.NewMilvusIndexerFactory,
		milvus.NewMilvusRetrieverFactory,
	),
)
