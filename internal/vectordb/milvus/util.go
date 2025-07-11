package milvus

import (
	"crypto/sha256"
	"encoding/hex"
)

func ToMilvusName(s string) string {
	hash := sha256.Sum256([]byte(s))
	return "_" + hex.EncodeToString(hash[:16])
}
