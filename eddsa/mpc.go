package ecdsa

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"os"
	"strconv"
	"sync"

	transport "teddsa/exchange"

	"github.com/bnb-chain/tss-lib/v2/common"
	"github.com/bnb-chain/tss-lib/v2/eddsa/keygen"
	"github.com/bnb-chain/tss-lib/v2/eddsa/signing"
	"github.com/bnb-chain/tss-lib/v2/tss"
	"github.com/decred/dcrd/dcrec/edwards/v2"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

func init() {
	tss.SetCurve(tss.Edwards())
}

var (
	msgURL2Round = map[string]uint8{
		// DKG
		"type.googleapis.com/binance.tsslib.eddsa.keygen.KGRound1Message":  1,
		"type.googleapis.com/binance.tsslib.eddsa.keygen.KGRound2Message1": 2,
		"type.googleapis.com/binance.tsslib.eddsa.keygen.KGRound2Message2": 3,

		// Signing
		"type.googleapis.com/binance.tsslib.eddsa.signing.SignRound1Message": 5,
		"type.googleapis.com/binance.tsslib.eddsa.signing.SignRound2Message": 6,
		"type.googleapis.com/binance.tsslib.eddsa.signing.SignRound3Message": 7,
	}

	broadcastMessages = map[string]struct{}{
		// DKG
		"type.googleapis.com/binance.tsslib.eddsa.keygen.KGRound1Message":  {},
		"type.googleapis.com/binance.tsslib.eddsa.keygen.KGRound2Message2": {},

		// Signing
		"type.googleapis.com/binance.tsslib.eddsa.signing.SignRound1Message": {},
		"type.googleapis.com/binance.tsslib.eddsa.signing.SignRound2Message": {},
		"type.googleapis.com/binance.tsslib.eddsa.signing.SignRound3Message": {},
	}
)

type Sender func(msg []byte, isBroadcast bool, to uint16)

type parties []*Party

func (parties parties) numericIDs() []uint16 {
	var res []uint16
	for _, p := range parties {
		res = append(res, uint16(big.NewInt(0).SetBytes(p.Id.Key).Uint64()))
	}

	return res
}

type Logger interface {
	Debugf(format string, a ...interface{})
	Warnf(format string, a ...interface{})
	Errorf(format string, a ...interface{})
	Infof(format string, a ...interface{})
}

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

func NewParty(id uint16, logger Logger) *Party {
	return &Party{
		Logger: logger,
		Id:     tss.NewPartyID(fmt.Sprintf("%d", id), "", big.NewInt(int64(id))),
		out:    make(chan tss.Message, 1000),
		in:     make(chan tss.Message, 1000),
	}
}

func (p *Party) ID() *tss.PartyID {
	return p.Id
}

func (p *Party) locatePartyIndex(id *tss.PartyID) int {
	for index, p := range p.params.Parties().IDs() {
		if bytes.Equal(p.Key, id.Key) {
			return index
		}
	}

	return -1
}

func (p *Party) ClassifyMsg(msgBytes []byte) (uint8, bool, error) {
	msg := &any.Any{}
	if err := proto.Unmarshal(msgBytes, msg); err != nil {
		p.Logger.Warnf("Received invalid message: %v", err)
		return 0, false, err
	}

	_, isBroadcast := broadcastMessages[msg.TypeUrl]

	round := msgURL2Round[msg.TypeUrl]
	if round > 4 {
		round = round - 4
	}
	return round, isBroadcast, nil
}

func (p *Party) OnMsg(msgBytes []byte, from uint16, broadcast bool) {
	id := tss.NewPartyID(fmt.Sprintf("%d", from), "", big.NewInt(int64(from)))
	id.Index = p.locatePartyIndex(id)
	msg, err := tss.ParseWireMessage(msgBytes, id, broadcast)
	if err != nil {
		p.Logger.Warnf("Received invalid message (%s) of %d bytes from %d: %v", base64.StdEncoding.EncodeToString(msgBytes), len(msgBytes), from, err)
		return
	}

	key := msg.GetFrom().KeyInt()
	if key == nil || key.Cmp(big.NewInt(int64(math.MaxUint16))) >= 0 {
		p.Logger.Warnf("Message received from invalid key: %v", key)
		return
	}

	claimedFrom := uint16(key.Uint64())
	if claimedFrom != from {
		p.Logger.Warnf("Message claimed to be from %d but was received from %d", claimedFrom, from)
		return
	}

	p.in <- msg
}

