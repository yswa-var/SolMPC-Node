package ecdsa

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bnb-chain/tss-lib/v2/tss"
	"github.com/stretchr/testify/assert"

	// tsslib "crypto/tss-lib"
	"go.uber.org/zap"
)

const threshold = 1

// init initializes the parties with the given senders.
func (parties parties) init(senders []Sender) {
	// Initialize each party with the numeric IDs of all parties, the threshold, and the corresponding sender function.
	for i, p := range parties {
		p.Init(parties.numericIDs(), threshold, senders[i])
	}
}

// setShareData sets the share data for each party.
func (parties parties) setShareData(shareData [][]byte) {
	// Set the share data for each party.
	for i, p := range parties {
		p.SetShareData(shareData[i])
	}
}

// sign signs the given message using each party's private key.
func (parties parties) sign(msg []byte) ([][]byte, error) {
	var lock sync.Mutex
	var sigs [][]byte
	var threadSafeError atomic.Value

	var wg sync.WaitGroup
	wg.Add(len(parties))

	for _, p := range parties {
		go func(p *Party) {
			defer wg.Done()
			// Each party signs the message.
			sig, err := p.Sign(context.Background(), msg)
			if err != nil {
				threadSafeError.Store(err.Error())
				return
			}

			lock.Lock()
			sigs = append(sigs, sig)
			lock.Unlock()
		}(p)
	}

	wg.Wait()

	err := threadSafeError.Load()
	if err != nil {
		return nil, fmt.Errorf(err.(string))
	}

	return sigs, nil
}

// keygen performs the distributed key generation protocol and returns the shares.
func (parties parties) keygen() ([][]byte, error) {
	var lock sync.Mutex
	shares := make([][]byte, len(parties))
	var threadSafeError atomic.Value

	var wg sync.WaitGroup
	wg.Add(len(parties))

	for i, p := range parties {
		go func(p *Party, i int) {
			defer wg.Done()
			// Each party performs key generation.
			share, err := p.KeyGen(context.Background())
			if err != nil {
				threadSafeError.Store(err.Error())
				return
			}

			lock.Lock()
			shares[i] = share
			lock.Unlock()
		}(p, i)
	}

	wg.Wait()

	err := threadSafeError.Load()
	if err != nil {
		return nil, fmt.Errorf(err.(string))
	}

	return shares, nil
}

// Mapping returns a map of party IDs to party objects.
func (parties parties) Mapping() map[string]*tss.PartyID {
	partyIDMap := make(map[string]*tss.PartyID)
	for _, id := range parties {
		partyIDMap[id.Id.Id] = id.Id
	}
	return partyIDMap
}

//sendMSG( msgBytes []byte, broadcast bool, to uint16)

// generateRandomHash generates a random SHA-256 hash string.
func solanaTransaction() (string, error) {
	// demo function whcih mimics the solana transaction for vaidation.
	// Create a new SHA-256 hash
	hash := sha256.New()

	// Generate a random number.
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %v", err)
	}

	// Write the random bytes to the hash.
	_, err = hash.Write(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to write to hash: %v", err)
	}

	// Compute the hash and return it as a hex string.
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// Validator represents a validator with an ID and name.
type Validator struct {
	ID   string
	Name string
}

// loadValidators loads the validators from a CSV file.
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

// TestTSS is a test function for the simulation of how validators will work parties == validators.
func TestTSS(t *testing.T) {
	// when a validator will start it will save id and name in the csv file.

	// Load the validators from the CSV file.
	validators, err := loadValidators("validators.csv")
	if err != nil {
		t.Fatalf("failed to load validators: %v", err)
	}

	// creating validators from validators.csv
	var parties parties
	for i, v := range validators {
		p := NewParty(uint16(i+1), logger(v.ID, t.Name()))
		parties = append(parties, p)
	}
	parties.init(senders(parties))

	t.Logf("Running DKG")

	t1 := time.Now()
	shares, err := parties.keygen()
	assert.NoError(t, err)
	t.Logf("DKG elapsed %s", time.Since(t1))

	parties.init(senders(parties))

	parties.setShareData(shares)

	t.Logf("Signing")

	msgToSign := []byte("bla bla")

	t.Logf("Signing message")
	t1 = time.Now()
	sigs, err := parties.sign(digest(msgToSign))
	assert.NoError(t, err)
	t.Logf("Signing completed in %v", time.Since(t1))

	sigSet := make(map[string]struct{})
	for _, s := range sigs {
		sigSet[string(s)] = struct{}{}
	}
	assert.Len(t, sigSet, 1)

	pk, err := parties[0].ThresholdPK()
	assert.NoError(t, err)

	assert.True(t, ed25519.Verify(pk, digest(msgToSign), sigs[0]))
}

// senders returns a slice of sender functions for each party.
func senders(parties parties) []Sender {
	var senders []Sender
	for _, src := range parties {
		src := src
		sender := func(msgBytes []byte, broadcast bool, to uint16) {
			messageSource := uint16(big.NewInt(0).SetBytes(src.Id.Key).Uint64())
			if broadcast {
				for _, dst := range parties {
					if dst.Id == src.Id {
						continue
					}
					dst.OnMsg(msgBytes, messageSource, broadcast)
				}
			} else {
				for _, dst := range parties {
					if to != uint16(big.NewInt(0).SetBytes(dst.Id.Key).Uint64()) {
						continue
					}
					dst.OnMsg(msgBytes, messageSource, broadcast)
				}
			}
		}
		senders = append(senders, sender)
	}
	return senders
}

// logger creates a logger with the given ID and test name.
func logger(id string, testName string) Logger {
	logConfig := zap.NewDevelopmentConfig()
	logger, _ := logConfig.Build()
	logger = logger.With(zap.String("t", testName)).With(zap.String("id", id))
	return logger.Sugar()
}
