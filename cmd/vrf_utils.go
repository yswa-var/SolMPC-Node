package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"time"
)

type Validator struct {
	ID      string
	Name    string
	Stake   float64
	Active  bool
	VRFHash *big.Int
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

// VRFProof stores the proof of randomness generation
type VRFProof struct {
	PublicKey    ed25519.PublicKey
	Beta         []byte // The VRF output
	Proof        []byte // The proof that beta was correctly computed
	RandomNumber *big.Int
}

// GenerateVRF uses the validator's private key to generate a verifiable random value
// from a given input (e.g., block hash + round number)
func GenerateVRF(privKey ed25519.PrivateKey, pubKey ed25519.PublicKey, input []byte) (*VRFProof, error) {
	// Create a message by combining input data
	message := input

	// In a real implementation, this would use a specialized VRF algorithm
	// For demonstration, we're using ed25519 signatures as a simplified VRF
	signature := ed25519.Sign(privKey, message)

	// Hash the signature to get the VRF output (beta)
	beta := sha256.Sum256(signature)

	// Convert beta to a big.Int for easier comparisons
	randomNumber := new(big.Int).SetBytes(beta[:])

	// In a real VRF, we would include a proper cryptographic proof
	// Here we're just using the signature itself as the "proof"
	return &VRFProof{
		PublicKey:    pubKey,
		Beta:         beta[:],
		Proof:        signature,
		RandomNumber: randomNumber,
	}, nil
}

// VerifyVRF verifies that the VRF output was correctly computed
func VerifyVRF(proof *VRFProof, input []byte) (bool, error) {
	// In a real VRF, we'd verify the cryptographic proof
	// For our simplified example, we just verify the signature
	return ed25519.Verify(proof.PublicKey, input, proof.Proof), nil
}

// GenerateCommonSeed creates a deterministic seed for a specific round/block
func GenerateCommonSeed(blockHeight uint64, timestamp time.Time) []byte {
	seed := make([]byte, 16)
	binary.BigEndian.PutUint64(seed[0:8], blockHeight)
	binary.BigEndian.PutUint64(seed[8:16], uint64(timestamp.Unix()))
	return seed
}

// SelectValidator uses VRF proofs from all validators to deterministically select one
func SelectValidator(validators []Validator, vrfProofs []*VRFProof) *Validator {
	if len(validators) == 0 || len(vrfProofs) == 0 {
		logWarning("No validators or VRF proofs available")
		return nil
	}

	// Find the lowest VRF output
	lowestValue, _ := new(big.Int).SetString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)
	var selectedIndex int = -1

	for i, proof := range vrfProofs {
		// Skip invalid proofs (nil or zero)
		if proof == nil || proof.RandomNumber.Cmp(big.NewInt(0)) == 0 {
			continue
		}

		logInfo(fmt.Sprintf("Validator %s VRF value: %s", validators[i].ID, proof.RandomNumber.String()))

		// Compare VRF outputs and select the lowest
		if proof.RandomNumber.Cmp(lowestValue) < 0 {
			lowestValue = proof.RandomNumber
			selectedIndex = i
			logInfo(fmt.Sprintf("New lowest VRF found: %s (Validator %s)",
				lowestValue.String(), validators[i].ID))
		}
	}

	if selectedIndex == -1 {
		logError("No validator was selected")
		return nil
	}

	logSuccess(fmt.Sprintf("Chosen Validator: %s", validators[selectedIndex].ID))
	return &validators[selectedIndex]
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
