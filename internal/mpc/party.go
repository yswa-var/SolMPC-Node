package ecdsa

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"strconv"

	transport "tilt-valid/internal/exchange"

	"github.com/bnb-chain/tss-lib/v2/eddsa/keygen"
	"github.com/bnb-chain/tss-lib/v2/tss"
	"github.com/decred/dcrd/dcrec/edwards/v2"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

// Initialize the TSS library with the Edwards curve.
func init() {
	tss.SetCurve(tss.Edwards())
}

// Define maps for message types and broadcast messages.
var (
	msgURL2Round = map[string]uint8{
		// DKG (Distributed Key Generation) messages
		"type.googleapis.com/binance.tsslib.eddsa.keygen.KGRound1Message":  1,
		"type.googleapis.com/binance.tsslib.eddsa.keygen.KGRound2Message1": 2,
		"type.googleapis.com/binance.tsslib.eddsa.keygen.KGRound2Message2": 3,
		// Signing messages
		"type.googleapis.com/binance.tsslib.eddsa.signing.SignRound1Message": 5,
		"type.googleapis.com/binance.tsslib.eddsa.signing.SignRound2Message": 6,
		"type.googleapis.com/binance.tsslib.eddsa.signing.SignRound3Message": 7,
	}

	broadcastMessages = map[string]struct{}{
		// DKG messages to be broadcast
		"type.googleapis.com/binance.tsslib.eddsa.keygen.KGRound1Message":  {},
		"type.googleapis.com/binance.tsslib.eddsa.keygen.KGRound2Message2": {},
		// Signing messages to be broadcast
		"type.googleapis.com/binance.tsslib.eddsa.signing.SignRound1Message": {},
		"type.googleapis.com/binance.tsslib.eddsa.signing.SignRound2Message": {},
		"type.googleapis.com/binance.tsslib.eddsa.signing.SignRound3Message": {},
	}
)

// Define types and methods for the Party and related structures.
type Sender func(msg []byte, isBroadcast bool, to uint16)

type parties []*Party

// Method to get numeric IDs of parties.
func (parties parties) numericIDs() []uint16 {
	var res []uint16
	for _, p := range parties {
		res = append(res, uint16(big.NewInt(0).SetBytes(p.Id.Key).Uint64()))
	}
	return res
}

// Logger interface for logging messages.
type Logger interface {
	Debugf(format string, a ...interface{})
	Warnf(format string, a ...interface{})
	Errorf(format string, a ...interface{})
	Infof(format string, a ...interface{})
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

// Method to get the Party ID.
func (p *Party) ID() *tss.PartyID {
	return p.Id
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

// Method to locate the index of a party.
func (p *Party) locatePartyIndex(id *tss.PartyID) int {
	for index, p := range p.params.Parties().IDs() {
		if bytes.Equal(p.Key, id.Key) {
			return index
		}
	}
	return -1
}

// Method to classify a message.
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

// Method to handle incoming messages.
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

// Method to get the threshold public key.
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

// Method to check if share data is set.
func (p *Party) CheckShareData() bool {
	return p.shareData == nil
}

// Method to set share data.
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

// Method to initialize the party.
func (p *Party) Init(parties []uint16, threshold int, sendMsg func(msg []byte, isBroadcast bool, to uint16)) {
	partyIDs := partyIDsFromNumbers(parties)
	ctx := tss.NewPeerContext(partyIDs)
	p.params = tss.NewParameters(tss.Edwards(), ctx, p.Id, len(parties), threshold)
	p.Id.Index = p.locatePartyIndex(p.Id)
	p.sendMsg = sendMsg
	p.closeChan = make(chan struct{})
	go p.sendMessages()
}

// Helper function to create party IDs from numbers.
func partyIDsFromNumbers(parties []uint16) []*tss.PartyID {
	var partyIDs []*tss.PartyID
	for _, p := range parties {
		pID := tss.NewPartyID(fmt.Sprintf("%d", p), "", big.NewInt(int64(p)))
		partyIDs = append(partyIDs, pID)
	}
	return tss.SortPartyIDs(partyIDs)
}

// Method to send messages.
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

// Method to save local party save data to a file.
func (p *Party) SaveLocalPartySaveData(shareData []byte) {
	WriteToFile("localsavedata_eddsa"+strconv.Itoa(p.ID().Index), shareData)
	p.Logger.Debugf("Saved Data locally")
}

// Method to load local party save data from a file.
func (p *Party) LoadLocalPartySaveData() {
	shareData, err := ReadFromFile("localsavedata_eddsa" + strconv.Itoa(p.ID().Index))
	if err != nil {
		p.Logger.Debugf("Failed to load data", err)
		return
	} else {
		p.SetShareData(shareData)
	}
}

// Function to compute the SHA-256 digest of input data.
func Digest(in []byte) []byte {
	h := sha256.New()
	h.Write(in)
	return h.Sum(nil)
}

// Function to copy a byte slice to a 32-byte array.
func copyBytes(aB []byte) *[32]byte {
	if aB == nil {
		return nil
	}
	s := new([32]byte)

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

// Function to write data to a file.
func WriteToFile(filename string, data []byte) error {
	err := os.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

// Function to read data from a file.
func ReadFromFile(filename string) ([]byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Method to get share data from a file.
func (p *Party) GetShareData() (*keygen.LocalPartySaveData, error) {
	fmt.Println("------------------------->>>>>>>ID:", p.ID().Id)
	fileName := fmt.Sprintf("localsavedata_eddsa%d", p.ID().Index)
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open key share file: %w", err)
	}
	defer file.Close()

	var shareData *keygen.LocalPartySaveData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&shareData); err != nil {
		return nil, fmt.Errorf("failed to parse key share data: %w", err)
	}

	return shareData, nil
}
