package milvus

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino-ext/components/indexer/milvus"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

func NewMilvusIndexer(cli *client.Client, emb embedding.Embedder) (*milvus.Indexer, error) {
	ctx := context.Background()

	indexer, err := milvus.NewIndexer(ctx, &milvus.IndexerConfig{
		Client:            *cli,
		Embedding:         emb,
		MetricType:        milvus.CONSINE,
		DocumentConverter: documentConverter,
		Fields:            getFields(),
		// EnableDynamicSchema: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create milvus indexer: %w", err)
	}

	return indexer, nil
}

type mySchema struct {
	ID       string    `json:"id" milvus:"name:id"`
	Content  string    `json:"content" milvus:"name:content"`
	Vector   []float32 `json:"vector" milvus:"name:vector"`
	Metadata []byte    `json:"metadata" milvus:"name:metadata"`
}

func getFields() []*entity.Field {
	return []*entity.Field{
		entity.NewField().
			WithName("id").
			WithDescription("document unique id").
			WithIsPrimaryKey(true).
			WithDataType(entity.FieldTypeVarChar).
			WithMaxLength(255),
		entity.NewField().
			WithName("vector").
			WithDescription("document vector").
			WithIsPrimaryKey(false).
			WithDataType(entity.FieldTypeFloatVector).
			WithDim(1024),
		entity.NewField().
			WithName("content").
			WithDescription("document content").
			WithIsPrimaryKey(false).
			WithDataType(entity.FieldTypeVarChar).
			WithMaxLength(10240),
		entity.NewField().
			WithName("metadata").
			WithDescription("document metadata").
			WithIsPrimaryKey(false).
			WithDataType(entity.FieldTypeJSON),
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
	// spew.Dump(rows)

	return rows, nil
}

func vec64To32(vec []float64) []float32 {
	result := make([]float32, len(vec))
	for i, v := range vec {
		result[i] = float32(v)
	}
	return result
}
