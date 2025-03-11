package ecdsa

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/bnb-chain/tss-lib/v2/eddsa/keygen"
)

// Method to generate a key.
func (p *Party) KeyGen(ctx context.Context) ([]byte, error) {
	log.Println("EDDSA Keygen Started")
	p.Logger.Debugf("Starting DKG")
	defer p.Logger.Debugf("Finished DKG")
	defer close(p.closeChan)

	end := make(chan *keygen.LocalPartySaveData, 1)
	party := keygen.NewLocalParty(p.params, p.out, end)

	var endWG sync.WaitGroup
	endWG.Add(1)

	go func() {
		defer endWG.Done()
		err := party.Start()
		if err != nil {
			p.Logger.Errorf("Failed generating key: %v", err)
		}
	}()

	defer endWG.Wait()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("DKG timed out: %w", ctx.Err())
		case dkgOut := <-end:
			dkgRawOut, err := json.Marshal(dkgOut)
			p.Logger.Infof("Key-Gen Completed, Key-Share:", dkgOut.Xi.String())

			if err != nil {
				return nil, fmt.Errorf("failed serializing DKG output: %w", err)
			}
			p.SaveLocalPartySaveData(dkgRawOut)
			return dkgRawOut, nil
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
