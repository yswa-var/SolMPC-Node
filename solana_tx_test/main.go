package main

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/blocto/solana-go-sdk/types"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func main() {
	// Step 1: Connect to Solana Devnet
	client := rpc.New("https://api.devnet.solana.com")

	// Step 2: Define Program ID
	// Use the program ID from your deployed Rust contract
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

	// Step 3: Define Recipients and Payment Details
	total_amount := uint64(500)
	recipients := []solana.PublicKey{
		solana.NewWallet().PublicKey(),
		solana.NewWallet().PublicKey(),
		solana.NewWallet().PublicKey(),
		solana.NewWallet().PublicKey(),
		solana.NewWallet().PublicKey(),
	}
	amounts := []uint64{100, 100, 100, 100, 100}

	// Step 4: Serialize Instruction Data
	instructionData, err := serializeInstructionData(amounts, total_amount, recipients)
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

	// Step 10: Send Transaction
	sig, err := client.SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{
		SkipPreflight:       false,
		PreflightCommitment: rpc.CommitmentFinalized,
	})
	if err != nil {
		log.Fatalf("Failed to send transaction: %v", err)
	}
	fmt.Printf("Transaction sent! Signature: %s\n", sig)
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
