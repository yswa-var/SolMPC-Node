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
	TiltDb          string
	Distribution    string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load("/Users/apple/Documents/GitHub/tv-solana_int/.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
		return nil, err
	}

	config := &Config{
		SolanaProductId: os.Getenv("SOLANA_PRODUCT_ID"),
		ValidatorPath:   os.Getenv("VALIDATOR_PATH"),
		TransportPath:   os.Getenv("TRANSPORT_PATH"),
		TiltDb:          os.Getenv("TILT_DB"),
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
