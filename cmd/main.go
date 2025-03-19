package main

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"tilt-valid/cmd/config"
	"tilt-valid/internal/distribution"
	"tilt-valid/internal/exchange"
	mpc "tilt-valid/internal/mpc"
	"tilt-valid/utils"

	"github.com/gagliardetto/solana-go"
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

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("error in loading config")
	}
	path := cfg.ValidatorPath

	validatorsFilePath := filepath.Join(path, "validators.csv")
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
	td, err := distribution.NewTiltDistribution("http://127.0.0.1:8899")
	if err != nil {
		log.Fatal(err)
	}

	sender, _ := solana.PrivateKeyFromSolanaKeygenFile("/Users/apple/Desktop/tilt/tilt/sender.json")
	tiltMint := solana.MustPublicKeyFromBase58("F9VF1XvPmfuXW95s6WkP3gL2sARZy1EAA5HJn5GeJcG1")
	recipient1Token := solana.MustPublicKeyFromBase58("HqcqSHMsek4ojuh57arkwvT2aiwPhB4ucscU8DBrxCEC") // Replace
	recipient2Token := solana.MustPublicKeyFromBase58("4DM24i3wrVYrh4NmVStd3REyCW5u5KWcBWXqfKZQiBNj") // Replace

	// Tilt creation and synchronization
	flagFilePath := "/Users/apple/Documents/GitHub/tilt-validator-main/utils/create-tilt-flag.txt"
	tiltDataFile := "/Users/apple/Documents/GitHub/tilt-validator-main/utils/current-tilt-data.txt"
	var distributions map[solana.PublicKey]uint64

	if id == 1 { // Leader validator
		flagFile, err := os.ReadFile(flagFilePath)
		flagValue := 0
		if err == nil {
			flagValue, _ = strconv.Atoi(string(flagFile))
		}
		if flagValue == 0 {
			// Create tilt
			t := utils.CreateTilt(sender.PublicKey(), recipient1Token, recipient2Token, true)
			_, err := utils.FetchTiltData(t)
			if err != nil {
				logError(fmt.Sprintf("Failed to fetch tilt data: %v", err))
				return
			}
			// Flatten tilt into distributions
			distributions, err = distribution.FlattenTilt(t, 1000000000) // 1 token with 9 decimals
			if err != nil {
				logError(fmt.Sprintf("Failed to flatten tilt: %v", err))
				return
			}
			// Convert to map[string]uint64 for JSON
			distStr := make(map[string]uint64)
			for k, v := range distributions {
				distStr[k.String()] = v
			}
			// Save to file
			distJSON, err := json.Marshal(distStr)
			if err != nil {
				logError(fmt.Sprintf("Failed to marshal distributions: %v", err))
				return
			}
			err = os.WriteFile(tiltDataFile, distJSON, 0644)
			if err != nil {
				logError(fmt.Sprintf("Failed to write distributions: %v", err))
				return
			}
			// Update flag
			err = os.WriteFile(flagFilePath, []byte("1"), 0644)
			if err != nil {
				logError(fmt.Sprintf("Failed to update flag: %v", err))
				return
			}
			logSuccess("Tilt created and distributions saved by leader")
		}
	} else { // Other validators
		for {
			flagFile, err := os.ReadFile(flagFilePath)
			if err != nil {
				logError(fmt.Sprintf("Failed to read flag file: %v", err))
				time.Sleep(1 * time.Second)
				continue
			}
			flagValue, err := strconv.Atoi(string(flagFile))
			if err != nil {
				logError(fmt.Sprintf("Invalid flag value: %v", err))
				time.Sleep(1 * time.Second)
				continue
			}
			if flagValue == 1 {
				break
			}
			time.Sleep(1 * time.Second)
		}
	}

	// All validators read the distributions
	distBytes, err := os.ReadFile(tiltDataFile)
	if err != nil {
		logError(fmt.Sprintf("Failed to read tilt data: %v", err))
		return
	}
	var distStr map[string]uint64
	fmt.Printf("distBytes: %s\n", distBytes)
	fmt.Printf("distStr: %v\n", distStr)
	err = json.Unmarshal(distBytes, &distStr)
	if err != nil {
		logError(fmt.Sprintf("Failed to unmarshal distributions: %v", err))
		// ERROR] Failed to unmarshal distributions: invalid character 'T' looking for beginning of
		return
	}
	distributions = make(map[solana.PublicKey]uint64)
	for k, v := range distStr {
		pk, err := solana.PublicKeyFromBase58(k)
		if err != nil {
			logWarning(fmt.Sprintf("Invalid public key in distributions: %s", k))
			continue
		}
		distributions[pk] = v
	}
	logInfo("Distributions loaded:")
	for k, v := range distributions {
		logInfo(fmt.Sprintf("  %s: %d lamports", k.String(), v))
	}

	mpcParty.SetShareData(keyShare)
	sortedDist := make(map[string]uint64)
	var keys []string
	for k := range distStr {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		sortedDist[k] = distStr[k]
	}
	msgToSign, err := json.Marshal(sortedDist)
	if err != nil {
		logError(fmt.Sprintf("Failed to marshal msgToSign: %v", err))
		return
	}
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

	// updating flag to: tilt complete.
	utils.UpdateTiltCounter(0)

	pk, err := mpcParty.ThresholdPK()
	if err != nil {
		logError("Failed to get threshold public key")
		return
	}
	logInfo(fmt.Sprintf("Threshold PK: %x", pk))

	// VRF logic implementation
	separator("VRF-based Validator Selection")

	// Generate VRF hash for this validator
	logInfo("Generating VRF hash...")
	vrfHash := generateVRFHash()
	logInfo(fmt.Sprintf("Generated VRF hash: %s", vrfHash.String()))

	// Update the VRF hash in the CSV file
	updateVRFHash(id, vrfHash, validatorsFilePath)

	// Wait for all validators to update their VRF hashes
	logInfo("Waiting for other validators to update their VRF hashes...")
	time.Sleep(5 * time.Second)

	// Reload validators to get updated VRF hashes
	updatedValidators, err := loadValidators(validatorsFilePath)
	if err != nil {
		logError(fmt.Sprintf("Error reloading validators: %v", err))
		return
	}
	selectedValidator, err := selectValidator(updatedValidators)
	if err != nil {
		logError(fmt.Sprintf("Error selecting validator: %v", err))
	} else {
		if selectedValidator == id {
			logSuccess(fmt.Sprintf("This validator (ID: %d) was selected for verification!", id))
			separator("Signature Verification")
			if ed25519.Verify(pk, digestMsg, sign) {
				logSuccess("✅ Signature verification successful!")
				sig, err := td.Distribute(&sender, tiltMint, distributions)
				if err != nil {
					logError(fmt.Sprintf("Failed to distribute: %v", err))
					return
				}
				logSuccess(fmt.Sprintf("Transaction Signature: %s", sig))
				logInfo("Final Distributions:")
				for k, v := range distributions {
					logInfo(fmt.Sprintf("  %s: %d lamports", k.String(), v))
				}
			} else {
				logError("❌ Signature verification failed!")
			}
		} else {
			logInfo(fmt.Sprintf("Validator ID: %d was selected for verification", selectedValidator))
		}
	}
}
