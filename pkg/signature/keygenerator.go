package signature

import (
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/share"
)

type KeyGenerator struct {
	suite pairing.Suite
	n     int // Total number of participants
	t     int // Threshold
}

func NewKeyGenerator(n, t int) *KeyGenerator {
	return &KeyGenerator{
		suite: bn256.NewSuiteG2(),
		n:     n,
		t:     t,
	}
}

func (kg *KeyGenerator) GenerateShares() (kyber.Scalar, []*share.PriShare, kyber.Point, []kyber.Point, error) {
	// Generate a random master private key
	privateKey := kg.suite.G2().Scalar().Pick(kg.suite.RandomStream())
	publicKey := kg.suite.G2().Point().Mul(privateKey, nil) // Master public key = g^privateKey

	// Split the private key into n shares with threshold t
	priPoly := share.NewPriPoly(kg.suite.G2(), kg.t, privateKey, kg.suite.RandomStream())
	shares := priPoly.Shares(kg.n)

	// Generate individual public keys for each share
	publicKeys := make([]kyber.Point, kg.n)
	for i, share := range shares {
		publicKeys[i] = kg.suite.G2().Point().Mul(share.V, nil) // pk_i = g^share_i
	}

	return privateKey, shares, publicKey, publicKeys, nil
}
