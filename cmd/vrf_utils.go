package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"
	"time"
)

// VRFProof stores the proof of randomness generation
type VRFProof struct {
	PublicKey    ed25519.PublicKey
	Beta         []byte // The VRF output
	Proof        []byte // The proof that beta was correctly computed
	RandomNumber *big.Int
}

// GenerateVRF uses the validator's private key to generate a verifiable random value
// from a given input (e.g., block hash + round number)
func GenerateVRF(privKey ed25519.PrivateKey, pubKey ed25519.PublicKey, input []byte) (*VRFProof, error) {
	// Create a message by combining input data
	message := input

	// In a real implementation, this would use a specialized VRF algorithm
	// For demonstration, we're using ed25519 signatures as a simplified VRF
	signature := ed25519.Sign(privKey, message)

	// Hash the signature to get the VRF output (beta)
	beta := sha256.Sum256(signature)

	// Convert beta to a big.Int for easier comparisons
	randomNumber := new(big.Int).SetBytes(beta[:])

	// In a real VRF, we would include a proper cryptographic proof
	// Here we're just using the signature itself as the "proof"
	return &VRFProof{
		PublicKey:    pubKey,
		Beta:         beta[:],
		Proof:        signature,
		RandomNumber: randomNumber,
	}, nil
}

// VerifyVRF verifies that the VRF output was correctly computed
func VerifyVRF(proof *VRFProof, input []byte) (bool, error) {
	// In a real VRF, we'd verify the cryptographic proof
	// For our simplified example, we just verify the signature
	return ed25519.Verify(proof.PublicKey, input, proof.Proof), nil
}

// GenerateCommonSeed creates a deterministic seed for a specific round/block
func GenerateCommonSeed(blockHeight uint64, timestamp time.Time) []byte {
	seed := make([]byte, 16)
	binary.BigEndian.PutUint64(seed[0:8], blockHeight)
	binary.BigEndian.PutUint64(seed[8:16], uint64(timestamp.Unix()))
	return seed
}

// SelectValidator uses VRF proofs from all validators to deterministically select one
func SelectValidator(validators []Validator, vrfProofs []*VRFProof) *Validator {
	if len(validators) == 0 || len(vrfProofs) == 0 {
		logWarning("No validators or VRF proofs available")
		return nil
	}

	// Find the lowest VRF output
	lowestValue, _ := new(big.Int).SetString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)
	var selectedIndex int = -1

	for i, proof := range vrfProofs {
		// Skip invalid proofs (nil or zero)
		if proof == nil || proof.RandomNumber.Cmp(big.NewInt(0)) == 0 {
			continue
		}

		logInfo(fmt.Sprintf("Validator %s VRF value: %s", validators[i].ID, proof.RandomNumber.String()))

		// Compare VRF outputs and select the lowest
		if proof.RandomNumber.Cmp(lowestValue) < 0 {
			lowestValue = proof.RandomNumber
			selectedIndex = i
			logInfo(fmt.Sprintf("New lowest VRF found: %s (Validator %s)",
				lowestValue.String(), validators[i].ID))
		}
	}

	if selectedIndex == -1 {
		logError("No validator was selected")
		return nil
	}

	logSuccess(fmt.Sprintf("Chosen Validator: %s", validators[selectedIndex].ID))
	return &validators[selectedIndex]
}
