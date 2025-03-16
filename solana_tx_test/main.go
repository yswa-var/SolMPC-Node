package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/program/system"
	"github.com/blocto/solana-go-sdk/types"
)

// Load keypair from file
func loadKeypair(path string) (types.Account, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return types.Account{}, err
	}

	var keypair []byte
	if err := json.Unmarshal(file, &keypair); err != nil {
		return types.Account{}, err
	}

	return types.AccountFromBytes(keypair)
}

func main() {
	// Initialize Solana Client
	c := client.NewClient("https://api.devnet.solana.com")

	// Load sender's wallet (validator keypair)
	// fromAccount, err := loadKeypair("validator-keypair.json")
	fromAccount, err := loadKeypair("/Users/peerfaheem/Documents/GitHub/tilt-validator/validator-keypair.json")

	if err != nil {
		log.Fatalf("Failed to load keypair: %v", err)
	}

	// Derive PDA using correct seeds
	pda, _, err := common.FindProgramAddress(
		[][]byte{[]byte("tilt-state"), []byte("1")}, // Adjust seeds as needed
		common.PublicKeyFromString("DwcQ8jLY9yhdzGRU5VyyRhYygLok8DSTeEdriS4wxUiS"),
	)
	if err != nil {
		log.Fatalf("Failed to derive PDA: %v", err)
	}

	// Use the PDA as the receiver
	toPubKey := pda

	// Fetch latest blockhash
	recentBlockhashResp, err := c.GetLatestBlockhash(context.Background())
	if err != nil {
		log.Fatalf("Failed to get blockhash: %v", err)
	}

	// Construct a versioned transaction message
	message := types.NewMessage(types.NewMessageParam{
		FeePayer:        fromAccount.PublicKey,
		RecentBlockhash: recentBlockhashResp.Blockhash,
		Instructions: []types.Instruction{
			{
				ProgramID: common.SystemProgramID,
				Accounts: []types.AccountMeta{
					{PubKey: fromAccount.PublicKey, IsSigner: true, IsWritable: true}, // Sender must be writable
					{PubKey: toPubKey, IsSigner: false, IsWritable: true},             // PDA must be writable!
				},
				Data: system.Transfer(system.TransferParam{
					From:   fromAccount.PublicKey,
					To:     toPubKey,
					Amount: 1_000_000, // Amount in lamports
				}).Data,
			},
		},
	})

	// Create a versioned transaction
	tx, err := types.NewTransaction(types.NewTransactionParam{
		Signers: []types.Account{fromAccount},
		Message: message,
	})
	if err != nil {
		log.Fatalf("Failed to create transaction: %v", err)
	}

	// Send the transaction
	txSig, err := c.SendTransaction(context.Background(), tx)
	if err != nil {
		log.Fatalf("Failed to send transaction: %v", err)
	}

	fmt.Println("Versioned transaction submitted successfully:", txSig)

	// Verify the transaction signature
	// txDetails, err := c.GetTransaction(context.Background(), txSig)
	// if err != nil {
	// 	log.Fatalf("Failed to fetch transaction details: %v", err)
	// }

	// if len(txDetails.Transaction.Signatures) > 0 {
	// 	fmt.Println("Transaction signature verified:", txDetails.Transaction.Signatures[0])
	// } else {
	// 	fmt.Println("Transaction signature verification failed!")
	// }
	// Verify the transaction signature
	txDetails, err := c.GetTransaction(context.Background(), txSig)
	if err != nil {
		log.Fatalf("Failed to fetch transaction details: %v", err)
	}

	// ✅ Add nil check to avoid runtime errors
	if txDetails == nil || len(txDetails.Transaction.Signatures) == 0 {
		log.Fatalf("Transaction details are nil. The transaction might still be processing.")
	}

	if len(txDetails.Transaction.Signatures) > 0 {
		fmt.Println("Transaction signature verified:", txDetails.Transaction.Signatures[0])
	} else {
		fmt.Println("Transaction signature verification failed!")
	}

}
