package milvus

import "github.com/milvus-io/milvus-sdk-go/v2/entity"

// mySchema is the schema of the milvus document
type mySchema struct {
	ID       string    `json:"id" milvus:"name:id"`
	Content  string    `json:"content" milvus:"name:content"`
	Vector   []float32 `json:"vector" milvus:"name:vector"`
	Metadata []byte    `json:"metadata" milvus:"name:metadata"`
}

// getFields returns the fields of the milvus document
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
			WithMaxLength(65535),
		entity.NewField().
			WithName("metadata").
			WithDescription("document metadata").
			WithIsPrimaryKey(false).
			WithDataType(entity.FieldTypeJSON),
	}
}

// vec64To32 converts a slice of float64 to a slice of float32
func vec64To32(vec []float64) []float32 {
	result := make([]float32, len(vec))
	for i, v := range vec {
		result[i] = float32(v)
	}
	return result
}
