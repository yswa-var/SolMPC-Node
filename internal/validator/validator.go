package validator

import (
	"fmt"
	"math/big"
	"sync"

	"tilt-sim/internal/message"
)

// Validator represents an entity that processes messages and participates in DKG
type Validator struct {
	ID              int         // Unique identifier for the validator
	Channel         chan string // Channel for receiving messages
	PrivateKeyShare *big.Int    // Private key share for the validator
	PublicShare     *big.Int    // Public key share for the validator
}

// ProcessValidator function that processes and forwards messages
// This function runs as a goroutine for each validator
func ProcessValidator(v *Validator, validators []*Validator, messageChannel chan message.ThrashMessage, wg *sync.WaitGroup) {
	defer wg.Done() // Ensure the wait group counter is decremented when the function exits

	// Message processing logic
	for msg := range v.Channel {
		// Forward the message to the messageChannel with the validator's ID
		messageChannel <- message.ThrashMessage{
			ID:      fmt.Sprintf("Validator-%d", v.ID),
			Content: msg,
		}
	}
}

// CreateValidators initializes validators with DKG preparation
// This function creates the specified number of validators, generates their secret shares,
// distributes the shares, and computes the group public key
func CreateValidators(numValidators int, threshold int) ([]*Validator, []*big.Int, *big.Int) {
	// Create a slice to hold the validators
	validators := make([]*Validator, numValidators)
	for i := 0; i < numValidators; i++ {
		// Initialize each validator with an ID and a message channel
		validators[i] = &Validator{
			ID:      i,
			Channel: make(chan string, 1),
		}
	}

	// Generate secret shares using Shamir's Secret Sharing
	privateShares, publicShares := GenerateSecretShares(numValidators, threshold)

	// Distribute the generated shares to the validators
	DistributeShares(validators, privateShares, publicShares)

	// Compute the group public key by combining the public shares of all validators
	groupPublicKey := ComputeGroupPublicKey(validators)

	// Return the list of validators, their public shares, and the group public key
	return validators, publicShares, groupPublicKey
}
