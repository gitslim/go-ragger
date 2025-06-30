package util

import "crypto/sha256"

func CalculateBytesHash(data []byte) []byte {
	hasher := sha256.New()
	hasher.Write(data)
	return hasher.Sum(nil)
}
