package main

import (
	"fmt"
	"tilt-validator/pkg/signature"

	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/sign/bdn"
)

func main() {
	// Parameters
	n := 5 // Total participants
	t := 3 // Threshold
	message := []byte("Hello, BDN Threshold Signature!")

	// Initialize suite
	suite := bn256.NewSuiteG2()

	// Step 1: Generate keys and shares
	kg := signature.NewKeyGenerator(n, t)
	privateKey, shares, publicKey, publicKeys, err := kg.GenerateShares()
	fmt.Println("private key:", privateKey)
	if err != nil {
		fmt.Println("Error generating shares:", err)
		return
	}
	fmt.Println("Master private key generated, shares distributed.")

	// Step 2: Simulate participants signing
	participants := make([]*signature.Participant, n)
	partialSigs := make(map[int][]byte)
	for i := 0; i < n; i++ {
		participants[i] = signature.NewParticipant(i+1, shares[i], suite)
		if i < t { // Only first t participants sign
			sig, err := participants[i].PartialSign(message)
			if err != nil {
				fmt.Printf("Participant %d failed to sign: %v\n", i+1, err)
				return
			}
			partialSigs[i+1] = sig
			fmt.Printf("Participant %d signed successfully.\n", i+1)
		}
	}

	// Step 3: Aggregate signatures
	agg := signature.NewAggregator(suite, t, publicKeys)
	aggregatedSig, err := agg.Aggregate(partialSigs, message, publicKey)
	if err != nil {
		fmt.Println("Aggregation failed:", err)
		return
	}
	fmt.Println("Signature aggregated successfully!")

	// Step 4: Verify the signature (for demonstration)
	err = bdn.Verify(suite, publicKey, message, aggregatedSig)
	if err != nil {
		fmt.Println("Verification failed:", err)
		return
	}
	fmt.Println("Signature verified successfully!")
}

// for this we might need to create a fully developed purchaser and a significat masking function
// partial signatures are not enought to verify the signature of the message
// what message are we signing?
// what is the message that we are signing?
// need to understad the pholosophy behind the what is the use for this thing.
