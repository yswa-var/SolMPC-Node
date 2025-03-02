package types

import (
    "crypto/ed25519"
    "encoding/base64"
)

// SignedDistribution represents a distribution signed by a validator
type SignedDistribution struct {
    ValidatorID  string `json:"validator_id"`
    Distribution string `json:"distribution"`
    Signature    string `json:"signature"`
    PublicKey    string `json:"public_key"`
}

// NewSignedDistribution creates a new signed distribution
func NewSignedDistribution(id string, dist string, sig []byte, pubKey ed25519.PublicKey) SignedDistribution {
    return SignedDistribution{
        ValidatorID:  id,
        Distribution: dist,
        Signature:    base64.StdEncoding.EncodeToString(sig),
        PublicKey:    base64.StdEncoding.EncodeToString(pubKey),
    }
}
