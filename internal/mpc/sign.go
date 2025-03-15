package ecdsa

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"math/big"
	"sync"

	"github.com/bnb-chain/tss-lib/v2/common"
	"github.com/bnb-chain/tss-lib/v2/eddsa/signing"
	"github.com/decred/dcrd/dcrec/edwards/v2"
)

// Sign generates a signature for the given message hash using threshold ECDSA.
func (p *Party) Sign(ctx context.Context, msgHash []byte) ([]byte, error) {
	if p.shareData == nil {
		return nil, fmt.Errorf("must call SetShareData() before attempting to sign")
	}

	log.Println("[INFO] Starting signing process")
	defer log.Println("[INFO] Signing process completed")
	defer close(p.closeChan)

	end := make(chan *common.SignatureData, 1)
	msgToSign := big.NewInt(0).SetBytes(msgHash)

	// Initialize local signing party
	party := signing.NewLocalParty(msgToSign, p.params, *p.shareData, p.out, end)

	var endWG sync.WaitGroup
	endWG.Add(1)

	go func() {
		defer endWG.Done()
		if err := party.Start(); err != nil {
			log.Printf("[ERROR] Failed signing: %v\n", err)
		}
	}()

	defer endWG.Wait()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("signing timed out: %w", ctx.Err())

		case sigOut := <-end:
			// Validate the signed message
			if !bytes.Equal(sigOut.M, msgToSign.Bytes()) {
				return nil, fmt.Errorf("message mismatch: expected %s, got %s",
					base64.StdEncoding.EncodeToString(msgHash),
					base64.StdEncoding.EncodeToString(sigOut.M))
			}

			// Extract R and S components of the signature
			var sig struct {
				R, S *big.Int
			}
			sig.R = new(big.Int).SetBytes(sigOut.R)
			sig.S = new(big.Int).SetBytes(sigOut.S)

			s := &edwards.Signature{R: sig.R, S: sig.S}
			log.Printf("[INFO] Signature generated:\n  R: %s\n  S: %s\n", sig.R.Text(16), sig.S.Text(16))
			return s.Serialize(), nil

		case msg := <-p.in:
			// Process incoming messages
			raw, routing, err := msg.WireBytes()
			if err != nil {
				log.Printf("[WARNING] Error serializing message: %v\n", err)
				continue
			}

			log.Printf("[INFO] Received message from %s\n", routing.From.Id)
			if ok, err := party.UpdateFromBytes(raw, routing.From, routing.IsBroadcast); !ok {
				log.Printf("[WARNING] Error updating party state: %v\n", err)
				continue
			}
		}
	}
}
