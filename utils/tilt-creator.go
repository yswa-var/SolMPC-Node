package utils

import (
	"fmt"
	"os"
	"strconv"

	"github.com/gagliardetto/solana-go"
)

// Tilt represents a distribution tree with sub-tilts
type Tilt struct {
	Name       string
	SubTilts   []Tilt
	Weights    []uint8
	Recipients []solana.PublicKey // Token accounts
}

func UpdateTiltCounter(value int) error {
	filePath := "/Users/apple/Documents/GitHub/tilt-validator-main/utils/create-tilt-flag.txt"
	return os.WriteFile(filePath, []byte(strconv.Itoa(value)), 0644)
}

// CreateTilt creates a tilt with 1 or 2 sub-tilts
func CreateTilt(sender, recipient1Token, recipient2Token solana.PublicKey, withTwoSubTilts bool) Tilt {

	err := UpdateTiltCounter(1)
	if err != nil {
		fmt.Println("Error updating tilt counter:", err)
		return Tilt{}
	}

	rootTilt := Tilt{
		Name:       "RootTilt",
		Weights:    []uint8{75, 25}, // 75% to sub-tilt(s), 25% to local
		Recipients: []solana.PublicKey{recipient1Token},
	}

	subTilt1 := Tilt{
		Name:       "SubTilt1",
		Recipients: []solana.PublicKey{recipient2Token},
	}

	if withTwoSubTilts {
		rootTilt.Weights = []uint8{50, 50} // 50% to sub-tilts, 50% to local
		subTilt2 := Tilt{
			Name:       "SubTilt2",
			Recipients: []solana.PublicKey{sender}, // Sender as recipient
		}
		rootTilt.SubTilts = []Tilt{subTilt1, subTilt2}
	} else {
		rootTilt.SubTilts = []Tilt{subTilt1}
	}

	err = UpdateTiltCounter(2)
	if err != nil {
		fmt.Println("Error updating tilt counter:", err)
	}
	return rootTilt
}
