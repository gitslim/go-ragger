package indexer

import (
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/gitslim/go-ragger/internal/indexer/milvus"
	"go.uber.org/fx"
)

var ModuleMilvusIndexer = fx.Module("milvus-indexer",
	fx.Provide(
		milvus.NewMilvusClient,
		fx.Annotate(
			milvus.NewMilvusIndexer,
			fx.As(new(indexer.Indexer)),
		),
		fx.Annotate(
			milvus.NewMilvusRetriever,
			fx.As(new(retriever.Retriever)),
		),
	),
	fx.Invoke(milvus.RunMilvusClient),
)
