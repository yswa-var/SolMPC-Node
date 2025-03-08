package ecdsa

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"math/big"
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

// TestTSS is a test function for the TSS protocol.
func TestTSS(t *testing.T) {
	// Create three validators with unique IDs and loggers.
	// in this case we are taking 3 parties, but we can take any number of parties.
	pA := NewParty(1, logger("pA", t.Name()))
	pB := NewParty(2, logger("pB", t.Name()))
	pC := NewParty(3, logger("pC", t.Name()))
	pA.Transport.ReadMsg()

	// Initialize the parties and run the distributed key generation (DKG).
	parties1 := parties{pA, pB, pC}
	parties1.init(senders(parties1))

	t.Logf("######### DKG STARTING #################")

	t1 := time.Now()

	shares, err := parties.keygen(parties1)
	assert.NoError(t, err)
	t.Logf("DKG elapsed %s", time.Since(t1))

	// Reinitialize two parties with the generated shares and perform signing.
	parties2 := parties{pA}
	parties2.init(senders(parties2))

	parties2.setShareData(shares)

	t.Logf("Signing")

	msgToSign := []byte("bla bla")

	t.Logf("Signing message")
	t1 = time.Now()
	sigs, err := parties2.sign(digest(msgToSign))
	assert.NoError(t, err)
	t.Logf("Signing completed in %v", time.Since(t1))

	// Verify that the signatures are consistent and valid.
	sigSet := make(map[string]struct{})
	for _, s := range sigs {
		sigSet[string(s)] = struct{}{}
	}
	assert.Len(t, sigSet, 1)

	pk, err := parties2[0].ThresholdPK()
	assert.NoError(t, err)

	assert.True(t, ed25519.Verify(pk, digest(msgToSign), sigs[0]))
}

// senders returns a slice of sender functions for each party.
// Each sender function is responsible for sending messages from one party to another.
// If the broadcast flag is true, the message is sent to all parties except the sender.
// If the broadcast flag is false, the message is sent only to the specified party.
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
