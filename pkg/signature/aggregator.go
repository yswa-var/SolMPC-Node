package signature

import (
	"fmt"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/sign"
	"go.dedis.ch/kyber/v3/sign/bdn"
)

type Aggregator struct {
	suite      pairing.Suite
	t          int // Threshold
	publicKeys []kyber.Point
}

func NewAggregator(suite pairing.Suite, t int, publicKeys []kyber.Point) *Aggregator {
	return &Aggregator{
		suite:      suite,
		t:          t,
		publicKeys: publicKeys,
	}
}

func (a *Aggregator) Aggregate(partialSigs map[int][]byte, message []byte, publicKey kyber.Point) ([]byte, error) {
	// Check if we have enough partial signatures
	if len(partialSigs) < a.t {
		return nil, fmt.Errorf("insufficient partial signatures: got %d, need %d", len(partialSigs), a.t)
	}

	// Create a mask for the participants who signed
	mask, err := sign.NewMask(a.suite, a.publicKeys, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create mask: %v", err)
	}
	for id := range partialSigs {
		// IDs are 1-based, indices are 0-based
		mask.SetBit(id-1, true)
	}

	// Aggregate the partial signatures
	var sigs [][]byte
	for _, sig := range partialSigs {
		sigs = append(sigs, sig)
	}

	aggregatedSigPoint, err := bdn.AggregateSignatures(a.suite, sigs, mask)
	if err != nil {
		return nil, fmt.Errorf("aggregation failed: %v", err)
	}

	// Marshal the aggregated signature to bytes
	aggregatedSig, err := aggregatedSigPoint.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal aggregated signature: %v", err)
	}

	// Verify the aggregated signature
	err = bdn.Verify(a.suite, publicKey, message, aggregatedSig)
	if err != nil {
		return nil, fmt.Errorf("aggregated signature verification failed: %v", err)
	}

	return aggregatedSig, nil
}
