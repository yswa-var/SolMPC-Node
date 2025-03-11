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

const threshold = 2

// parties represents a collection of parties in the distributed key generation protocol.

// init initializes the parties with the given senders.
func (parties parties) init(senders []Sender) {
	// tsslib.SetCurve(tss.Edwards())
	for i, p := range parties {
		p.Init(parties.numericIDs(), threshold, senders[i])
	}
}

// setShareData sets the share data for each party.
func (parties parties) setShareData(shareData [][]byte) {
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

// TestTSS is a test function for the TSS protocol.
func TestTSS(t *testing.T) {
	pA := NewParty(1, logger("pA", t.Name()))
	pB := NewParty(2, logger("pB", t.Name()))
	pC := NewParty(3, logger("pC", t.Name()))

	t.Logf("Created parties")

	parties := parties{pA, pB, pC}
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
	sigs, err := parties.sign(Digest(msgToSign))
	assert.NoError(t, err)
	t.Logf("Signing completed in %v", time.Since(t1))

	sigSet := make(map[string]struct{})
	for _, s := range sigs {
		sigSet[string(s)] = struct{}{}
	}
	assert.Len(t, sigSet, 1)

	pk, err := parties[0].ThresholdPK()
	assert.NoError(t, err)

	assert.True(t, ed25519.Verify(pk, Digest(msgToSign), sigs[0]))
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
