package vrf

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"

	"golang.org/x/crypto/ed25519"
)

// VRFKeyPair stores the public and private keys for VRF
type VRFKeyPair struct {
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
}

// GenerateKeyPair generates a new Ed25519 key pair for VRF
func GenerateKeyPair() (*VRFKeyPair, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate VRF key pair: %v", err)
	}
	return &VRFKeyPair{PublicKey: publicKey, PrivateKey: privateKey}, nil
}

// Evaluate computes the VRF output and proof
func Evaluate(sk ed25519.PrivateKey, message []byte) ([]byte, []byte, error) {
	// Hash message with SHA-256
	hashedMessage := sha256.Sum256(message)

	// Sign the hashed message using Ed25519 private key
	signature := ed25519.Sign(sk, hashedMessage[:])

	// The VRF output is the hashed signature
	vrfOutput := sha256.Sum256(signature)

	return vrfOutput[:], signature, nil
}

// Verify checks if the proof is valid for the given input
func Verify(pk ed25519.PublicKey, message, vrfOutput, proof []byte) bool {
	// Hash the message
	hashedMessage := sha256.Sum256(message)

	// Verify the signature
	if !ed25519.Verify(pk, hashedMessage[:], proof) {
		return false
	}

	// Recompute the VRF output and compare
	expectedOutput := sha256.Sum256(proof)
	return string(vrfOutput) == string(expectedOutput[:])
}
