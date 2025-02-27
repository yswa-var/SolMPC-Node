// # Generate 5 validators and save to validators.csv
// go run main.go generate 5 validators.csv

// # Sign a message using validators from validators.csv
// go run main.go sign validators.csv "Hello, Solana!" signatures.csv

// # Validate and aggregate signatures
// go run main.go aggregate signatures.csv aggregated.txt

package signature

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
)

// Constants
const (
	THRESHOLD = 3 // Number of validators required for threshold
)

// Validator represents a node in the system
type Validator struct {
	ID         int
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
}

// PartialSignature represents a signature from a single validator
type PartialSignature struct {
	ValidatorID int
	Signature   []byte
	PublicKey   []byte
	Message     []byte
}

// LagrangeCoefficient calculates Lagrange coefficients for TSS
func LagrangeCoefficient(indices []int, idx int, prime *big.Int) *big.Int {
	result := big.NewInt(1)
	idxBig := big.NewInt(int64(idx))

	for _, j := range indices {
		if j == idx {
			continue
		}

		jBig := big.NewInt(int64(j))
		num := big.NewInt(0).Sub(big.NewInt(0), jBig)
		den := big.NewInt(0).Sub(idxBig, jBig)
		den.Mod(den, prime)
		denInv := big.NewInt(0).ModInverse(den, prime)

		if denInv == nil {
			log.Fatalf("Error: cannot compute ModInverse for denominator %v", den)
		}

		term := big.NewInt(0).Mul(num, denInv)
		term.Mod(term, prime)
		result.Mul(result, term)
		result.Mod(result, prime)
	}

	return result
}

// GenerateValidators creates validators and saves keys to CSV
func GenerateValidators(count int, outputFile string) []Validator {
	validators := make([]Validator, count)

	file, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Could not create validator file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"ID", "PrivateKey", "PublicKey"})

	for i := 0; i < count; i++ {
		pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			log.Fatalf("Failed to generate key pair: %v", err)
		}

		validators[i] = Validator{
			ID:         i + 1,
			PrivateKey: privKey,
			PublicKey:  pubKey,
		}

		// Write to CSV
		writer.Write([]string{
			strconv.Itoa(validators[i].ID),
			hex.EncodeToString(validators[i].PrivateKey),
			hex.EncodeToString(validators[i].PublicKey),
		})
	}

	fmt.Printf("Generated %d validators and saved to %s\n", count, outputFile)
	return validators
}

// LoadValidators loads validators from CSV file
func LoadValidators(inputFile string) []Validator {
	file, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("Could not open validator file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}

	validators := make([]Validator, 0, len(records)-1)

	// Skip header
	for i := 1; i < len(records); i++ {
		id, err := strconv.Atoi(records[i][0])
		if err != nil {
			log.Fatalf("Invalid ID: %v", err)
		}

		privKeyHex := records[i][1]
		pubKeyHex := records[i][2]

		privKey, err := hex.DecodeString(privKeyHex)
		if err != nil {
			log.Fatalf("Invalid private key: %v", err)
		}

		pubKey, err := hex.DecodeString(pubKeyHex)
		if err != nil {
			log.Fatalf("Invalid public key: %v", err)
		}

		validators = append(validators, Validator{
			ID:         id,
			PrivateKey: privKey,
			PublicKey:  pubKey,
		})
	}

	fmt.Printf("Loaded %d validators from %s\n", len(validators), inputFile)
	return validators
}

// SignMessage signs a message with a validator's private key
func SignMessage(validator Validator, message []byte) PartialSignature {
	signature := ed25519.Sign(validator.PrivateKey, message)

	return PartialSignature{
		ValidatorID: validator.ID,
		Signature:   signature,
		PublicKey:   validator.PublicKey,
		Message:     message,
	}
}

// SaveSignatures saves signatures to a CSV file
func SaveSignatures(signatures []PartialSignature, outputFile string) {
	file, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Could not create signatures file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"ValidatorID", "Signature", "PublicKey", "Message"})

	for _, sig := range signatures {
		writer.Write([]string{
			strconv.Itoa(sig.ValidatorID),
			hex.EncodeToString(sig.Signature),
			hex.EncodeToString(sig.PublicKey),
			hex.EncodeToString(sig.Message),
		})
	}

	fmt.Printf("Saved %d signatures to %s\n", len(signatures), outputFile)
}

