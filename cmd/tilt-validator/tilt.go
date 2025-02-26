package main

import (
	"fmt"
	"math/rand"
	"time"
)

// TiltAccount represents the simulated on-chain data for the Tilt PDA.
type TiltAccount struct {
	Balance            uint64   // Tilt’s balance (e.g., in lamports)
	BusinessRules      string   // Business rules or configuration info
	EphemeralWeighting float64  // Some weight/score value
	SubTiltReferences  []string // References to sub-tilt accounts or entities
}

// generateRandomTiltAccount simulates fetching Tilt account data by generating random values.
func generateRandomTiltAccount() TiltAccount {
	// Seed the random number generator
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate a random balance (e.g., up to 1,000,000 lamports)
	balance := uint64(r.Int63n(1_000_000))

	// Simulate business rules as a random rule identifier
	businessRules := fmt.Sprintf("Rule-%d", r.Intn(1000))

	// Generate a random ephemeral weighting between 0 and 100
	ephemeralWeighting := r.Float64() * 100

	// Generate between 1 and 5 sub-tilt references with random identifiers
	numSubTilts := r.Intn(5) + 1
	subTiltRefs := make([]string, numSubTilts)
	for i := 0; i < numSubTilts; i++ {
		subTiltRefs[i] = fmt.Sprintf("SubTilt-%d", r.Intn(1000))
	}

	return TiltAccount{
		Balance:            balance,
		BusinessRules:      businessRules,
		EphemeralWeighting: ephemeralWeighting,
		SubTiltReferences:  subTiltRefs,
	}
}
