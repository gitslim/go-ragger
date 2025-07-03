package milvus

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/retriever/milvus"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

func NewMilvusRetriever(cli *client.Client, emb embedding.Embedder) (*milvus.Retriever, error) {
	ctx := context.Background()

	retriever, err := milvus.NewRetriever(ctx, &milvus.RetrieverConfig{
		Client:      *cli,
		Collection:  "eino_collection",
		Partition:   nil,
		VectorField: "vector",
		OutputFields: []string{
			"id",
			"content",
			"metadata",
		},
		DocumentConverter: nil,
		MetricType:        entity.COSINE,
		TopK:              0,
		ScoreThreshold:    5,
		Sp:                &entity.IndexHNSWSearchParam{},
		Embedding:         emb,
		VectorConverter:   vectorConverter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create milvus retriever: %w", err)
	}

	return retriever, nil
}

func vectorConverter(ctx context.Context, vectors [][]float64) ([]entity.Vector, error) {
	vec := make([]entity.Vector, 0, len(vectors))
	for _, vector := range vectors {
		vec = append(vec, entity.FloatVector(vec64To32(vector)))
	}
	return vec, nil
}
