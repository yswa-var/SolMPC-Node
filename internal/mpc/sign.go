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

// Method to sign a message.
func (p *Party) Sign(ctx context.Context, msgHash []byte) ([]byte, error) {
	if p.shareData == nil {
		return nil, fmt.Errorf("must call SetShareData() before attempting to sign")
	}
	p.Logger.Debugf("Starting signing")
	defer p.Logger.Debugf("Finished signing")
	defer close(p.closeChan)

	end := make(chan *common.SignatureData, 1)
	msgToSign := big.NewInt(0).SetBytes(msgHash)
	party := signing.NewLocalParty(msgToSign, p.params, *p.shareData, p.out, end)

	var endWG sync.WaitGroup
	endWG.Add(1)

	go func() {
		defer endWG.Done()
		err := party.Start()
		if err != nil {
			p.Logger.Errorf("Failed signing: %v", err)
		}
	}()

	defer endWG.Wait()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("signing timed out: %w", ctx.Err())
		case sigOut := <-end:
			if !bytes.Equal(sigOut.M, msgToSign.Bytes()) {
				return nil, fmt.Errorf("message we requested to sign is %s but actual message signed is %s",
					base64.StdEncoding.EncodeToString(msgHash),
					base64.StdEncoding.EncodeToString(sigOut.M))
			}
			var sig struct {
				R, S *big.Int
			}
			sig.R = big.NewInt(0)
			sig.S = big.NewInt(0)
			sig.R.SetBytes(sigOut.R)
			sig.S.SetBytes(sigOut.S)

			s := &edwards.Signature{R: sig.R, S: sig.S}
			log.Println("Signature Generated\n\n", "R:", sig.R.Text(16), "\n S:", sig.S.Text(16), "\n\n ")
			return s.Serialize(), nil
		case msg := <-p.in:
			raw, routing, err := msg.WireBytes()
			if err != nil {
				p.Logger.Warnf("Received error when serializing message: %v", err)
				continue
			}
			p.Logger.Debugf("%s Got message from %s", p.Id.Id, routing.From.Id)
			ok, err := party.UpdateFromBytes(raw, routing.From, routing.IsBroadcast)
			if !ok {
				p.Logger.Warnf("Received error when updating party: %v", err.Error())
				continue
			}
		}
	}
}
