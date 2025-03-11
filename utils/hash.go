package utils

import (
	"crypto/rand"
	"crypto/sha256"
)

func GenerateTransactionHash() []byte {
	randomBytes := make([]byte, 32)
	_, _ = rand.Read(randomBytes)
	hash := sha256.Sum256(randomBytes)
	return hash[:]
}
