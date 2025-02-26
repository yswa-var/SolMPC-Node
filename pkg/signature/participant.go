package signature

import (
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/bdn"
)

type Participant struct {
	id    int
	share *share.PriShare
	suite pairing.Suite
}

func NewParticipant(id int, share *share.PriShare, suite pairing.Suite) *Participant {
	return &Participant{
		id:    id,
		share: share,
		suite: suite,
	}
}

func (p *Participant) PartialSign(message []byte) ([]byte, error) {
	// Sign the message using the participant's private key share with BDN
	partialSig, err := bdn.Sign(p.suite, p.share.V, message)
	if err != nil {
		return nil, err
	}
	return partialSig, nil
}
