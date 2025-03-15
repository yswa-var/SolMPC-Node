package ecdsa

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/bnb-chain/tss-lib/v2/eddsa/keygen"
)

// KeyGen performs Distributed Key Generation (DKG) for EDDSA.
// This method initiates a key generation process, handles incoming messages,
// and ensures the generated key share is properly saved.
func (p *Party) KeyGen(ctx context.Context) ([]byte, error) {
	log.Println("[INFO] EDDSA Key Generation started.")
	defer log.Println("[INFO] EDDSA Key Generation completed.")
	defer close(p.closeChan)

	// Channel to receive the final DKG output
	end := make(chan *keygen.LocalPartySaveData, 1)

	// Initialize the local party for key generation
	party := keygen.NewLocalParty(p.params, p.out, end)

	var endWG sync.WaitGroup
	endWG.Add(1)

	go func() {
		defer endWG.Done()
		if err := party.Start(); err != nil {
			log.Printf("[ERROR] Failed to generate key: %v\n", err)
		}
	}()

	defer endWG.Wait() // Ensure key generation completes before returning

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("[ERROR] DKG timed out: %w", ctx.Err())

		// Handle successful key generation
		case dkgOut := <-end:
			dkgRawOut, err := json.Marshal(dkgOut)
			if err != nil {
				return nil, fmt.Errorf("[ERROR] Failed to serialize DKG output: %w", err)
			}

			log.Printf("[INFO] Key-Generation completed. Key-Share: %s\n", dkgOut.Xi.String())
			p.SaveLocalPartySaveData(dkgRawOut)
			return dkgRawOut, nil

		// Process incoming messages for the keygen process
		case msg := <-p.in:
			raw, routing, err := msg.WireBytes()
			if err != nil {
				log.Printf("[WARNING] Error serializing incoming message: %v\n", err)
				continue
			}

			log.Printf("[INFO] Received message from Party %s\n", routing.From.Id)

			ok, err := party.UpdateFromBytes(raw, routing.From, routing.IsBroadcast)
			if !ok {
				log.Printf(".", err)
				continue
			}
		}
	}
}
