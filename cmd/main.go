package main

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"tilt-valid/cmd/config"
	"tilt-valid/internal/distribution"
	"tilt-valid/internal/exchange"
	mpc "tilt-valid/internal/mpc"
	"tilt-valid/utils"

	"github.com/blocto/solana-go-sdk/types"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// Define a flag for selecting the tilt type
var tiltType string

func init() {
	flag.StringVar(&tiltType, "tilt-type", "two_subtilts", "Type of tilt to create (simple, one_subtilt, two_subtilts, nested)")
}

const threshold = 2

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

	// Initialize RPC client for Devnet
	client := rpc.New("https://api.devnet.solana.com")

	programID, err := solana.PublicKeyFromBase58("EM7AAngMgQPXizeuwAKaBvci79DhRxJMBYjRVoJWYEH3")
	if err != nil {
		log.Fatalf("Invalid program ID: %v", err)
	}

	// Payer setup (replace with your actual keypair)
	privateKeyBytes := [64]byte{16, 246, 168, 249, 237, 255, 125, 101, 217, 247, 127, 166, 74, 53, 162, 51, 171, 210, 214, 143, 114, 231, 90, 39, 199, 152, 51, 247, 155, 89, 49, 209, 188, 164, 235, 18, 201, 90, 220, 112, 187, 42, 70, 106, 82, 127, 58, 134, 94, 39, 122, 20, 109, 110, 8, 203, 126, 148, 192, 140, 5, 77, 75, 60}
	wallet, err := types.AccountFromBytes(privateKeyBytes[:])
	if err != nil {
		log.Fatalf("Failed to create wallet from private key: %v", err)
	}

	// loading config
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("error in loading config")
	}
	path := cfg.ValidatorPath

	// Check for tilt-type flag in args
	for _, arg := range args {
		if len(arg) > 11 && arg[:11] == "--tilt-type=" {
			tiltType = arg[11:]
			break
		}
	}
	if id == 1 {
		// Create a new tilt
		utils.GetTestTilt(cfg.TiltDb, tiltType)
	}

	validatorsFilePath := filepath.Join(path, "validators.csv")
	validators, err := loadValidators(validatorsFilePath)
	if err != nil {
		logError(fmt.Sprintf("Error loading validators: %v", err))
		return
	}
	// transaction creation successfully created.

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

	// TODO: 2 simplify in required format
	tilt, _ := utils.ReadTiltData(cfg.TiltDb)
	convertedTilt := make(map[string]map[string]interface{})
	for key, value := range tilt {
		if innerMap, ok := value.(map[string]interface{}); ok {
			convertedTilt[key] = innerMap
		} else {
			logError(fmt.Sprintf("Invalid type for key %s in tilt data", key))
			return
		}
	}
	allocation, err := distribution.AllocateAmounts(convertedTilt, "1")
	if err != nil {
		logError(fmt.Sprintf("Error allocating amounts: %v", err))
		return
	}

	var amounts []uint64
	var totalAmount uint64
	var recipients []solana.PublicKey

	for _, alloc := range allocation {
		amounts = append(amounts, uint64(alloc.Amount))
		totalAmount += uint64(alloc.Amount)

		// Trim spaces and check for invalid characters
		receiver := strings.TrimSpace(alloc.Receiver)

		pubKey, err := solana.PublicKeyFromBase58(receiver)
		if err != nil {
			logError(fmt.Sprintf("Invalid public key: %v (Receiver: %s)", err, receiver))
			return
		}
		recipients = append(recipients, pubKey)
	}

	// Step 4: Serialize Instruction Data
	instructionData, err := serializeInstructionData(amounts, totalAmount, recipients)
	if err != nil {
		log.Fatalf("Failed to serialize instruction data: %v", err)
	}

	// Step 5: Prepare Accounts
	accounts := []*solana.AccountMeta{
		{PublicKey: solana.PublicKeyFromBytes(wallet.PublicKey[:]), IsSigner: true, IsWritable: true}, // Sender
	}

	// Add recipient accounts
	for _, recipient := range recipients {
		accounts = append(accounts, &solana.AccountMeta{
			PublicKey:  recipient,
			IsSigner:   false,
			IsWritable: true,
		})
	}

	// Step 6: Create Instruction
	instruction := solana.NewInstruction(programID, accounts, instructionData)
	// Step 7: Get Recent Blockhash
	ctx := context.Background()
	recent, err := client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		log.Fatalf("Failed to get recent blockhash: %v", err)
	}

	// Step 8: Build Transaction
	tx, err := solana.NewTransaction(
		[]solana.Instruction{instruction},
		recent.Value.Blockhash,
		solana.TransactionPayer(solana.PublicKeyFromBytes(wallet.PublicKey[:])),
	)
	if err != nil {
		log.Fatalf("Failed to create transaction: %v", err)
	}

	// Step 9: Sign Transaction
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(solana.PublicKeyFromBytes(wallet.PublicKey[:])) {
			privateKey := solana.PrivateKey(wallet.PrivateKey)
			return &privateKey
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to sign transaction: %v", err)
	}

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

	instructionBytes, err := json.Marshal(instruction)
	if err != nil {
		logError(fmt.Sprintf("Failed to marshal instruction: %v", err))
		return
	}
	msgToSign, err := json.Marshal(instructionBytes)
	if err != nil {
		logError(fmt.Sprintf("Failed to marshal msgToSign: %v", err))
		return
	}
	digestMsg := mpc.Digest(msgToSign) // if we use msgToSign the signing process works fine but n using instructions we get error

	separator("Signing Process")
	wg.Add(1)
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
	time.Sleep(3 * time.Second)

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
		if id == selectedValidator {
			logSuccess(fmt.Sprintf("This validator (ID: %d) was selected for verification!", 1))
			separator("Signature Verification")
			if ed25519.Verify(pk, digestMsg, sign) {
				logSuccess("✅ Signature verification successful!")
			} else {
				logError("❌ Signature verification failed!")
			}

			// Step 10: Send Transaction
			sig, err := client.SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{
				SkipPreflight:       false,
				PreflightCommitment: rpc.CommitmentFinalized,
			})
			if err != nil {
				log.Fatalf("Failed to send transaction: %v", err)
			}
			fmt.Printf("Transaction sent! Signature: %s\n", sig)

		} else {
			logInfo(fmt.Sprintf("Validator ID: %d was selected for verification", selectedValidator))
		}
	}
}

// serializeInstructionData creates the instruction data for validate_payment_distribution
func serializeInstructionData(amounts []uint64, totalAmount uint64, recipients []solana.PublicKey) ([]byte, error) {
	var data []byte

	// Discriminator: First 8 bytes of SHA256("global:validate_payment_distribution")
	hash := sha256.Sum256([]byte("global:validate_payment_distribution"))
	data = append(data, hash[:8]...)

	// Serialize total_amount (u64)
	totalBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(totalBytes, totalAmount)
	data = append(data, totalBytes...)

	// Serialize receivers (Vec<Pubkey>)
	// Length of receivers
	lengthBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(lengthBytes, uint32(len(recipients)))
	data = append(data, lengthBytes...)

	// Add each receiver's public key (32 bytes)
	for _, recipient := range recipients {
		data = append(data, recipient.Bytes()...)
	}

	// Serialize amounts (Vec<u64>)
	// Length of amounts
	amountLengthBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(amountLengthBytes, uint32(len(amounts)))
	data = append(data, amountLengthBytes...)

	// Add each amount
	for _, amount := range amounts {
		amountBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(amountBytes, amount)
		data = append(data, amountBytes...)
	}

	return data, nil
}
