package milvus

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/retriever/milvus"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

type MilvusRetrieverFactory func(context.Context, *MilvusRetrieverConfig) (*milvus.Retriever, error)

type MilvusRetrieverConfig struct {
	Collection string
}

func NewMilvusRetrieverFactory(cli *client.Client, emb embedding.Embedder) MilvusRetrieverFactory {
	return func(ctx context.Context, config *MilvusRetrieverConfig) (*milvus.Retriever, error) {

		retriever, err := milvus.NewRetriever(ctx, &milvus.RetrieverConfig{
			Client:      *cli,
			Collection:  config.Collection,
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

		// Retrieve documents example
		// _, err := retriever.Retrieve(ctx, "search query")
		// if err != nil {
		// 	fmt.Printf("Failed to retrieve milvus documents: %v", err)
		// }

		// for i, doc := range documents {
		// 	fmt.Printf("Document %d:\n", i)
		// 	fmt.Printf("title: %s\n", doc.ID)
		// 	fmt.Printf("content: %s\n", doc.Content)
		// 	fmt.Printf("metadata: %v\n", doc.MetaData)
		// }

		return retriever, nil
	}
}

func vectorConverter(ctx context.Context, vectors [][]float64) ([]entity.Vector, error) {
	vec := make([]entity.Vector, 0, len(vectors))
	for _, vector := range vectors {
		vec = append(vec, entity.FloatVector(vec64To32(vector)))
	}
	return vec, nil
}
