package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	SolanaProductId string
	ValidatorPath   string
	TransportPath   string
	TiltDb          string
	Distribution    string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load("/Users/yash/Documents/SolMPC-Node/.env")
	if err != nil {
		// Don't fail if .env file doesn't exist, just log it
		fmt.Printf("Warning: .env file not found, using default values\n")
	}

	// Set default values if environment variables are not set
	tiltDb := os.Getenv("TILT_DB")
	if tiltDb == "" {
		tiltDb = "/Users/yash/Documents/SolMPC-Node/utils/tiltdb.csv"
	}

	config := &Config{
		SolanaProductId: os.Getenv("SOLANA_PRODUCT_ID"),
		ValidatorPath:   os.Getenv("VALIDATOR_PATH"),
		TransportPath:   os.Getenv("TRANSPORT_PATH"),
		TiltDb:          tiltDb,
		Distribution:    os.Getenv("DISTRIBUTION_DUMP"),
	}

	return config, nil
}

// func main() {
// 	cfg, err := config.LoadConfig()
// 	if err != nil {
// 		log.Fatalf("Error loading config: %v", err)
// 	}

// 	fmt.Printf("Solana Product ID: %s\n", cfg.SolanaProductId)
// 	fmt.Printf("Validator Path: %s\n", cfg.ValidatorPath)
// 	fmt.Printf("Transport Path: %s\n", cfg.TransportPath)
// }
