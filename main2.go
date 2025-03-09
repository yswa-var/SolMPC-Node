package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	mpc "teddsa/eddsa"
	"teddsa/exchange"

	"time"

	"go.uber.org/zap"
)

const threshold = 1

// Validator represents a validator with an ID and name.
type Validator struct {
	ID   string
	Name string
}

func Digest(in []byte) []byte {
	h := sha256.New()
	h.Write(in)
	return h.Sum(nil)
}

func createSolanaTransaction() string {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		fmt.Println("Failed to generate random bytes:", err)
		return ""
	}
	hash := sha256.Sum256(randomBytes)

	return fmt.Sprintf("%x", hash)
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
		if len(record) != 2 {
			return nil, fmt.Errorf("invalid record: %v", record)
		}
		validators = append(validators, Validator{ID: record[0], Name: record[1]})
	}

	return validators, nil
}

func loadShareData(p *mpc.Party, partyID int) error {
	fileName := fmt.Sprintf("localsavedata_eddsa%d", partyID-1) // Construct the file name
	data, err := os.ReadFile(fileName)                          // Read the file content
	if err != nil {
		return fmt.Errorf("failed to read key share file %s: %w", fileName, err)
	}

	err = p.SetShareData(data)
	if err != nil {
		return fmt.Errorf("failed to set share data: %w", err)
	}

	fmt.Println("Successfully loaded and set share data for party", partyID)
	return nil
}

func main() {
	args := os.Args[1:]

	wg := sync.WaitGroup{}

	if len(args) < 1 {
		fmt.Println("Usage: main2 <validator_id>")
		return
	}
	id, _ := strconv.Atoi(args[0]) // Convert the first argument to an integer
	fmt.Println("ARG:", id)
	validators, err := loadValidators("validators.csv") // Load validators from the CSV file
	if err != nil {
		fmt.Println("Error loading validators:", err)
	}
	// to receive the incomming message from the other parties.
	receiveChan := make(chan []byte, 10000) // Create a buffered channel to receive messages

	// Create a list of party IDs
	var parties []uint16
	for i, _ := range validators[1:] {
		parties = append(parties, uint16(i+1)) // Append party IDs to the list
	}

	// Create a new transport for message exchange
	transport := exchange.NewTransport(id, parties) // Create a new transport for message exchange

	// Create a new local mpc party. with correct partyID
	mpcParty := mpc.NewParty(uint16(id), logger(validators[id].ID, "main")) // Create a new MPC party

	// Initialize the MPC party, with the transport send message function
	mpcParty.Init(parties, threshold, transport.SendMsg) // Initialize the MPC party

	println("DKG")

	wg.Add(1)
	t1 := time.Now()
	defer func() {
		println("DKG completed in", time.Since(t1))
	}()

	// Start watching the file for incoming messages, as soon as we get a message in the file ,we pass it to receiveChan
	go transport.WatchFile(1*time.Millisecond, receiveChan) // Start watching the file for incoming messages

	// Start the DKG process in a goroutine
	go func() {
		defer wg.Done()

		// here sign the message.
		keyShare, err := mpcParty.KeyGen(context.Background()) // Perform distributed key generation
		if err != nil {
			fmt.Println("Error performing DKG:", err)
		}
		println("KeyShare:", len(keyShare))
	}()

	// When we get message into the receive Chan, we pass it to the mpcParty using OnMsg function of the party
	for msg := range receiveChan { // Receive messages from the channel
		// Unmarshal the JSON message, json messsage structre is defined in exchange package
		var msgStructured exchange.Msg
		json.Unmarshal(msg, &msgStructured)
		//Pass the recieved message into the mpc party
		mpcParty.OnMsg(msgStructured.Message, uint16(msgStructured.From), msgStructured.Broadcast) // Handle the message
	}
	wg.Wait() // Wait for all goroutines to finish
	println("DKG completed in", time.Since(t1))

	// We saved the key share in the file, we can read it from the file,
	// Now we read the corresponding to party iD key share from the file,

	// Now we use this key share and load into the party. Using SetShareData() function.
	// Now we define a new message to sign. and pass it to mpcparty.sign() function. and get the final signature..
	// sructure of the singature is.
	// After DKG completes, load the key share and set it in the party
	// keyShare, err := mpcParty.GetShareData()
	// if err != nil {
	// 	fmt.Println("Failed to retrieve key share:", err)
	// 	return
	// }

	// Serialize key share before setting it
	// keyShareBytes, err := json.Marshal(keyShare)
	// if err != nil {
	// 	fmt.Println("Failed to serialize key share:", err)
	// 	return
	// }

	msgToSign := []byte(createSolanaTransaction()) // Create a new message to sign
	digest := Digest(msgToSign)

	fmt.Println("Starting the signing process...") // Debug statement before signing
	sign, err := mpcParty.Sign(context.Background(), digest)
	if err != nil {
		fmt.Println("Failed to sign message:", err)
	} else {
		fmt.Println("Signature:", sign)
	}
	fmt.Println("Signing process completed.") // Debug statement after signing

	sigSet := make(map[string]struct{})
	for _, s := range sign {
		sigSet[string(s)] = struct{}{}
	}
	pk, err := mpcParty.ThresholdPK()
	if err != nil {
		fmt.Println("Failed to get threshold public key:", err)
	}
	println("Threshold PK:", pk)

	ed25519.Verify(pk, Digest(msgToSign), sign)

	// type Signature struct {
	// 	R *big.Int
	// 	S *big.Int
	// }

}

func logger(id string, testName string) mpc.Logger {
	logConfig := zap.NewDevelopmentConfig()
	logger, _ := logConfig.Build()
	logger = logger.With(zap.String("t", testName)).With(zap.String("id", id))
	return logger.Sugar()
}