func (p *Party) ThresholdPK() ([]byte, error) {
	if p.shareData == nil {
		return nil, fmt.Errorf("must call SetShareData() before attempting to sign")
	}

	pk := p.shareData.EDDSAPub

	edPK := &edwards.PublicKey{
		Curve: tss.Edwards(),
		X:     pk.X(),
		Y:     pk.Y(),
	}

	pkBytes := copyBytes(edPK.Serialize())
	return pkBytes[:], nil
}

func (p *Party) CheckShareData() bool {
	return p.shareData == nil
}

func (p *Party) SetShareData(shareData []byte) error {
	var localSaveData keygen.LocalPartySaveData
	err := json.Unmarshal(shareData, &localSaveData)
	if err != nil {
		return fmt.Errorf("failed deserializing shares: %w", err)
	}
	localSaveData.EDDSAPub.SetCurve(tss.Edwards())
	for _, xj := range localSaveData.BigXj {
		xj.SetCurve(tss.Edwards())
	}
	p.shareData = &localSaveData
	return nil
}

func (p *Party) Init(parties []uint16, threshold int, sendMsg func(msg []byte, isBroadcast bool, to uint16)) {
	partyIDs := partyIDsFromNumbers(parties)
	ctx := tss.NewPeerContext(partyIDs)
	p.params = tss.NewParameters(tss.Edwards(), ctx, p.Id, len(parties), threshold)
	p.Id.Index = p.locatePartyIndex(p.Id)
	p.sendMsg = sendMsg
	p.closeChan = make(chan struct{})
	go p.sendMessages()
}

func partyIDsFromNumbers(parties []uint16) []*tss.PartyID {
	var partyIDs []*tss.PartyID
	for _, p := range parties {
		pID := tss.NewPartyID(fmt.Sprintf("%d", p), "", big.NewInt(int64(p)))
		partyIDs = append(partyIDs, pID)
	}
	return tss.SortPartyIDs(partyIDs)
}

func (p *Party) Sign(ctx context.Context, msgHash []byte) ([]byte, error) {
	if p.shareData == nil {
		return nil, fmt.Errorf("must call SetShareData() before attempting to sign")
	}
	p.Logger.Debugf("Starting signing")
	defer p.Logger.Debugf("Finished signing")

	defer close(p.closeChan)

	end := make(chan *common.SignatureData, 1)

	msgToSign := big.NewInt(0).SetBytes(msgHash)
	party := signing.NewLocalParty(msgToSign, p.params, *p.shareData, p.out, end)

	var endWG sync.WaitGroup
	endWG.Add(1)

	go func() {
		defer endWG.Done()
		err := party.Start()
		if err != nil {
			p.Logger.Errorf("Failed signing: %v", err)
		}
	}()

	defer endWG.Wait()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("signing timed out: %w", ctx.Err())
		case sigOut := <-end:
			if !bytes.Equal(sigOut.M, msgToSign.Bytes()) {
				return nil, fmt.Errorf("message we requested to sign is %s but actual message signed is %s",
					base64.StdEncoding.EncodeToString(msgHash),
					base64.StdEncoding.EncodeToString(sigOut.M))
			}
			var sig struct {
				R, S *big.Int
			}
			sig.R = big.NewInt(0)
			sig.S = big.NewInt(0)
			sig.R.SetBytes(sigOut.R)
			sig.S.SetBytes(sigOut.S)

			s := &edwards.Signature{R: sig.R, S: sig.S}
			log.Println("Signature Generated\n\n", "R:", sig.R.Text(16), "\n S:", sig.S.Text(16), "\n\n ")
			return s.Serialize(), nil
		case msg := <-p.in:
			raw, routing, err := msg.WireBytes()
			if err != nil {
				p.Logger.Warnf("Received error when serializing message: %v", err)
				continue
			}
			p.Logger.Debugf("%s Got message from %s", p.Id.Id, routing.From.Id)
			ok, err := party.UpdateFromBytes(raw, routing.From, routing.IsBroadcast)
			if !ok {
				p.Logger.Warnf("Received error when updating party: %v", err.Error())
				continue
			}
		}
	}
}

