package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	mpc "teddsa/eddsa"
	"teddsa/exchange"
	"time"

	"go.uber.org/zap"
)

const threshold = 1

// keygen performs the distributed key generation protocol and returns the shares.
func keygen(p *mpc.Party) ([]byte, error) {
	errChan := make(chan error)
	res := make(chan []byte)

	var wg sync.WaitGroup
	wg.Add(1)
	go func(p *mpc.Party) {
		defer wg.Done()
		// Each party performs key generation.
		share, err := p.KeyGen(context.Background())
		if err != nil {
			errChan <- err
			return
		}
		res <- share
	}(p)

	wg.Wait()

	select {
	case share := <-res:
		return share, nil
	case err := <-errChan:
		return nil, fmt.Errorf(err.Error())
	}

}

// generateRandomHash generates a random SHA-256 hash string.
func solanaTransaction() (string, error) {
	// demo function whcih mimics the solana transaction for vaidation.
	// Create a new SHA-256 hash
	hash := sha256.New()

	// Generate a random number.
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %v", err)
	}

	// Write the random bytes to the hash.
	_, err = hash.Write(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to write to hash: %v", err)
	}

	// Compute the hash and return it as a hex string.
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// Validator represents a validator with an ID and name.
type Validator struct {
	ID   string
	Name string
}

// loadValidators loads the validators from a CSV file.
func loadValidators(filePath string) ([]Validator, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV file: %v", err)
	}

	var validators []Validator
	for _, record := range records {
		if len(record) != 2 {
			return nil, fmt.Errorf("invalid record: %v", record)
		}
		validators = append(validators, Validator{ID: record[0], Name: record[1]})
	}

	return validators, nil
}

func main() {
	args := os.Args[1:]

	wg := sync.WaitGroup{}

	if len(args) < 1 {
		fmt.Println("Usage: main2 <validator_id>")
		return
	}
	id, _ := strconv.Atoi(args[0])
	fmt.Println("ARG:", id)
	validators, err := loadValidators("validators.csv")
	if err != nil {
		fmt.Println("Error loading validators:", err)
	}

	receiveChan := make(chan []byte, 10000)

	var parties []uint16
	for i, _ := range validators[1:] {
		parties = append(parties, uint16(i+1))
	}
	transport := exchange.NewTransport(id, parties)

	mpcParty := mpc.NewParty(uint16(id), logger(validators[id].ID, "main"))

	mpcParty.Init(parties, threshold, transport.SendMsg)

	println("DKG")

	wg.Add(1)
	t1 := time.Now()
	go transport.WatchFile(1*time.Millisecond, receiveChan)

	go func() {
		defer wg.Done()
		keyShare, err := mpcParty.KeyGen(context.Background())
		if err != nil {
			fmt.Println("Error performing DKG:", err)
		}
		println("KeyShare:", len(keyShare))
	}()

	for {
		select {
		case msg := <-receiveChan:

			var msgStructured exchange.Msg
			fmt.Println("Received message:", string(msg))
			json.Unmarshal(msg, &msgStructured)
			mpcParty.OnMsg(msgStructured.Message, uint16(msgStructured.From), msgStructured.Broadcast)
		}
	}
	wg.Wait()

	println("DKG completed in", time.Since(t1))

}

// logger creates a logger with the given ID and test name.
func logger(id string, testName string) mpc.Logger {
	logConfig := zap.NewDevelopmentConfig()
	logger, _ := logConfig.Build()
	logger = logger.With(zap.String("t", testName)).With(zap.String("id", id))
	return logger.Sugar()
}
