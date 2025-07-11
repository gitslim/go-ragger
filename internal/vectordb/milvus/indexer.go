package milvus

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino-ext/components/indexer/milvus"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
)

type MilvusIndexerFactory func(context.Context, *MilvusIndexerConfig) (*milvus.Indexer, error)

type MilvusIndexerConfig struct {
	Collection string
}

func NewMilvusIndexerFactory(cli *client.Client, emb embedding.Embedder) MilvusIndexerFactory {

	return func(ctx context.Context, config *MilvusIndexerConfig) (*milvus.Indexer, error) {
		indexer, err := milvus.NewIndexer(ctx, &milvus.IndexerConfig{
			Collection:        config.Collection,
			Client:            *cli,
			Embedding:         emb,
			MetricType:        milvus.CONSINE,
			DocumentConverter: documentConverter,
			Fields:            getFields(),
			// EnableDynamicSchema: true,
		})
		if err != nil {
			return nil, err
		}

		return indexer, nil
	}
}

func documentConverter(ctx context.Context, docs []*schema.Document, vectors [][]float64) ([]any, error) {
	em := make([]mySchema, 0, len(docs))
	texts := make([]string, 0, len(docs))
	rows := make([]any, 0, len(docs))

	for _, doc := range docs {
		metadata, err := sonic.Marshal(doc.MetaData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		em = append(em, mySchema{
			ID:       doc.ID,
			Content:  doc.Content,
			Vector:   nil,
			Metadata: metadata,
		})
		texts = append(texts, doc.Content)
	}

	for idx, vec := range vectors {
		em[idx].Vector = vec64To32(vec)
		rows = append(rows, &em[idx])
	}

	return rows, nil
}
