package validator

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GenerateSecretShares creates secret shares using Shamir's Secret Sharing
// This function generates a random secret and creates a polynomial of degree (threshold - 1).
// It then evaluates the polynomial at different points to generate private shares for each validator.
// Public shares are simulated by squaring the private shares.
func GenerateSecretShares(totalValidators, threshold int) ([]*big.Int, []*big.Int) {
	// Generate a random secret key
	secret, _ := rand.Int(rand.Reader, big.NewInt(1e18))
	fmt.Println("🔐 Original Secret Key:", secret)

	// Generate a polynomial of degree (threshold - 1)
	coefficients := make([]*big.Int, threshold)
	coefficients[0] = secret // The constant term (secret)
	for i := 1; i < threshold; i++ {
		coefficients[i], _ = rand.Int(rand.Reader, big.NewInt(1e18))
	}

	// Generate shares for each validator
	privateShares := make([]*big.Int, totalValidators)
	publicShares := make([]*big.Int, totalValidators)

	for i := 1; i <= totalValidators; i++ {
		x := big.NewInt(int64(i))
		privateShares[i-1] = evaluatePolynomial(coefficients, x)

		// Simulate public share (in real implementation, use elliptic curve cryptography)
		publicShares[i-1] = new(big.Int).Exp(privateShares[i-1], big.NewInt(2), nil)
	}

	return privateShares, publicShares
}

// Evaluate polynomial at a given x-value (Horner's method)
// This function evaluates a polynomial at a given x-value using Horner's method.
// Horner's method is an efficient way to evaluate polynomials.
func evaluatePolynomial(coeffs []*big.Int, x *big.Int) *big.Int {
	result := new(big.Int).Set(coeffs[0])

	for i := 1; i < len(coeffs); i++ {
		result.Mul(result, x)
		result.Add(result, coeffs[i])
	}

	return result
}

// DistributeShares assigns private and public shares to validators
// This function distributes the generated private and public shares to the validators.
func DistributeShares(validators []*Validator, privateShares, publicShares []*big.Int) {
	for i, v := range validators {
		v.PrivateKeyShare = privateShares[i]
		v.PublicShare = publicShares[i]
		fmt.Printf("🔑 Validator %d -> Private Share: %v, Public Share: %v\n",
			v.ID, v.PrivateKeyShare, v.PublicShare)
	}
}

// ComputeGroupPublicKey combines public shares to create a group public key
// This function combines the public shares of all validators to create a group public key.
func ComputeGroupPublicKey(validators []*Validator) *big.Int {
	groupPublicKey := big.NewInt(0)
	for _, v := range validators {
		groupPublicKey.Add(groupPublicKey, v.PublicShare)
	}
	fmt.Println("🌐 Final Group Public Key:", groupPublicKey)
	return groupPublicKey
}

// ReconstructSecret demonstrates how threshold validators can reconstruct the secret
// This function simulates the reconstruction of the secret using the private shares of the validators.
// In a real system, Lagrange interpolation would be used.
func ReconstructSecret(validators []*Validator, threshold int) *big.Int {
	// In a real system, you'd use Lagrange interpolation
	// This is a simplified simulation
	reconstructedSecret := big.NewInt(0)
	for i := 0; i < threshold; i++ {
		reconstructedSecret.Add(reconstructedSecret, validators[i].PrivateKeyShare)
	}

	fmt.Println("🔓 Reconstructed Secret:", reconstructedSecret)
	return reconstructedSecret
}
