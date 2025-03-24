package distribution

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"tilt-valid/utils"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// Receiver represents the Receiver struct in the smart contract
type Receiver struct {
	Pubkey solana.PublicKey
	Amount uint64
}

// GeneratePDA generates a Program Derived Address (PDA) given a seed, sender public key, and program ID.
func GeneratePDA(seed []byte, senderPubkey solana.PublicKey, programID solana.PublicKey) (solana.PublicKey, error) {
	seeds := [][]byte{seed, senderPubkey[:]}
	pda, _, err := solana.FindProgramAddress(seeds, programID)
	return pda, err
}

// SerializeInitializeData serializes the data for the "initialize" instruction.
func SerializeInitializeData(businessRules [10]byte, receivers [10]Receiver, subTilts []string) ([]byte, error) {
	var data []byte

	// Add instruction discriminator
	hash := sha256.Sum256([]byte("global:initialize"))
	discriminator := hash[:8]
	data = append(data, discriminator...)

	// Serialize business_rules
	data = append(data, businessRules[:]...)

	// Serialize receivers
	for _, receiver := range receivers {
		data = append(data, receiver.Pubkey[:]...)
		amountBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(amountBytes, receiver.Amount)
		data = append(data, amountBytes...)
	}

	// Serialize sub_tilts
	subTiltsLen := uint32(len(subTilts))
	subTiltsLenBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(subTiltsLenBytes, subTiltsLen)
	data = append(data, subTiltsLenBytes...)
	for _, subTilt := range subTilts {
		subTiltBytes := []byte(subTilt)
		subTiltLen := uint32(len(subTiltBytes))
		subTiltLenBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(subTiltLenBytes, subTiltLen)
		data = append(data, subTiltLenBytes...)
		data = append(data, subTiltBytes...)
	}

	return data, nil
}

// GetInitializeAccounts returns the account metadata for the "initialize" instruction.
func GetInitializeAccounts(distributionPDA, receiverListPDA, subTiltListPDA, senderPubkey solana.PublicKey) []*solana.AccountMeta {
	return []*solana.AccountMeta{
		{PublicKey: distributionPDA, IsSigner: false, IsWritable: true},
		{PublicKey: receiverListPDA, IsSigner: false, IsWritable: true},
		{PublicKey: subTiltListPDA, IsSigner: false, IsWritable: true},
		{PublicKey: senderPubkey, IsSigner: true, IsWritable: true},
		{PublicKey: solana.SystemProgramID, IsSigner: false, IsWritable: false},
	}
}

// CreateInitializeInstruction creates the "initialize" instruction for the Solana program.
func CreateInitializeInstruction(programID, senderPubkey solana.PublicKey, businessRules [10]byte, receivers [10]Receiver, subTilts []string) (solana.Instruction, error) {
	// Generate PDAs
	distributionPDA, err := GeneratePDA([]byte("distribution"), senderPubkey, programID)
	if err != nil {
		return nil, err
	}
	receiverListPDA, err := GeneratePDA([]byte("receiver_list"), senderPubkey, programID)
	if err != nil {
		return nil, err
	}
	subTiltListPDA, err := GeneratePDA([]byte("sub_tilt_list"), senderPubkey, programID)
	if err != nil {
		return nil, err
	}

	// Serialize instruction data
	data, err := SerializeInitializeData(businessRules, receivers, subTilts)
	if err != nil {
		return nil, err
	}

	// Get accounts
	accounts := GetInitializeAccounts(distributionPDA, receiverListPDA, subTiltListPDA, senderPubkey)

	// Create instruction
	instruction := solana.NewInstruction(programID, accounts, data)
	return instruction, nil
}

// SendTransaction creates, signs, and sends a transaction with the given instructions.
func SendTransaction(client *rpc.Client, instructions []solana.Instruction, senderPrivateKey solana.PrivateKey) (solana.Signature, error) {
	senderPubkey := senderPrivateKey.PublicKey()

	// Get the latest blockhash
	recent, err := client.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return solana.Signature{}, err
	}

	// Create the transaction
	tx, err := solana.NewTransaction(
		instructions,
		recent.Value.Blockhash,
		solana.TransactionPayer(senderPubkey),
	)
	if err != nil {
		return solana.Signature{}, err
	}

	// Sign the transaction
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(senderPubkey) {
			return &senderPrivateKey
		}
		return nil
	})
	if err != nil {
		return solana.Signature{}, err
	}

	// Send the transaction
	sig, err := client.SendTransaction(context.Background(), tx)
	if err != nil {
		return solana.Signature{}, err
	}

	return sig, nil
}

// DistributionData represents the node data with float64 for amounts
type DistributionData struct {
	Amount        float64   `json:"amount"`
	BusinessRules []float64 `json:"business_rules"`
	ID            int       `json:"id"`
	Receiver      []string  `json:"receiver"`
	Sender        string    `json:"sender"`
	Subtilt       []int     `json:"subtilt"`
}

