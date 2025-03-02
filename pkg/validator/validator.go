package validator

import (
    "crypto/ed25519"
    "crypto/rand"
    "fmt"
    "log"
    "encoding/json"

    "github.com/nats-io/nats.go"
    "github.com/replit/tilt-validator/pkg/types"
)

type Validator struct {
    ID        string
    publicKey ed25519.PublicKey
    privateKey ed25519.PrivateKey
    nc        *nats.Conn
}

func NewValidator(id string, nc *nats.Conn) (*Validator, error) {
    pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
    if err != nil {
        return nil, fmt.Errorf("failed to generate key pair: %w", err)
    }

    return &Validator{
        ID:        id,
        publicKey: pubKey,
        privateKey: privKey,
        nc:        nc,
    }, nil
}

func (v *Validator) Run() error {
    // Generate tilt distribution
    distribution := fmt.Sprintf("Tilt Distribution from Validator %s", v.ID)

    // Sign the distribution
    signature := ed25519.Sign(v.privateKey, []byte(distribution))

    // Create signed distribution
    signedDist := types.NewSignedDistribution(
        v.ID,
        distribution,
        signature,
        v.publicKey,
    )

    // Marshal to JSON
    data, err := json.Marshal(signedDist)
    if err != nil {
        return fmt.Errorf("failed to marshal signed distribution: %w", err)
    }

    // Publish to NATS
    err = v.nc.Publish("tilt.signatures", data)
    if err != nil {
        return fmt.Errorf("failed to publish signature: %w", err)
    }

    log.Printf("Validator %s: Published signed distribution", v.ID)
    return nil
}