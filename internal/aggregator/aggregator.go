package aggregator

import (
	"fmt"
	"sync"

	"tilt-sim/internal/message"
)

// Aggregate collects messages and calls SignatureVerification when the threshold is met
// This function runs as a goroutine and processes messages from the messageChannel
func Aggregate(messageChannel chan message.ThrashMessage, threshold int, wg *sync.WaitGroup) {
	defer wg.Done() // Ensure the wait group counter is decremented when the function exits

	// messageSet is a map to store unique messages
	messageSet := make(map[string]message.ThrashMessage)

	for {
		// Read a message from the messageChannel
		msg, open := <-messageChannel
		if !open {
			// If the channel is closed, exit the loop
			break
		}

		// Collect only unique messages
		if _, exists := messageSet[msg.ID]; !exists {
			// If the message ID is not already in the set, add it
			messageSet[msg.ID] = msg
			fmt.Printf("Aggregator collected message: %v\n", msg)
		}

		// If the threshold is reached, process collected messages
		if len(messageSet) >= threshold {
			// Convert the map to a slice for processing
			messages := make([]message.ThrashMessage, 0, len(messageSet))
			for _, msg := range messageSet {
				messages = append(messages, msg)
			}
			// Call the signature verification function
			signatureVerification(messages)
			fmt.Println("Code completion message: Threshold met, closing all processes.")
			return
		}
	}
}

// signatureVerification simulates the signature verification process
// This function prints the details of each message being verified
func signatureVerification(messages []message.ThrashMessage) {
	fmt.Println("\n--- Signature Verification Started ---")
	for _, msg := range messages {
		// Print the ID and content of each message
		fmt.Printf("Verifying: ID=%s, Content=%s\n", msg.ID, msg.Content)
	}
	fmt.Println("--- Signature Verification Completed ---")
}
