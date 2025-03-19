package utils

import (
	"fmt"

	"github.com/gagliardetto/solana-go"
)

// func GenerateTransactionHash() string {
// 	return "ImBatman"
// }

func GenerateTransactionHash(t Tilt, totalAmount uint64) (map[solana.PublicKey]uint64, error) {
	distributions := make(map[solana.PublicKey]uint64)
	if len(t.SubTilts) == 0 {
		if len(t.Recipients) == 0 {
			return nil, fmt.Errorf("no recipients in leaf tilt %s", t.Name)
		}
		for _, recipient := range t.Recipients {
			distributions[recipient] = totalAmount / uint64(len(t.Recipients))
		}
		return distributions, nil
	}

	weightSum := uint8(0)
	for _, w := range t.Weights {
		weightSum += w
	}
	if weightSum != 100 {
		return nil, fmt.Errorf("weights must sum to 100 in %s, got %d", t.Name, weightSum)
	}

	subTiltWeight := t.Weights[0]
	subTiltAmount := totalAmount * uint64(subTiltWeight) / 100
	localAmount := totalAmount - subTiltAmount

	subTiltCount := len(t.SubTilts)
	if subTiltCount > 0 {
		perSubTiltAmount := subTiltAmount / uint64(subTiltCount)
		for _, subTilt := range t.SubTilts {
			subDist, err := GenerateTransactionHash(subTilt, perSubTiltAmount)
			if err != nil {
				return nil, err
			}
			for k, v := range subDist {
				distributions[k] = v
			}
		}
	}

	if len(t.Recipients) > 0 {
		for _, recipient := range t.Recipients {
			distributions[recipient] = localAmount / uint64(len(t.Recipients))
		}
	}

	return distributions, nil
}