func (p *Party) KeyGen(ctx context.Context) ([]byte, error) {
	log.Println("EDDSA Keygen Started")
	p.Logger.Debugf("Starting DKG")
	defer p.Logger.Debugf("Finished DKG")

	defer close(p.closeChan)

	end := make(chan *keygen.LocalPartySaveData, 1)
	party := keygen.NewLocalParty(p.params, p.out, end)

	var endWG sync.WaitGroup
	endWG.Add(1)

	go func() {
		defer endWG.Done()
		err := party.Start()
		if err != nil {
			p.Logger.Errorf("Failed generating key: %v", err)
		}
	}()

	defer endWG.Wait()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("DKG timed out: %w", ctx.Err())
		case dkgOut := <-end:
			dkgRawOut, err := json.Marshal(dkgOut)
			p.Logger.Infof("Key-Gen Completed, Key-Share:", dkgOut.Xi.String())

			if err != nil {
				return nil, fmt.Errorf("failed serializing DKG output: %w", err)
			}
			p.SaveLocalPartySaveData(dkgRawOut)
			return dkgRawOut, nil
		case msg := <-p.in:
			raw, routing, err := msg.WireBytes()
			if err != nil {
				p.Logger.Warnf("Received error when serializing message: %v", err)
				continue
			}
			p.Logger.Debugf("%s Got message from %s", p.Id.Id, routing.From.Id)
			ok, err := party.UpdateFromBytes(raw, routing.From, routing.IsBroadcast)
			if !ok {
				p.Logger.Warnf("Received error when updating party: %v", err.Error())
				continue
			}
		}
	}
}

func (p *Party) sendMessages() {
	for {
		select {
		case <-p.closeChan:
			return
		case msg := <-p.out:
			msgBytes, routing, err := msg.WireBytes()
			if err != nil {
				p.Logger.Warnf("Failed marshaling message: %v", err)
				continue
			}
			if routing.IsBroadcast {
				p.sendMsg(msgBytes, routing.IsBroadcast, 0)
			} else {
				for _, to := range msg.GetTo() {
					p.sendMsg(msgBytes, routing.IsBroadcast, uint16(big.NewInt(0).SetBytes(to.Key).Uint64()))
				}
			}
		}
	}
}

func (p *Party) SaveLocalPartySaveData(shareData []byte) {
	//save bytes data locally
	WriteToFile("localsavedata_eddsa"+strconv.Itoa(p.ID().Index), shareData)
	p.Logger.Debugf("Saved Data locally")

}

func (p *Party) LoadLocalPartySaveData() {
	shareData, err := ReadFromFile("localsavedata_eddsa" + strconv.Itoa(p.ID().Index))
	if err != nil {
		p.Logger.Debugf("Failed to load data", err)
		return
	} else {
		p.SetShareData(shareData)
	}
}

func digest(in []byte) []byte {
	h := sha256.New()
	h.Write(in)
	return h.Sum(nil)
}

// copyBytes copies a byte slice to a 32 byte array.
func copyBytes(aB []byte) *[32]byte {
	if aB == nil {
		return nil
	}
	s := new([32]byte)

	// If we have a short byte string, expand
	// it so that it's long enough.
	aBLen := len(aB)
	if aBLen < 32 {
		diff := 32 - aBLen
		for i := 0; i < diff; i++ {
			aB = append([]byte{0x00}, aB...)
		}
	}

	for i := 0; i < 32; i++ {
		s[i] = aB[i]
	}

	return s
}

func WriteToFile(filename string, data []byte) error {
	// Write data to the specified filename with permissions
	err := os.WriteFile(filename, data, 0644) // 0644 allows owner to read/write and others to read
	if err != nil {
		return err
	}
	return nil
}

func ReadFromFile(filename string) ([]byte, error) {
	// Read the data from the file
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return data, nil
}
