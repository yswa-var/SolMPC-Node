// Entry point for the application: sets up necessary components and starts the validator.

package main

import (
	"fmt"
	_ "tilt-validator/internal/models"
)

func main() {
	// Step 1: Load configuration
	config, err := config_loader.LoadConfig()
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		return
	}

	// Step 2: Initialize Solana client using configuration data
	solanaClient, err := client.NewClient(config.SolanaEndpoint)
	if err != nil {
		fmt.Println("Error initializing Solana client:", err)
		return
	}

	// Step 3: Start the validator process
	if err := start.RunValidator(solanaClient); err != nil {
		fmt.Println("Error starting validator process:", err)
		return
	}

	fmt.Println("Validator process started successfully!")
}
