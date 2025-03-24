package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	SolanaProductId string
	ValidatorPath   string
	TransportPath   string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load("/Users/apple/Documents/GitHub/tilt-validator-main/.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
		return nil, err
	}

	config := &Config{
		SolanaProductId: os.Getenv("SOLANA_PRODUCT_ID"),
		ValidatorPath:   os.Getenv("VALIDATOR_PATH"),
		TransportPath:   os.Getenv("TRANSPORT_PATH"),
	}

	return config, nil
}

// func main() { pubkey: DhPW6ne1DgYUZw3Dz94qfx68PvqSaiTfPijyf4MH9onK
// 	cfg, err := config.LoadConfig()
// 	if err != nil {
// 		log.Fatalf("Error loading config: %v", err)
// 	}

// 	fmt.Printf("Solana Product ID: %s\n", cfg.SolanaProductId)
// 	fmt.Printf("Validator Path: %s\n", cfg.ValidatorPath)
// 	fmt.Printf("Transport Path: %s\n", cfg.TransportPath)
// }
