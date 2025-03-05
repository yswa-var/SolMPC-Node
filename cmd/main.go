package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"tilt-sim/internal/aggregator"
	"tilt-sim/internal/message"
	"tilt-sim/internal/validator"
)

func main() {
	// Seed the random number generator with the current time
	rand.Seed(time.Now().UnixNano())

	const numValidators = 10
	const threshold = 7 // Minimum validators required for reconstruction

	// 1. Create validators with DKG (Distributed Key Generation)
	// This initializes the validators and generates their private and public shares
	validators, _, groupPublicKey := validator.CreateValidators(numValidators, threshold)

	// Print DKG details
	fmt.Printf("\n🔐 Distributed Key Generation Simulation 🔐\n")
	fmt.Printf("Total Validators: %d\n", numValidators)
	fmt.Printf("Threshold: %d\n", threshold)
	fmt.Printf("Group Public Key: %v\n", groupPublicKey)

	// Demonstrate secret reconstruction
	// This simulates the reconstruction of the secret using the private shares of the validators
	reconstructedSecret := validator.ReconstructSecret(validators, threshold)
	println("🔑 Reconstructed Secret:", reconstructedSecret)

	// Messaging channel and wait group
	// The messageChannel is used for communication between validators and the aggregator
	// The wait group is used to wait for all goroutines to finish
	messageChannel := make(chan message.ThrashMessage, 10)
	var wg sync.WaitGroup

	// 2. Start validators as goroutines
	// Each validator runs in its own goroutine and processes messages
	wg.Add(numValidators)
	for _, v := range validators {
		go validator.ProcessValidator(v, validators, messageChannel, &wg)
	}

	// 3. Start aggregator in a separate goroutine
	// The aggregator collects messages and verifies signatures when the threshold is met
	wg.Add(1)
	go aggregator.Aggregate(messageChannel, threshold, &wg)

	// 4. Select two random validators to start messaging
	// Randomly select two different validators to initiate the message flow
	startValidator1 := validators[rand.Intn(numValidators)]
	startValidator2 := validators[rand.Intn(numValidators)]
	for startValidator1.ID == startValidator2.ID {
		startValidator2 = validators[rand.Intn(numValidators)]
	}

	// 5. Initiate message flow
	// Send initial messages from the selected validators
	fmt.Printf("\n📡 Starting message flow between Validator %d and Validator %d\n",
		startValidator1.ID, startValidator2.ID)
	startValidator1.Channel <- "Hello from validator 1!"
	startValidator2.Channel <- "Hello from validator 2!"

	// 6. Let the process run for some time
	// Allow the validators and aggregator to process messages for 10 seconds
	time.Sleep(10 * time.Second)

	// 7. Cleanup: Close channels
	// Close the channels to signal the end of message processing
	for _, v := range validators {
		close(v.Channel)
	}
	close(messageChannel)

	// 8. Wait for all goroutines to finish
	// Wait for all validators and the aggregator to complete their work
	wg.Wait()

	// Final simulation summary
	// Print a summary of the simulation results
	fmt.Println("\n🎉 DKG Simulation Completed Successfully!")
	fmt.Printf("Validators Used: %d\n", numValidators)
	fmt.Printf("Reconstruction Threshold: %d\n", threshold)
}
