package main

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"tilt-valid/internal/exchange"
	mpc "tilt-valid/internal/mpc"
	"tilt-valid/utils"
)

const threshold = 2

type Validator struct {
	ID      string
	Name    string
	Stake   float64
	Active  bool
	VRFHash *big.Int
}

type Tilt struct {
	ID         string  `json:"id"`
	Amount     float64 `json:"amount"`
	ReceiverID string  `json:"receiver_id"`
}

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
)

func logInfo(message string) {
	fmt.Printf("%s[INFO]%s %s\n", Cyan, Reset, message)
}

func logSuccess(message string) {
	fmt.Printf("%s[SUCCESS]%s %s\n", Green, Reset, message)
}

func logError(message string) {
	fmt.Printf("%s[ERROR]%s %s\n", Red, Reset, message)
}

func logWarning(message string) {
	fmt.Printf("%s[WARNING]%s %s\n", Yellow, Reset, message)
}

func separator(title string) {
	fmt.Printf("\n%s===== %s =====%s\n\n", Blue, title, Reset)
}

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
		if len(record) != 5 {
			return nil, fmt.Errorf("invalid record: %v", record)
		}
		stake, _ := strconv.ParseFloat(record[2], 64)
		active, _ := strconv.ParseBool(record[3])
		vrfHash := new(big.Int)
		vrfHash.SetString(record[4], 10)
		validators = append(validators, Validator{
			ID:      record[0],
			Name:    record[1],
			Stake:   stake,
			Active:  active,
			VRFHash: vrfHash,
		})
	}
	return validators, nil
}

// GenerateVRFHash generates a verifiable random hash for the validator
func generateVRFHash() *big.Int {
	// Generate a random seed based on current time and other inputs
	seed := time.Now().UnixNano()
	randomSeed := big.NewInt(seed)

	// Create hash of the seed
	h := sha256.New()
	h.Write(randomSeed.Bytes())
	hashBytes := h.Sum(nil)

	// Convert hash to big.Int
	vrfHash := new(big.Int).SetBytes(hashBytes)
	return vrfHash
}

// UpdateVRFHash updates the validator's VRF hash in the CSV file
func updateVRFHash(validatorID int, vrfHash *big.Int, filePath string) error {
	// Read existing file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to read CSV file: %v", err)
	}
	file.Close()

	// Update the record for the validator
	if validatorID < len(records) {
		records[validatorID][4] = vrfHash.String()
	} else {
		return fmt.Errorf("validator ID %d out of range", validatorID)
	}

	// Write back to file
	tempFile := filePath + ".tmp"
	outFile, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}

	writer := csv.NewWriter(outFile)
	err = writer.WriteAll(records)
	if err != nil {
		outFile.Close()
		return fmt.Errorf("failed to write to CSV file: %v", err)
	}

	writer.Flush()
	outFile.Close()

	// Replace the original file with the temporary file
	err = os.Rename(tempFile, filePath)
	if err != nil {
		return fmt.Errorf("failed to rename temp file: %v", err)
	}

	return nil
}

// SelectValidator selects a validator based on the combined VRF hashes
func selectValidator(validators []Validator) (int, error) {
	if len(validators) == 0 {
		return 0, fmt.Errorf("no validators available")
	}

	// Combine all VRF hashes
	combinedHash := new(big.Int)
	for _, validator := range validators {
		if validator.Active {
			combinedHash.Xor(combinedHash, validator.VRFHash)
		}
	}

	// Count active validators
	activeCount := 0
	for _, validator := range validators {
		if validator.Active {
			activeCount++
		}
	}

	if activeCount == 0 {
		return 0, fmt.Errorf("no active validators")
	}

	// Use the combined hash to select a validator (mod active count)
	// Add 1 because validator IDs are 1-based in this implementation
	selectedIndex := new(big.Int).Mod(combinedHash, big.NewInt(int64(activeCount))).Int64() + 1

	return int(selectedIndex), nil
}

