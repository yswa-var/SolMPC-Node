package ecdsa

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// BenchmarkDKG benchmarks the Distributed Key Generation process
func BenchmarkDKG(b *testing.B) {
	suite := NewIntegrationTestSuite(&testing.T{})

	// Initialize parties once
	suite.parties = make([]*Party, testValidators)
	for i := 0; i < testValidators; i++ {
		suite.parties[i] = NewParty(uint16(i+1), suite.createLogger(fmt.Sprintf("bench_party_%d", i+1)))
	}

	suite.senders = suite.createSenders(suite.parties)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Initialize parties with senders for each iteration
		for idx, party := range suite.parties {
			var parties []uint16
			for j := 0; j < testValidators; j++ {
				if j != idx {
					parties = append(parties, uint16(j+1))
				}
			}
			party.Init(parties, testThreshold, suite.senders[idx])
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

		_, err := suite.performDKG(ctx)
		require.NoError(b, err)

		cancel()
	}
}

// BenchmarkMPCSigning benchmarks the MPC signing process
func BenchmarkMPCSigning(b *testing.B) {
	suite := NewIntegrationTestSuite(&testing.T{})

	// Setup DKG once
	suite.parties = make([]*Party, testValidators)
	for i := 0; i < testValidators; i++ {
		suite.parties[i] = NewParty(uint16(i+1), suite.createLogger(fmt.Sprintf("bench_party_%d", i+1)))
	}

	suite.senders = suite.createSenders(suite.parties)

	// Initialize parties
	for idx, party := range suite.parties {
		var parties []uint16
		for j := 0; j < testValidators; j++ {
			if j != idx {
				parties = append(parties, uint16(j+1))
			}
		}
		party.Init(parties, testThreshold, suite.senders[idx])
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	shares, err := suite.performDKG(ctx)
	require.NoError(b, err)

	// Set share data
	for i, party := range suite.parties {
		party.SetShareData(shares[i])
	}

	testMessage := []byte("benchmark test message")
	digest := Digest(testMessage)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := suite.performMPCSign(ctx, digest)
		require.NoError(b, err)
	}
}

// BenchmarkFullFlow benchmarks the complete end-to-end flow
func BenchmarkFullFlow(b *testing.B) {
	suite := NewIntegrationTestSuite(&testing.T{})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		// DKG
		suite.parties = make([]*Party, testValidators)
		for j := 0; j < testValidators; j++ {
			suite.parties[j] = NewParty(uint16(j+1), suite.createLogger(fmt.Sprintf("bench_party_%d", j+1)))
		}

		suite.senders = suite.createSenders(suite.parties)

		for idx, party := range suite.parties {
			var parties []uint16
			for k := 0; k < testValidators; k++ {
				if k != idx {
					parties = append(parties, uint16(k+1))
				}
			}
			party.Init(parties, testThreshold, suite.senders[idx])
		}

		shares, err := suite.performDKG(ctx)
		require.NoError(b, err)

		// Create ballot and votes
		ballot := suite.createTestBallot()
		votes := suite.simulateVoting(ballot)
		voteCounts, totalVotes := suite.tallyVotes(votes, ballot)

		// Sign transaction
		_, err = suite.createAndSignSolanaTransaction(ctx, shares, voteCounts, totalVotes)
		require.NoError(b, err)

		cancel()
	}
}

// BenchmarkValidatorScaling tests performance with different numbers of validators
func BenchmarkValidatorScaling(b *testing.B) {
	validatorCounts := []int{3, 5, 7, 10}

	for _, count := range validatorCounts {
		b.Run(fmt.Sprintf("Validators_%d", count), func(b *testing.B) {
			suite := NewIntegrationTestSuite(&testing.T{})
			threshold := (count * 2) / 3 // 2/3 threshold

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// Initialize parties
				suite.parties = make([]*Party, count)
				for j := 0; j < count; j++ {
					suite.parties[j] = NewParty(uint16(j+1), suite.createLogger(fmt.Sprintf("scale_party_%d", j+1)))
				}

				suite.senders = suite.createScalableSenders(suite.parties)

				for idx, party := range suite.parties {
					var parties []uint16
					for k := 0; k < count; k++ {
						if k != idx {
							parties = append(parties, uint16(k+1))
						}
					}
					party.Init(parties, threshold, suite.senders[idx])
				}

				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(count*5)*time.Second)

				_, err := suite.performDKG(ctx)
				require.NoError(b, err)

				cancel()
			}
		})
	}
}

// createScalableSenders creates senders for scalability testing
func (suite *IntegrationTestSuite) createScalableSenders(parties []*Party) []Sender {
	var senders []Sender
	for i := range parties {
		srcIndex := i
		sender := func(msgBytes []byte, broadcast bool, to uint16) {
			messageSource := uint16(srcIndex + 1)
			if broadcast {
				for j, dst := range parties {
					if j == srcIndex {
						continue
					}
					dst.OnMsg(msgBytes, messageSource, broadcast)
				}
			} else {
				for j, dst := range parties {
					if uint16(j+1) != to {
						continue
					}
					dst.OnMsg(msgBytes, messageSource, broadcast)
				}
			}
		}
		senders = append(senders, sender)
	}
	return senders
}

// BenchmarkMessageThroughput tests message handling throughput
func BenchmarkMessageThroughput(b *testing.B) {
	suite := NewIntegrationTestSuite(&testing.T{})

	// Setup parties
	suite.parties = make([]*Party, testValidators)
	for i := 0; i < testValidators; i++ {
		suite.parties[i] = NewParty(uint16(i+1), suite.createLogger(fmt.Sprintf("throughput_party_%d", i+1)))
	}

	suite.senders = suite.createSenders(suite.parties)

	for idx, party := range suite.parties {
		var parties []uint16
		for j := 0; j < testValidators; j++ {
			if j != idx {
				parties = append(parties, uint16(j+1))
			}
		}
		party.Init(parties, testThreshold, suite.senders[idx])
	}

	// Start DKG to generate some realistic message traffic
	_, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	testMessage := []byte("throughput test message")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate message broadcast
		for _, sender := range suite.senders {
			sender(testMessage, true, 0)
		}
	}
}

// BenchmarkMemoryUsage tests memory usage patterns
func BenchmarkMemoryUsage(b *testing.B) {
	suite := NewIntegrationTestSuite(&testing.T{})

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Initialize parties
		suite.parties = make([]*Party, testValidators)
		for j := 0; j < testValidators; j++ {
			suite.parties[j] = NewParty(uint16(j+1), suite.createLogger(fmt.Sprintf("memory_party_%d", j+1)))
		}

		// Setup senders and initialize
		suite.senders = suite.createSenders(suite.parties)

		for idx, party := range suite.parties {
			var parties []uint16
			for k := 0; k < testValidators; k++ {
				if k != idx {
					parties = append(parties, uint16(k+1))
				}
			}
			party.Init(parties, testThreshold, suite.senders[idx])
		}

		// Perform DKG to measure memory usage
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

		shares, err := suite.performDKG(ctx)
		require.NoError(b, err)

		// Set share data
		for k, party := range suite.parties {
			party.SetShareData(shares[k])
		}

		// Perform signing
		testMessage := []byte("memory test message")
		_, err = suite.performMPCSign(ctx, Digest(testMessage))
		require.NoError(b, err)

		cancel()

		// Clear references to help GC
		suite.parties = nil
		suite.senders = nil
	}
}