// Allocation defines the output format
type Allocation struct {
	Receiver string
	Amount   float64
}

// Distribution parses input into DistributionData
func Distribution(distBytes map[string]any) (DistributionData, error) {
	receivers, ok := distBytes["receiver"].([]string)
	if !ok {
		return DistributionData{}, fmt.Errorf("invalid receiver: expected []string")
	}
	senderStr, ok := distBytes["sender"].(string)
	if !ok {
		return DistributionData{}, fmt.Errorf("invalid sender: expected string")
	}
	amount, ok := distBytes["amount"].(float64)
	if !ok {
		// Try int conversion as a fallback
		if amtInt, ok := distBytes["amount"].(int); ok {
			amount = float64(amtInt)
		} else {
			return DistributionData{}, fmt.Errorf("invalid amount: expected float64")
		}
	}
	businessRules, ok := distBytes["business_rules"].([]float64)
	if !ok {
		// Try []int conversion as a fallback
		if brInt, ok := distBytes["business_rules"].([]int); ok {
			businessRules = make([]float64, len(brInt))
			for i, v := range brInt {
				businessRules[i] = float64(v)
			}
		} else {
			return DistributionData{}, fmt.Errorf("invalid business_rules: expected []float64")
		}
	}
	id, ok := distBytes["id"].(int)
	if !ok {
		return DistributionData{}, fmt.Errorf("invalid id: expected int")
	}
	subtilt, ok := distBytes["subtilt"].([]int)
	if !ok {
		return DistributionData{}, fmt.Errorf("invalid subtilt: expected []int")
	}

	return DistributionData{
		Amount:        amount,
		BusinessRules: businessRules,
		ID:            id,
		Receiver:      receivers,
		Sender:        senderStr,
		Subtilt:       subtilt,
	}, nil
}

// AllocateAmounts distributes amounts across the hierarchy
func AllocateAmounts(data DistributionData) ([]Allocation, error) {
	// Handle empty BusinessRules for leaf nodes
	if len(data.BusinessRules) == 0 {
		if len(data.Subtilt) == 0 && len(data.Receiver) > 0 {
			data.BusinessRules = []float64{100.0} // Default for leaf node with receivers
		} else {
			return nil, fmt.Errorf("invalid business rules: empty rules not allowed for node ID %d with %d subtilts", data.ID, len(data.Subtilt))
		}
	}

	// // Validate business rules length
	// if len(data.BusinessRules) != len(data.Subtilt)+1 {
	//     return nil, fmt.Errorf("invalid business rules: length mismatch for node ID %d, expected %d, got %d", data.ID, len(data.Subtilt)+1, len(data.BusinessRules))
	// }

	// Validate total percentage
	totalPercentage := 0.0
	for _, percentage := range data.BusinessRules {
		if percentage < 0 {
			return nil, fmt.Errorf("invalid business rules: negative percentage %f for node ID %d", percentage, data.ID)
		}
		totalPercentage += percentage
	}
	if totalPercentage != 100.0 {
		return nil, fmt.Errorf("invalid business rules: total percentage must be 100 for node ID %d, got %f", data.ID, totalPercentage)
	}

	// Validate receiver allocation
	if len(data.Receiver) == 0 && data.BusinessRules[0] != 0 {
		return nil, fmt.Errorf("invalid business rules: cannot allocate %f%% to non-existent receivers for node ID %d", data.BusinessRules[0], data.ID)
	}

	// Use a map to accumulate amounts
	resultMap := make(map[string]float64)

	// Allocate to receivers
	receiverAmount := data.Amount * (data.BusinessRules[0] / 100.0)
	if len(data.Receiver) > 0 {
		share := receiverAmount / float64(len(data.Receiver))
		for _, receiver := range data.Receiver {
			resultMap[receiver] += share
		}
	}

	// Allocate to sub-tilts
	for i, subID := range data.Subtilt {
		subAmount := data.Amount * (data.BusinessRules[i+1] / 100.0)
		subData, err := utils.ReadTiltDataByID(subID)
		if err != nil {
			return nil, fmt.Errorf("failed to read sub-tilt data for ID %d: %w", subID, err)
		}

		// Construct subDistData with type safety
		subDistData, err := Distribution(subData)
		if err != nil {
			return nil, fmt.Errorf("invalid sub-tilt data for ID %d: %w", subID, err)
		}
		subDistData.Amount += subAmount // Add parent's contribution

		// Recursive call
		subResult, err := AllocateAmounts(subDistData)
		if err != nil {
			return nil, fmt.Errorf("failed to allocate amounts for sub-tilt ID %d: %w", subID, err)
		}
		for _, alloc := range subResult {
			resultMap[alloc.Receiver] += alloc.Amount
		}
	}

	// Convert map to list
	var allocations []Allocation
	for receiver, amount := range resultMap {
		allocations = append(allocations, Allocation{Receiver: receiver, Amount: amount})
	}

	return allocations, nil
}
