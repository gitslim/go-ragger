package milvus

import (
	"crypto/sha256"
	"encoding/hex"
)

// ToMilvusName converts a string to valid milvus name
func ToMilvusName(s string) string {
	hash := sha256.Sum256([]byte(s))
	return "_" + hex.EncodeToString(hash[:16])
}
