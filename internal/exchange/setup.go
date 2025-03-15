package exchange

import (
	"fmt"
	"strconv"
	"sync"
	"tilt-valid/cmd/config"
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
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("error in loading config")
	}
	path := cfg.TransportPath
	return path + strconv.Itoa(t.partyID) + ".csv"
}

func (t *Transport) GetReceiverFileName(id string) string {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("error in loading config")
	}
	path := cfg.TransportPath
	return path + id + ".csv"
}

func (t *Transport) getParties() []uint16 {
	return t.parties
}