func main() {
	args := os.Args[1:]
	wg := sync.WaitGroup{}

	if len(args) < 1 {
		logError("Usage: go run main.go <validator_id>")
		return
	}
	id, _ := strconv.Atoi(args[0])
	separator(fmt.Sprintf("Starting Validator ID: %d", id))

	validatorsFilePath := filepath.Join("/Users/apple/Documents/GitHub/tilt-validator-main/data", "validators.csv")
	validators, err := loadValidators(validatorsFilePath)
	if err != nil {
		logError(fmt.Sprintf("Error loading validators: %v", err))
		return
	}

	// Create a channel for incoming messages
	receiveChan := make(chan []byte, 10000)

	// Create a list of party IDs for message exchange
	var parties []uint16
	for i := range validators[1:] {
		parties = append(parties, uint16(i+1))
	}

	// Set up the transport and MPC party
	transport := exchange.NewTransport(id, parties)
	mpcParty := mpc.NewParty(uint16(id), utils.Logger(validators[id].ID, "main"))
	mpcParty.Init(parties, threshold, transport.SendMsg)

	separator("Distributed Key Generation (DKG)")
	logInfo("Initiating DKG process...")

	wg.Add(1)
	startTime := time.Now()
	var keyShare []byte
	go func() {
		defer wg.Done()
		keyShare, err = mpcParty.KeyGen(context.Background())
		if err != nil {
			logError(fmt.Sprintf("Error performing DKG: %v", err))
		} else {
			logSuccess(fmt.Sprintf("DKG completed. KeyShare length: %d", len(keyShare)))
		}
	}()

	// Start a goroutine to watch for incoming messages
	go transport.WatchFile(1*time.Millisecond, receiveChan)
	go func() {
		for msg := range receiveChan {
			var msgStructured exchange.Msg
			json.Unmarshal(msg, &msgStructured)
			mpcParty.OnMsg(msgStructured.Message, uint16(msgStructured.From), msgStructured.Broadcast)
		}
	}()

	wg.Wait() // Wait for DKG to complete
	logInfo(fmt.Sprintf("DKG completed in %.2f seconds", time.Since(startTime).Seconds()))
	transport.DeleteFileData()
	time.Sleep(2 * time.Second)

	// Signing process
	mpcParty.SetShareData(keyShare)
	msgToSign := []byte(utils.GenerateTransactionHash())
	digestMsg := mpc.Digest(msgToSign)

	separator("Signing Process")
	wg.Add(1)
	ctx := context.Background()
	var sign []byte

	// Reinitialize for signing
	mpcParty.Init(parties, threshold, transport.SendMsg)
	go func() {
		defer wg.Done()
		logInfo("Starting the signing process...")
		sign, err = mpcParty.Sign(ctx, digestMsg)
		if err != nil {
			logError(fmt.Sprintf("Failed to sign message: %v", err))
		} else {
			logSuccess(fmt.Sprintf("Signature generated: %x", sign))
		}
	}()
	wg.Wait()
	logInfo("Signing process completed.")

	pk, err := mpcParty.ThresholdPK()
	if err != nil {
		logError("Failed to get threshold public key")
		return
	}
	logInfo(fmt.Sprintf("Threshold PK: %x", pk))

	// VRF logic implementation
	separator("VRF-based Validator Selection")

	// Step 1: Generate VRF hash for this validator
	logInfo("Generating VRF hash...")
	vrfHash := generateVRFHash()
	logInfo(fmt.Sprintf("Generated VRF hash: %s", vrfHash.String()))

	// Step 2: Update the VRF hash in the CSV file
	updateVRFHash(id, vrfHash, validatorsFilePath)

	// Step 3: Wait for all validators to update their VRF hashes
	logInfo("Waiting for other validators to update their VRF hashes...")
	time.Sleep(5 * time.Second)

	// Step 4: Reload validators to get updated VRF hashes
	updatedValidators, err := loadValidators(validatorsFilePath)
	if err != nil {
		logError(fmt.Sprintf("Error reloading validators: %v", err))
		return
	}

	// Step 5: Select a validator based on combined VRF hashes
	selectedValidator, err := selectValidator(updatedValidators)
	if err != nil {
		logError(fmt.Sprintf("Error selecting validator: %v", err))
	} else {
		if selectedValidator == id {
			logSuccess(fmt.Sprintf("This validator (ID: %d) was selected for verification!", id))
			separator("Signature Verification")
			if ed25519.Verify(pk, digestMsg, sign) {
				logSuccess("✅ Signature verification successful!")
			} else {
				logError("❌ Signature verification failed!")
			}
		} else {
			logInfo(fmt.Sprintf("Validator ID: %d was selected for verification", selectedValidator))
		}
	}
}
