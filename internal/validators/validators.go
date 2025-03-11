package eddsa

import (
	"fmt"
	"math/big"
	transport "tilt-valid/internal/exchange"

	"github.com/bnb-chain/tss-lib/v2/eddsa/keygen"
	"github.com/bnb-chain/tss-lib/v2/tss"
)

// Define types and methods for the Party and related structures.
type Sender func(msg []byte, isBroadcast bool, to uint16)

// Logger interface for logging messages.
type Logger interface {
	Debugf(format string, a ...interface{})
	Warnf(format string, a ...interface{})
	Errorf(format string, a ...interface{})
	Infof(format string, a ...interface{})
}

// Method to create a new Party.
func NewParty(id uint16, logger Logger) *Party {
	return &Party{
		Logger: logger,
		Id:     tss.NewPartyID(fmt.Sprintf("%d", id), "", big.NewInt(int64(id))),
		out:    make(chan tss.Message, 1000),
		in:     make(chan tss.Message, 1000),
	}
}

// Party structure representing a participant in the TSS protocol.
type Party struct {
	Transport *transport.Transport
	Logger    Logger
	sendMsg   Sender
	Id        *tss.PartyID
	params    *tss.Parameters
	out       chan tss.Message
	in        chan tss.Message
	shareData *keygen.LocalPartySaveData
	closeChan chan struct{}
}
