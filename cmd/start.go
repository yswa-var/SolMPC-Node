package cmd

import (
	"fmt"
	"log"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/spf13/cobra"

	"github.com/replit/tilt-validator/pkg/aggregator"
	"github.com/replit/tilt-validator/pkg/validator"
)

var (
	validatorCount int
	threshold      int
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Tilt Validator simulation",
	RunE:  runStart,
}

func init() {
	startCmd.Flags().IntVarP(&validatorCount, "validators", "v", 10, "Number of validators")
	startCmd.Flags().IntVarP(&threshold, "threshold", "t", 7, "Signature threshold")
	rootCmd.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) error {
	// Connect to NATS
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}
	defer nc.Close()

	// Start aggregator
	agg, err := aggregator.NewAggregator(threshold, nc)
	if err != nil {
		return fmt.Errorf("failed to create aggregator: %w", err)
	}

	go func() {
		if err := agg.Run(); err != nil {
			log.Printf("Aggregator error: %v", err)
		}
	}()

	// Start validators
	var wg sync.WaitGroup
	for i := 0; i < validatorCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			v, err := validator.NewValidator(fmt.Sprintf("validator-%d", id), nc)
			if err != nil {
				log.Printf("Failed to create validator %d: %v", id, err)
				return
			}
			if err := v.Run(); err != nil {
				log.Printf("Validator %d error: %v", id, err)
			}
		}(i)
	}

	wg.Wait()
	return nil
}