package aggregator

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/replit/tilt-validator/pkg/types"
	"github.com/replit/tilt-validator/pkg/ui"
)

type Aggregator struct {
	signatures []types.SignedDistribution
	threshold  int
	nc         *nats.Conn
	mu         sync.Mutex
	done       chan struct{}
	closeOnce  sync.Once
	dashboard  *ui.Dashboard
}

func NewAggregator(threshold int, nc *nats.Conn) (*Aggregator, error) {
	dashboard, err := ui.NewDashboard(threshold)
	if err != nil {
		return nil, fmt.Errorf("failed to create dashboard: %w", err)
	}

	return &Aggregator{
		signatures: make([]types.SignedDistribution, 0),
		threshold:  threshold,
		nc:         nc,
		done:       make(chan struct{}),
		dashboard:  dashboard,
	}, nil
}

func (a *Aggregator) Run() error {
	// Start the dashboard
	go a.dashboard.Run()
	defer a.dashboard.Close()

	// Subscribe to signatures
	sub, err := a.nc.Subscribe("tilt.signatures", func(msg *nats.Msg) {
		var sig types.SignedDistribution
		if err := json.Unmarshal(msg.Data, &sig); err != nil {
			log.Printf("Error unmarshaling signature: %v", err)
			return
		}

		a.mu.Lock()
		if len(a.signatures) >= a.threshold {
			a.mu.Unlock()
			return // Skip if threshold already met
		}
		a.signatures = append(a.signatures, sig)
		count := len(a.signatures)
		reachedThreshold := count >= a.threshold
		a.mu.Unlock()

		// Update dashboard
		a.dashboard.UpdateProgress(count)
		a.dashboard.AddSignature(sig.ValidatorID)

		if reachedThreshold {
			if a.verify() {
				a.closeOnce.Do(func() {
					close(a.done)
				})
			}
		}
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}
	defer sub.Unsubscribe()

	// Wait for completion
	<-a.done
	return nil
}

func (a *Aggregator) verify() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.signatures) >= a.threshold
}