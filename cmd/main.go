package main

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"flag"
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
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/mr-tron/base58"
)

// Define a flag for selecting the tilt type
var tiltType string

func init() {
	flag.StringVar(&tiltType, "tilt-type", "two_subtilts", "Type of tilt to create (simple, one_subtilt, two_subtilts, nested)")
}

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
	flag.Parse()

	if len(args) < 1 {
		logError("Usage: go run main.go <validator_id> [--tilt-type=<tilt_type>]")
		return
	}
	id, _ := strconv.Atoi(args[0])
	separator(fmt.Sprintf("Starting Validator ID: %d", id))

	// rpc bullshit
	// Initialize RPC client for Devnet
	client := rpc.New("https://api.devnet.solana.com")

	// Set up sender's keypair
	rawKey := []byte{164, 32, 125, 222, 175, 46, 12, 156, 205, 52, 159, 66, 48, 140, 67, 30, 202, 130, 25, 221, 94, 98, 188, 181, 101, 240, 113, 202, 30, 224, 226, 175, 50, 182, 177, 75, 11, 202, 132, 188, 121, 143, 136, 254, 127, 220, 33, 51, 201, 52, 42, 223, 221, 50, 176, 171, 16, 144, 64, 7, 231, 129, 21, 151}
	base58Encoded := base58.Encode(rawKey)
	senderPrivateKey, err := solana.PrivateKeyFromBase58(base58Encoded)
	if err != nil {
		log.Fatalf("Failed to parse sender private key: %v", err)
	}
	senderPubkey := senderPrivateKey.PublicKey()

	// Define the program ID
	programID := solana.MustPublicKeyFromBase58("3pH5Q1nfYuGKECw1Ljj7otGvm3VfRjFWxqraVQPACRiM")

	// Define parameters for the initialize function
	businessRules := [10]byte{10, 10, 10, 10, 10, 10, 10, 10, 10, 10} // Sums to 100

	// Generate 10 receiver public keys and amounts
	var receivers [10]distribution.Receiver
	for i := range receivers {
		kp := solana.NewWallet()
		receivers[i] = distribution.Receiver{
			Pubkey: kp.PublicKey(),
			Amount: 1000, // Example amount in lamports
		}
	}

	subTilts := []string{"sub_tilt1", "sub_tilt2"}

	// Create the instruction using the distribution package
	instruction, err := distribution.CreateInitializeInstruction(programID, senderPubkey, businessRules, receivers, subTilts)
	if err != nil {
		log.Fatalf("Failed to create initialize instruction: %v", err)
	}

	// Check for tilt-type flag in args
	for _, arg := range args {
		if len(arg) > 11 && arg[:11] == "--tilt-type=" {
			tiltType = arg[11:]
			break
		}
	}

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
	flag, _ := utils.ReadTiltCounter()
	if flag == 0 {
		senderKey := sha256.Sum256(make([]byte, ed25519.PublicKeySize))
		senderKeyStr := fmt.Sprintf("%x", senderKey)
		utils.GetTestTilt(tiltType, senderKeyStr)
		utils.UpdateTiltCounter(1)
	}
	time.Sleep(2 * time.Second)

	// All validators read the distributions
	// Read the tilt data for the given validator ID
	distBytes, err := utils.ReadTiltDataByID(1)
	if err != nil {
		logError(fmt.Sprintf("Error reading tilt data: %v", err))
		return
	}
	currentTilt, err := distribution.Distribution(distBytes)
	flattenData, err := distribution.AllocateAmounts(currentTilt)
	if err != nil {
		logError(fmt.Sprintf("Failed to distribute: %v", err))
		return
	}
	// Save flattenData to distribution-dump.csv
	dumpFilePath := filepath.Join("/Users/yash/Downloads/exercises/tilt-validator/utils/", "distribution-dump.csv")
	dumpFile, err := os.Create(dumpFilePath)
	if err != nil {
		logError(fmt.Sprintf("Failed to create distribution-dump.csv file: %v", err))
		return
	}
	defer dumpFile.Close()

	writer := csv.NewWriter(dumpFile)
	defer writer.Flush()

	// Write headers
	headers := []string{"ReceiverID", "Amount"}
	if err := writer.Write(headers); err != nil {
		logError(fmt.Sprintf("Failed to write headers to CSV file: %v", err))
		return
	}

	// Write data
	for receiverID, amount := range flattenData {
		record := []string{strconv.Itoa(receiverID), fmt.Sprintf("%f", amount)}
		if err := writer.Write(record); err != nil {
			logError(fmt.Sprintf("Failed to write record to CSV file: %v", err))
			return
		}
	}

	logSuccess("Flatten data saved to distribution-dump.csv")
	// Convert flattenData to a map[string]interface{} for msgToSign

	distStr := make(map[string]uint64)

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
	msgToSign, err := json.Marshal(instruction)
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
	// updating flag to: tilt complete.
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

			// seding signature to the network
			sig, err := distribution.SendTransaction(client, []solana.Instruction{instruction}, senderPrivateKey)
			if err != nil {
				log.Fatalf("Failed to send transaction: %v", err)
			}

			log.Printf("Transaction sent successfully: %s", sig)

		} else {
			logInfo(fmt.Sprintf("Validator ID: %d was selected for verification", selectedValidator))
		}
	}
	utils.UpdateTiltCounter(0)
	// Clear the content of the tiltdb.csv file
	tiltDBFilePath := filepath.Join("/Users/yash/Downloads/exercises/tilt-validator/utils/", "tiltdb.csv")
	tiltDBFile, err := os.OpenFile(tiltDBFilePath, os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		logError(fmt.Sprintf("Failed to open tiltdb.csv file: %v", err))
		return
	}
	tiltDBFile.Close()
}
