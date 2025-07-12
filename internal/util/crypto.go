package util

import "crypto/sha256"

// CalculateHash calculates the hash of the given data using SHA-256 algorithm.
func CalculateBytesHash(data []byte) []byte {
	hasher := sha256.New()
	hasher.Write(data)
	return hasher.Sum(nil)
}
