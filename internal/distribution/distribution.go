package distribution

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"tilt-valid/utils"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// TiltDistribution manages payment distribution
type TiltDistribution struct {
	client *rpc.Client
}

// NewTiltDistribution initializes a new distributor
func NewTiltDistribution(endpoint string) (*TiltDistribution, error) {
	client := rpc.New(endpoint)
	// Test the endpoint
	_, err := client.GetVersion(context.Background())
	if err != nil {
		return nil, fmt.Errorf("RPC endpoint %s unavailable: %v", endpoint, err)
	}
	return &TiltDistribution{client: client}, nil
}

// Distribute executes the distribution on-chain
func (td *TiltDistribution) Distribute(sender *solana.PrivateKey, tiltMint solana.PublicKey, distributions map[solana.PublicKey]uint64) (solana.Signature, error) {
	programID := solana.MustPublicKeyFromBase58("8ctzNPg4MzNDu3A8BxpBmP34NCEH5emAtgtESuGq3tfN")
	distributionAccount, bump, err := solana.FindProgramAddress([][]byte{[]byte("distribution")}, programID)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to derive distribution account: %v", err)
	}
	senderTokenAccount := solana.MustPublicKeyFromBase58("4gVrPcoM9iidqcvQR1ZXNy2DM3B6Lx647xTY1MNLudKY")

	recent, err := td.client.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to get recent blockhash: %v", err)
	}

	// Initialize distribution account
	initData := append([]byte{0}, bump)
	initAccounts := []*solana.AccountMeta{
		{PublicKey: distributionAccount, IsSigner: false, IsWritable: true},
		{PublicKey: tiltMint, IsSigner: false, IsWritable: false},
		{PublicKey: sender.PublicKey(), IsSigner: true, IsWritable: true},
		{PublicKey: solana.SystemProgramID, IsSigner: false, IsWritable: false},
	}
	initInstr := solana.NewInstruction(programID, initAccounts, initData)

	var instructions []solana.Instruction
	instructions = append(instructions, initInstr)

	for recipientToken, amount := range distributions {
		discriminator := sha256.Sum256([]byte("global:distribute_tilt"))
		data := make([]byte, 8+32+8)
		copy(data[0:8], discriminator[:8])
		copy(data[8:40], recipientToken[:])
		binary.LittleEndian.PutUint64(data[40:48], amount)
		accounts := []*solana.AccountMeta{
			{PublicKey: distributionAccount, IsSigner: false, IsWritable: false},
			{PublicKey: tiltMint, IsSigner: false, IsWritable: false},
			{PublicKey: senderTokenAccount, IsSigner: false, IsWritable: true},
			{PublicKey: recipientToken, IsSigner: false, IsWritable: true},
			{PublicKey: sender.PublicKey(), IsSigner: true, IsWritable: false},
			{PublicKey: solana.TokenProgramID, IsSigner: false, IsWritable: false},
			{PublicKey: solana.SystemProgramID, IsSigner: false, IsWritable: false},
		}
		instructions = append(instructions, solana.NewInstruction(programID, accounts, data))
	}

	tx, err := solana.NewTransaction(instructions, recent.Value.Blockhash, solana.TransactionPayer(sender.PublicKey()))
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to create transaction: %v", err)
	}

	// Serialize transaction for hash
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to serialize transaction: %v", err)
	}
	hash := sha256.Sum256(txBytes)
	txHash := hex.EncodeToString(hash[:])

	fmt.Printf("Transaction Hash for Validation: %s\n", txHash)

	// Sign the transaction
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(sender.PublicKey()) {
			return sender
		}
		return nil
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to sign transaction: %v", err)
	}

	// Send the transaction with verbose error handling
	sig, err := td.client.SendTransaction(context.Background(), tx)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to send transaction: %v", err)
	}

	return sig, nil
}

// flattenTilt recursively flattens the tilt tree into a payment map
func FlattenTilt(t utils.Tilt, totalAmount uint64) (map[solana.PublicKey]uint64, error) {
	distributions := make(map[solana.PublicKey]uint64)
	if len(t.SubTilts) == 0 {
		if len(t.Recipients) == 0 {
			return nil, fmt.Errorf("no recipients in leaf tilt %s", t.Name)
		}
		for _, recipient := range t.Recipients {
			distributions[recipient] = totalAmount / uint64(len(t.Recipients))
		}
		return distributions, nil
	}

	weightSum := uint8(0)
	for _, w := range t.Weights {
		weightSum += w
	}
	if weightSum != 100 {
		return nil, fmt.Errorf("weights must sum to 100 in %s, got %d", t.Name, weightSum)
	}

	subTiltWeight := t.Weights[0]
	subTiltAmount := totalAmount * uint64(subTiltWeight) / 100
	localAmount := totalAmount - subTiltAmount

	subTiltCount := len(t.SubTilts)
	if subTiltCount > 0 {
		perSubTiltAmount := subTiltAmount / uint64(subTiltCount)
		for _, subTilt := range t.SubTilts {
			subDist, err := FlattenTilt(subTilt, perSubTiltAmount)
			if err != nil {
				return nil, err
			}
			for k, v := range subDist {
				distributions[k] = v
			}
		}
	}

	if len(t.Recipients) > 0 {
		for _, recipient := range t.Recipients {
			distributions[recipient] = localAmount / uint64(len(t.Recipients))
		}
	}

	return distributions, nil
}

// simulateValidator runs the validator simulation script
// func simulateValidator(txHash string) (string, error) {
// 	cmd := exec.Command("bash", "../validator_sim.sh") // Adjust path
// 	cmd.Env = append(cmd.Env, fmt.Sprintf("TX_HASH=%s", txHash))
// 	output, err := cmd.CombinedOutput()
// 	if err != nil {
// 		return "", fmt.Errorf("validator simulation failed: %v, output: %s", err, output)
// 	}
// 	return string(output), nil
// }