// LoadSignatures loads signatures from a CSV file
func LoadSignatures(inputFile string) []PartialSignature {
	file, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("Could not open signatures file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}

	signatures := make([]PartialSignature, 0, len(records)-1)

	// Skip header
	for i := 1; i < len(records); i++ {
		id, err := strconv.Atoi(records[i][0])
		if err != nil {
			log.Fatalf("Invalid ID: %v", err)
		}

		sigHex := records[i][1]
		pubKeyHex := records[i][2]
		messageHex := records[i][3]

		sig, err := hex.DecodeString(sigHex)
		if err != nil {
			log.Fatalf("Invalid signature: %v", err)
		}

		pubKey, err := hex.DecodeString(pubKeyHex)
		if err != nil {
			log.Fatalf("Invalid public key: %v", err)
		}

		message, err := hex.DecodeString(messageHex)
		if err != nil {
			log.Fatalf("Invalid message: %v", err)
		}

		signatures = append(signatures, PartialSignature{
			ValidatorID: id,
			Signature:   sig,
			PublicKey:   pubKey,
			Message:     message,
		})
	}

	fmt.Printf("Loaded %d signatures from %s\n", len(signatures), inputFile)
	return signatures
}

// ValidateSignature validates a single signature
func ValidateSignature(signature PartialSignature) bool {
	isValid := ed25519.Verify(signature.PublicKey, signature.Message, signature.Signature)
	return isValid
}

// ValidateSignatures validates multiple signatures
func ValidateSignatures(signatures []PartialSignature) []PartialSignature {
	validSignatures := make([]PartialSignature, 0)

	for _, sig := range signatures {
		if ValidateSignature(sig) {
			validSignatures = append(validSignatures, sig)
			fmt.Printf("Signature from validator %d is valid\n", sig.ValidatorID)
		} else {
			fmt.Printf("Signature from validator %d is invalid\n", sig.ValidatorID)
		}
	}

	return validSignatures
}

// AggregateSignatures combines multiple signatures into one
// FIXED VERSION: Fixed the buffer overflow issue
func AggregateSignatures(signatures []PartialSignature) ([]byte, []byte, error) {
	if len(signatures) < THRESHOLD {
		return nil, nil, fmt.Errorf("need at least %d signatures, got %d", THRESHOLD, len(signatures))
	}

	// Ensure all signatures are for the same message
	message := signatures[0].Message
	for _, sig := range signatures {
		if !strings.EqualFold(hex.EncodeToString(sig.Message), hex.EncodeToString(message)) {
			return nil, nil, fmt.Errorf("all signatures must be for the same message")
		}
	}

	// Get validator indices for Lagrange coefficient calculation
	indices := make([]int, len(signatures))
	for i, sig := range signatures {
		indices[i] = sig.ValidatorID
	}

	// Ed25519 curve order (l)
	prime, _ := new(big.Int).SetString("7237005577332262213973186563042994240857116359379907606001950938285454250989", 10)

	// We're using a simplified approach for demonstration
	// In a real implementation, we would combine the signatures properly using the TSS algorithm

	// This is the fixed part: create a buffer of the appropriate size
	combinedSignature := make([]byte, ed25519.SignatureSize)

	// Create a combined public key (just for demonstration - not cryptographically valid)
	combinedPubKey := make([]byte, ed25519.PublicKeySize)
	copy(combinedPubKey, signatures[0].PublicKey)

	// Hash to combine signatures with Lagrange weights
	// Note: This is a simplified approach for demonstration
	h := sha256.New()

	for i, sig := range signatures[:THRESHOLD] {
		coef := LagrangeCoefficient(indices[:THRESHOLD], sig.ValidatorID, prime)
		coefBytes := coef.Bytes()

		h.Write(sig.Signature)
		h.Write(coefBytes)

		fmt.Println(i) // Debug line to show progress
	}

	finalHash := h.Sum(nil)

	// Fix: Make sure we only copy up to the minimum of the length of finalHash and the length of combinedSignature
	copyLength := min(len(finalHash), len(combinedSignature))
	copy(combinedSignature, finalHash[:copyLength])

	fmt.Println("Created aggregated signature")
	return combinedSignature, combinedPubKey, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ValidatorSignMessage demonstrates a validator signing a message
func ValidatorSignMessage(validatorFile string, messageStr string, outputFile string) {
	validators := LoadValidators(validatorFile)
	message := []byte(messageStr)

	signatures := make([]PartialSignature, len(validators))
	for i, validator := range validators {
		signatures[i] = SignMessage(validator, message)
		fmt.Printf("Validator %d signed the message\n", validator.ID)
	}

	SaveSignatures(signatures, outputFile)
}

// AggregatorValidateAndMerge demonstrates an aggregator validating and merging signatures
func AggregatorValidateAndMerge(signaturesFile string, outputFile string) {
	signatures := LoadSignatures(signaturesFile)

	// Validate signatures
	validSignatures := ValidateSignatures(signatures)
	fmt.Printf("Found %d valid signatures out of %d\n", len(validSignatures), len(signatures))

	if len(validSignatures) < THRESHOLD {
		fmt.Printf("Cannot aggregate: need at least %d valid signatures\n", THRESHOLD)
		return
	}

	// Aggregate signatures
	aggSignature, combinedPubKey, err := AggregateSignatures(validSignatures[:THRESHOLD])
	if err != nil {
		log.Fatalf("Failed to aggregate signatures: %v", err)
	}

	// Save the aggregated result
	result := fmt.Sprintf("Message: %s\nAggregated Signature: %s\nCombined Public Key: %s\n",
		hex.EncodeToString(validSignatures[0].Message),
		hex.EncodeToString(aggSignature),
		hex.EncodeToString(combinedPubKey))

	err = ioutil.WriteFile(outputFile, []byte(result), 0644)
	if err != nil {
		log.Fatalf("Failed to save aggregated result: %v", err)
	}

	fmt.Printf("Saved aggregated result to %s\n", outputFile)
}

func main() {
	// Check command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Available commands:")
		fmt.Println("  generate <count> <output_file> - Generate validators")
		fmt.Println("  sign <validator_file> <message> <output_file> - Sign a message")
		fmt.Println("  aggregate <signatures_file> <output_file> - Validate and aggregate signatures")
		return
	}

	command := os.Args[1]

	switch command {
	case "generate":
		if len(os.Args) < 4 {
			fmt.Println("Usage: generate <count> <output_file>")
			return
		}
		count, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("Invalid count: %v", err)
		}
		outputFile := os.Args[3]
		GenerateValidators(count, outputFile)

	case "sign":
		if len(os.Args) < 5 {
			fmt.Println("Usage: sign <validator_file> <message> <output_file>")
			return
		}
		validatorFile := os.Args[2]
		message := os.Args[3]
		outputFile := os.Args[4]
		ValidatorSignMessage(validatorFile, message, outputFile)

	case "aggregate":
		if len(os.Args) < 4 {
			fmt.Println("Usage: aggregate <signatures_file> <output_file>")
			return
		}
		signaturesFile := os.Args[2]
		outputFile := os.Args[3]
		AggregatorValidateAndMerge(signaturesFile, outputFile)

	default:
		fmt.Println("Unknown command:", command)
	}
}
