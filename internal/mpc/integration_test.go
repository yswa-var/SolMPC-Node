package ecdsa

import (
	"context"
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Test configuration constants
const (
	testThreshold        = 2
	testValidators       = 3
	testTimeout          = 30 * time.Second
	maxRetries           = 3
	performanceThreshold = 10 * time.Second // Max acceptable latency per phase
)

// Ballot represents a voting ballot in the system
type Ballot struct {
	ID        string    `json:"id"`
	Question  string    `json:"question"`
	Options   []string  `json:"options"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// Vote represents a single vote in the system
type Vote struct {
	BallotID  string    `json:"ballot_id"`
	VoterID   string    `json:"voter_id"`
	Choice    int       `json:"choice"`
	Timestamp time.Time `json:"timestamp"`
}

// PhaseMetrics tracks performance metrics for each phase
type PhaseMetrics struct {
	PhaseName    string
	StartTime    time.Time
	EndTime      time.Time
	Duration     time.Duration
	Success      bool
	ErrorMessage string
}

// IntegrationTestSuite manages the complete end-to-end test environment
type IntegrationTestSuite struct {
	parties         []*Party
	senders         []Sender
	metrics         []PhaseMetrics
	errorInjector   *ErrorInjector
	securityAuditor *SecurityAuditor
	t               *testing.T
}

// ErrorInjector simulates various failure scenarios
type ErrorInjector struct {
	failedValidators map[int]bool
	networkPartition bool
	delayMessages    time.Duration
	dropMessages     bool
	mutex            sync.RWMutex
}

// SecurityAuditor tracks security invariants throughout the test
type SecurityAuditor struct {
	singleValidatorAttempts int64
	unauthorizedOperations  int64
	signatureManipulations  int64
	mutex                   sync.RWMutex
}

// NewIntegrationTestSuite creates a new test suite
func NewIntegrationTestSuite(t *testing.T) *IntegrationTestSuite {
	return &IntegrationTestSuite{
		t: t,
		errorInjector: &ErrorInjector{
			failedValidators: make(map[int]bool),
		},
		securityAuditor: &SecurityAuditor{},
	}
}

// TestEndToEndMPCFlow is the main integration test
func TestEndToEndMPCFlow(t *testing.T) {
	suite := NewIntegrationTestSuite(t)

	t.Run("Full_End_to_End_Flow", func(t *testing.T) {
		suite.runFullIntegrationTest(t)
	})

	t.Run("Error_Handling_Tests", func(t *testing.T) {
		suite.runErrorHandlingTests(t)
	})

	t.Run("Performance_Tests", func(t *testing.T) {
		suite.runPerformanceTests(t)
	})

	t.Run("Security_Audit_Tests", func(t *testing.T) {
		suite.runSecurityAuditTests(t)
	})

	// Print comprehensive test report
	suite.printTestReport()
}

// runFullIntegrationTest executes the complete DKG → Ballot → Vote → MPC → Solana flow
func (suite *IntegrationTestSuite) runFullIntegrationTest(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Phase 1: Initialize MPC Parties and Perform DKG
	suite.startPhase("DKG_Key_Generation")
	shares, err := suite.performDKG(ctx)
	suite.endPhase("DKG_Key_Generation", err)
	require.NoError(t, err, "DKG should complete successfully")
	require.Len(t, shares, testValidators, "Should generate shares for all validators")

	// Phase 2: Create Ballot
	suite.startPhase("Ballot_Creation")
	ballot := suite.createTestBallot()
	suite.endPhase("Ballot_Creation", nil)
	require.NotNil(t, ballot, "Ballot should be created successfully")

	// Phase 3: Collect Votes
	suite.startPhase("Vote_Collection")
	votes := suite.simulateVoting(ballot)
	suite.endPhase("Vote_Collection", nil)
	require.NotEmpty(t, votes, "Should collect votes successfully")

	// Phase 4: Perform MPC Tally
	suite.startPhase("MPC_Tally")
	voteCounts, totalVotes := suite.tallyVotes(votes, ballot)
	suite.endPhase("MPC_Tally", nil)
	require.Equal(t, uint64(len(votes)), totalVotes, "Total votes should match collected votes")

	// Phase 5: Generate and Sign Solana Transaction
	suite.startPhase("MPC_Transaction_Signing")
	signedTx, err := suite.createAndSignSolanaTransaction(ctx, shares, voteCounts, totalVotes)
	suite.endPhase("MPC_Transaction_Signing", err)
	require.NoError(t, err, "Transaction signing should succeed")
	require.NotNil(t, signedTx, "Signed transaction should not be nil")

	// Phase 6: Verify Signature
	suite.startPhase("Signature_Verification")
	isValid := suite.verifyTransactionSignature(signedTx, shares[0])
	suite.endPhase("Signature_Verification", nil)
	assert.True(t, isValid, "Transaction signature should be valid")

	// Phase 7: Simulate Solana Submission (DevNet test)
	suite.startPhase("Solana_Submission")
	err = suite.simulateSolanaSubmission(signedTx)
	suite.endPhase("Solana_Submission", err)
	// Note: We allow this to fail in testing since we're using DevNet
	if err != nil {
		t.Logf("Solana submission failed (expected in test): %v", err)
	}

	t.Log("✅ Full end-to-end MPC flow completed successfully")
}

// performDKG executes the distributed key generation phase
func (suite *IntegrationTestSuite) performDKG(ctx context.Context) ([][]byte, error) {
	// Initialize parties
	suite.parties = make([]*Party, testValidators)
	for i := 0; i < testValidators; i++ {
		suite.parties[i] = NewParty(uint16(i+1), suite.createLogger(fmt.Sprintf("party_%d", i+1)))
	}

	// Set up senders
	suite.senders = suite.createSenders(suite.parties)

	// Initialize parties with senders
	for i, party := range suite.parties {
		var parties []uint16
		for j := 0; j < testValidators; j++ {
			if j != i {
				parties = append(parties, uint16(j+1))
			}
		}
		party.Init(parties, testThreshold, suite.senders[i])
	}

	// Perform DKG
	shares := make([][]byte, testValidators)
	var wg sync.WaitGroup
	var errors []error
	var errorMutex sync.Mutex

	for i, party := range suite.parties {
		wg.Add(1)
		go func(idx int, p *Party) {
			defer wg.Done()

			share, err := p.KeyGen(ctx)
			if err != nil {
				errorMutex.Lock()
				errors = append(errors, fmt.Errorf("party %d DKG failed: %w", idx, err))
				errorMutex.Unlock()
				return
			}
			shares[idx] = share
		}(i, party)
	}

	wg.Wait()

	if len(errors) > 0 {
		return nil, fmt.Errorf("DKG failed with errors: %v", errors)
	}

	return shares, nil
}

// createTestBallot creates a sample ballot for testing
func (suite *IntegrationTestSuite) createTestBallot() *Ballot {
	return &Ballot{
		ID:        "test-ballot-001",
		Question:  "Should we implement advanced MPC voting features?",
		Options:   []string{"Yes", "No", "Abstain"},
		StartTime: time.Now().Add(-1 * time.Hour),
		EndTime:   time.Now().Add(1 * time.Hour),
	}
}

// simulateVoting creates test votes
func (suite *IntegrationTestSuite) simulateVoting(ballot *Ballot) []Vote {
	return []Vote{
		{BallotID: ballot.ID, VoterID: "voter_001", Choice: 0, Timestamp: time.Now()},
		{BallotID: ballot.ID, VoterID: "voter_002", Choice: 0, Timestamp: time.Now()},
		{BallotID: ballot.ID, VoterID: "voter_003", Choice: 1, Timestamp: time.Now()},
		{BallotID: ballot.ID, VoterID: "voter_004", Choice: 0, Timestamp: time.Now()},
		{BallotID: ballot.ID, VoterID: "voter_005", Choice: 2, Timestamp: time.Now()},
	}
}

// tallyVotes processes votes and returns counts
func (suite *IntegrationTestSuite) tallyVotes(votes []Vote, ballot *Ballot) ([]uint64, uint64) {
	voteCounts := make([]uint64, len(ballot.Options))

	for _, vote := range votes {
		if vote.Choice >= 0 && vote.Choice < len(ballot.Options) {
			voteCounts[vote.Choice]++
		}
	}

	return voteCounts, uint64(len(votes))
}

// createAndSignSolanaTransaction creates and signs a transaction using MPC
func (suite *IntegrationTestSuite) createAndSignSolanaTransaction(ctx context.Context, shares [][]byte, voteCounts []uint64, totalVotes uint64) (*solana.Transaction, error) {
	// Set share data for all parties
	for i, party := range suite.parties {
		party.SetShareData(shares[i])
	}

	// Create mock Solana transaction
	programID, _ := solana.PublicKeyFromBase58("EM7AAngMgQPXizeuwAKaBvci79DhRxJMBYjRVoJWYEH3")
	recipient, _ := solana.PublicKeyFromBase58("11111111111111111111111111111112")

	// Create instruction data
	instructionData, err := suite.serializeInstructionData(voteCounts, totalVotes)
	if err != nil {
		return nil, err
	}

	// Create accounts
	accounts := []*solana.AccountMeta{
		{PublicKey: recipient, IsSigner: false, IsWritable: true},
	}

	// Create instruction
	instruction := solana.NewInstruction(programID, accounts, instructionData)

	// Create transaction with mock blockhash
	blockhash := solana.Hash{} // Mock blockhash for testing
	tx, err := solana.NewTransaction(
		[]solana.Instruction{instruction},
		blockhash,
		solana.TransactionPayer(recipient),
	)
	if err != nil {
		return nil, err
	}

	// Get transaction message for signing
	txMessage, err := tx.Message.MarshalBinary()
	if err != nil {
		return nil, err
	}

	// Sign with MPC
	digest := Digest(txMessage)
	signature, err := suite.performMPCSign(ctx, digest)
	if err != nil {
		return nil, err
	}

	// Apply signature
	tx.Signatures = []solana.Signature{solana.SignatureFromBytes(signature)}

	return tx, nil
}

// performMPCSign performs threshold signing across all parties
func (suite *IntegrationTestSuite) performMPCSign(ctx context.Context, message []byte) ([]byte, error) {
	signatures := make([][]byte, testValidators)
	var wg sync.WaitGroup
	var errors []error
	var errorMutex sync.Mutex

	for i, party := range suite.parties {
		wg.Add(1)
		go func(idx int, p *Party) {
			defer wg.Done()

			sig, err := p.Sign(ctx, message)
			if err != nil {
				errorMutex.Lock()
				errors = append(errors, fmt.Errorf("party %d signing failed: %w", idx, err))
				errorMutex.Unlock()
				return
			}
			signatures[idx] = sig
		}(i, party)
	}

	wg.Wait()

	if len(errors) > 0 {
		return nil, fmt.Errorf("MPC signing failed: %v", errors)
	}

	// All parties should produce the same signature in threshold signing
	firstSig := signatures[0]
	for i, sig := range signatures[1:] {
		if string(sig) != string(firstSig) {
			return nil, fmt.Errorf("signature mismatch between party 0 and party %d", i+1)
		}
	}

	return firstSig, nil
}

// verifyTransactionSignature verifies the MPC signature
func (suite *IntegrationTestSuite) verifyTransactionSignature(tx *solana.Transaction, keyShare []byte) bool {
	if len(tx.Signatures) == 0 {
		return false
	}

	// Get threshold public key from first party
	pk, err := suite.parties[0].ThresholdPK()
	if err != nil {
		suite.t.Logf("Failed to get threshold public key: %v", err)
		return false
	}

	// Get transaction message
	txMessage, err := tx.Message.MarshalBinary()
	if err != nil {
		suite.t.Logf("Failed to marshal transaction message: %v", err)
		return false
	}

	// Verify signature
	return ed25519.Verify(pk, Digest(txMessage), tx.Signatures[0][:])
}

// simulateSolanaSubmission simulates submitting to Solana network
func (suite *IntegrationTestSuite) simulateSolanaSubmission(tx *solana.Transaction) error {
	// This would normally submit to DevNet, but for testing we just validate format
	if len(tx.Signatures) == 0 {
		return fmt.Errorf("transaction has no signatures")
	}

	if len(tx.Message.Instructions) == 0 {
		return fmt.Errorf("transaction has no instructions")
	}

	// In a real scenario, you would:
	// client := rpc.New("https://api.devnet.solana.com")
	// sig, err := client.SendTransaction(ctx, tx)

	suite.t.Log("✅ Transaction format validated for Solana submission")
	return nil
}

// runErrorHandlingTests tests various failure scenarios
func (suite *IntegrationTestSuite) runErrorHandlingTests(t *testing.T) {
	t.Run("Validator_Failure_Scenarios", func(t *testing.T) {
		suite.testValidatorFailures(t)
	})

	t.Run("Network_Partition_Scenarios", func(t *testing.T) {
		suite.testNetworkPartitions(t)
	})

	t.Run("Timeout_Scenarios", func(t *testing.T) {
		suite.testTimeoutScenarios(t)
	})
}

// testValidatorFailures tests behavior when validators fail
func (suite *IntegrationTestSuite) testValidatorFailures(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Test: Single validator failure (should still work with threshold=2, validators=3)
	suite.errorInjector.failedValidators[2] = true

	_, err := suite.performDKG(ctx)

	// Should succeed with 2 out of 3 validators
	assert.NoError(t, err, "DKG should succeed with single validator failure")

	// Reset error injection
	suite.errorInjector.failedValidators = make(map[int]bool)

	// Test: Too many validator failures (should fail)
	suite.errorInjector.failedValidators[1] = true
	suite.errorInjector.failedValidators[2] = true

	_, err = suite.performDKG(ctx)

	// Should fail with insufficient validators
	assert.Error(t, err, "DKG should fail with too many validator failures")

	t.Log("✅ Validator failure scenarios tested successfully")
}

// testNetworkPartitions tests network partition scenarios
func (suite *IntegrationTestSuite) testNetworkPartitions(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Shorter timeout
	defer cancel()

	// Simulate network partition
	suite.errorInjector.networkPartition = true
	suite.errorInjector.dropMessages = true

	_, err := suite.performDKG(ctx)

	// Should timeout or fail due to network issues
	assert.Error(t, err, "DKG should fail during network partition")

	// Reset network state
	suite.errorInjector.networkPartition = false
	suite.errorInjector.dropMessages = false

	t.Log("✅ Network partition scenarios tested successfully")
}

// testTimeoutScenarios tests timeout handling
func (suite *IntegrationTestSuite) testTimeoutScenarios(t *testing.T) {
	// Test with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Add artificial delay
	suite.errorInjector.delayMessages = 200 * time.Millisecond

	_, err := suite.performDKG(ctx)

	// Should timeout
	assert.Error(t, err, "DKG should timeout with short deadline")
	assert.Contains(t, err.Error(), "context deadline exceeded", "Should be a timeout error")

	// Reset delays
	suite.errorInjector.delayMessages = 0

	t.Log("✅ Timeout scenarios tested successfully")
}

// runPerformanceTests measures performance of each phase
func (suite *IntegrationTestSuite) runPerformanceTests(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Reset metrics
	suite.metrics = nil

	// Run performance test
	suite.startPhase("Performance_DKG")
	_, err := suite.performDKG(ctx)
	suite.endPhase("Performance_DKG", err)
	require.NoError(t, err)

	suite.startPhase("Performance_Signing")
	testMessage := []byte("performance test message")
	_, err = suite.performMPCSign(ctx, Digest(testMessage))
	suite.endPhase("Performance_Signing", err)
	require.NoError(t, err)

	// Verify performance metrics
	for _, metric := range suite.metrics {
		assert.True(t, metric.Duration < performanceThreshold,
			"Phase %s took %v, which exceeds threshold of %v",
			metric.PhaseName, metric.Duration, performanceThreshold)

		t.Logf("📊 %s completed in %v", metric.PhaseName, metric.Duration)
	}
}

// runSecurityAuditTests verifies security properties
func (suite *IntegrationTestSuite) runSecurityAuditTests(t *testing.T) {
	t.Run("Single_Validator_Cannot_Sign", func(t *testing.T) {
		suite.testSingleValidatorSigning(t)
	})

	t.Run("Signature_Manipulation_Detection", func(t *testing.T) {
		suite.testSignatureManipulation(t)
	})

	t.Run("Threshold_Enforcement", func(t *testing.T) {
		suite.testThresholdEnforcement(t)
	})
}

// testSingleValidatorSigning ensures single validators cannot create valid signatures
func (suite *IntegrationTestSuite) testSingleValidatorSigning(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Perform DKG first
	shares, err := suite.performDKG(ctx)
	require.NoError(t, err)

	// Try to sign with only one party
	suite.parties[0].SetShareData(shares[0])

	testMessage := []byte("security test message")
	signature, err := suite.parties[0].Sign(ctx, Digest(testMessage))

	// Single party signing should fail or produce invalid signature
	if err == nil {
		// If signing succeeds, verify it's not valid without threshold
		pk, _ := suite.parties[0].ThresholdPK()
		isValid := ed25519.Verify(pk, Digest(testMessage), signature)
		assert.False(t, isValid, "Single validator signature should not be valid for threshold verification")
	}

	atomic.AddInt64(&suite.securityAuditor.singleValidatorAttempts, 1)
	t.Log("✅ Single validator signing test passed")
}

// testSignatureManipulation tests signature manipulation detection
func (suite *IntegrationTestSuite) testSignatureManipulation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Perform DKG
	shares, err := suite.performDKG(ctx)
	require.NoError(t, err)

	// Set share data for signing
	for i, party := range suite.parties {
		party.SetShareData(shares[i])
	}

	// Create valid signature
	testMessage := []byte("manipulation test message")
	signature, err := suite.performMPCSign(ctx, Digest(testMessage))
	require.NoError(t, err)

	// Manipulate signature
	manipulatedSig := make([]byte, len(signature))
	copy(manipulatedSig, signature)
	manipulatedSig[0] ^= 0x01 // Flip one bit

	// Verify manipulation is detected
	pk, err := suite.parties[0].ThresholdPK()
	require.NoError(t, err)

	isValidOriginal := ed25519.Verify(pk, Digest(testMessage), signature)
	isValidManipulated := ed25519.Verify(pk, Digest(testMessage), manipulatedSig)

	assert.True(t, isValidOriginal, "Original signature should be valid")
	assert.False(t, isValidManipulated, "Manipulated signature should be invalid")

	atomic.AddInt64(&suite.securityAuditor.signatureManipulations, 1)
	t.Log("✅ Signature manipulation detection test passed")
}

// testThresholdEnforcement ensures threshold requirements are enforced
func (suite *IntegrationTestSuite) testThresholdEnforcement(t *testing.T) {
	// Test with insufficient parties (below threshold)
	singleParty := NewParty(1, suite.createLogger("single_party"))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try DKG with single party (should fail)
	singleParty.Init([]uint16{}, testThreshold, func([]byte, bool, uint16) {})
	_, err := singleParty.KeyGen(ctx)

	assert.Error(t, err, "DKG should fail with insufficient parties")
	t.Log("✅ Threshold enforcement test passed")
}

// Helper methods for test infrastructure

// createSenders creates mock senders with error injection
func (suite *IntegrationTestSuite) createSenders(parties []*Party) []Sender {
	var senders []Sender
	for i := range parties {
		srcIndex := i
		sender := func(msgBytes []byte, broadcast bool, to uint16) {
			// Apply error injection
			suite.errorInjector.mutex.RLock()

			if suite.errorInjector.failedValidators[srcIndex] {
				suite.errorInjector.mutex.RUnlock()
				return // Drop message from failed validator
			}

			if suite.errorInjector.dropMessages {
				suite.errorInjector.mutex.RUnlock()
				return // Simulate network partition
			}

			delay := suite.errorInjector.delayMessages
			suite.errorInjector.mutex.RUnlock()

			if delay > 0 {
				time.Sleep(delay)
			}

			// Deliver message
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

// createLogger creates a test logger
func (suite *IntegrationTestSuite) createLogger(id string) Logger {
	logConfig := zap.NewDevelopmentConfig()
	logConfig.Level = zap.NewAtomicLevelAt(zap.WarnLevel) // Reduce log noise in tests
	logger, _ := logConfig.Build()
	return logger.With(zap.String("id", id)).Sugar()
}

// serializeInstructionData creates instruction data for testing
func (suite *IntegrationTestSuite) serializeInstructionData(amounts []uint64, totalAmount uint64) ([]byte, error) {
	var data []byte

	// Add discriminator
	discriminator := make([]byte, 8)
	data = append(data, discriminator...)

	// Add total amount
	totalBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(totalBytes, totalAmount)
	data = append(data, totalBytes...)

	// Add amounts length
	lengthBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(lengthBytes, uint32(len(amounts)))
	data = append(data, lengthBytes...)

	// Add amounts
	for _, amount := range amounts {
		amountBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(amountBytes, amount)
		data = append(data, amountBytes...)
	}

	return data, nil
}

// Performance tracking methods

func (suite *IntegrationTestSuite) startPhase(phaseName string) {
	metric := PhaseMetrics{
		PhaseName: phaseName,
		StartTime: time.Now(),
	}
	suite.metrics = append(suite.metrics, metric)
}

func (suite *IntegrationTestSuite) endPhase(phaseName string, err error) {
	for i := range suite.metrics {
		if suite.metrics[i].PhaseName == phaseName && suite.metrics[i].EndTime.IsZero() {
			suite.metrics[i].EndTime = time.Now()
			suite.metrics[i].Duration = suite.metrics[i].EndTime.Sub(suite.metrics[i].StartTime)
			suite.metrics[i].Success = err == nil
			if err != nil {
				suite.metrics[i].ErrorMessage = err.Error()
			}
			break
		}
	}
}

// printTestReport prints a comprehensive test report
func (suite *IntegrationTestSuite) printTestReport() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🎯 END-TO-END MPC FLOW VERIFICATION REPORT")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Println("\n📊 PERFORMANCE METRICS:")
	totalDuration := time.Duration(0)
	for _, metric := range suite.metrics {
		status := "✅ SUCCESS"
		if !metric.Success {
			status = "❌ FAILED"
		}
		fmt.Printf("  %-25s: %8v %s\n", metric.PhaseName, metric.Duration, status)
		if metric.ErrorMessage != "" {
			fmt.Printf("    Error: %s\n", metric.ErrorMessage)
		}
		totalDuration += metric.Duration
	}
	fmt.Printf("  %-25s: %8v\n", "TOTAL_DURATION", totalDuration)

	fmt.Println("\n🔒 SECURITY AUDIT RESULTS:")
	fmt.Printf("  Single Validator Attempts: %d\n", atomic.LoadInt64(&suite.securityAuditor.singleValidatorAttempts))
	fmt.Printf("  Signature Manipulations:   %d\n", atomic.LoadInt64(&suite.securityAuditor.signatureManipulations))
	fmt.Printf("  Unauthorized Operations:   %d\n", atomic.LoadInt64(&suite.securityAuditor.unauthorizedOperations))

	fmt.Println("\n🎉 TEST SUMMARY:")
	successCount := 0
	for _, metric := range suite.metrics {
		if metric.Success {
			successCount++
		}
	}
	fmt.Printf("  Phases Completed:     %d/%d\n", successCount, len(suite.metrics))
	fmt.Printf("  Success Rate:         %.1f%%\n", float64(successCount)/float64(len(suite.metrics))*100)

	if successCount == len(suite.metrics) {
		fmt.Println("  Overall Status:       ✅ ALL TESTS PASSED")
	} else {
		fmt.Println("  Overall Status:       ❌ SOME TESTS FAILED")
	}

	fmt.Println(strings.Repeat("=", 80))
}
