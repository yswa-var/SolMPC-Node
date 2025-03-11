package exchange

import (
	"strconv"
	"sync"
)

type Transport struct {
	Mutex   sync.Mutex
	partyID int
	parties []uint16
}

func NewTransport(partyID int, parties []uint16) *Transport {
	return &Transport{partyID: partyID, parties: parties}
}

func (t *Transport) GetFileName() string {
	return "/Users/apple/Documents/GitHub/tilt-validator-main/internal/Transport" + strconv.Itoa(t.partyID) + ".csv"
}

func (t *Transport) GetReceiverFileName(id string) string {
	return "/Users/apple/Documents/GitHub/tilt-validator-main/internal/Transport" + id + ".csv"
}

func (t *Transport) getParties() []uint16 {
	return t.parties
}
