package ui

import (
	"fmt"
	"log"
	"sync"
)

type Dashboard struct {
	threshold   int
	mu         sync.Mutex
	signatures []string
	ttyMode    bool
}

func NewDashboard(threshold int) (*Dashboard, error) {
	return &Dashboard{
		threshold:  threshold,
		signatures: make([]string, 0),
		ttyMode:    false, // Default to non-TTY mode
	}, nil
}

func (d *Dashboard) Run() {
	// In non-TTY mode, just log the initial status
	log.Printf("🚀 Tilt Validator Simulation Started")
	log.Printf("⚡ Services Status:")
	log.Printf("   HTTP Server: Running on :5000")
	log.Printf("   NATS Server: Running on :4222")
	log.Printf("📊 Waiting for validator signatures...")
}

func (d *Dashboard) UpdateProgress(count int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	percent := (count * 100) / d.threshold
	if percent > 100 {
		percent = 100
	}

	log.Printf("✓ Progress: %d%% (%d/%d signatures)", percent, count, d.threshold)
	if count >= d.threshold {
		log.Printf("🎉 Threshold Met! Distribution finalized!")
	}
}

func (d *Dashboard) AddSignature(validatorID string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	signature := fmt.Sprintf("➜ Signature received from %s", validatorID)
	d.signatures = append(d.signatures, signature)
	log.Printf("📝 %s", signature)
}

func (d *Dashboard) Close() {
	// Nothing to do in non-TTY mode
}