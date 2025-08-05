package ballot

import (
	"context"
	"fmt"

	mpc "tilt-valid/internal/mpc"
)

// MPCBallotService wraps an existing MPC party to provide ballot-specific functionality
type MPCBallotService struct {
	party *mpc.Party
}

// NewMPCBallotService creates a new MPC ballot service wrapper
func NewMPCBallotService(party *mpc.Party) *MPCBallotService {
	return &MPCBallotService{
		party: party,
	}
}

// Sign signs a message hash using MPC
func (mbs *MPCBallotService) Sign(ctx context.Context, msgHash []byte) ([]byte, error) {
	if mbs.party == nil {
		return nil, fmt.Errorf("MPC party not initialized")
	}

	return mbs.party.Sign(ctx, msgHash)
}

// GetPublicKey returns the MPC public key
func (mbs *MPCBallotService) GetPublicKey() ([]byte, error) {
	if mbs.party == nil {
		return nil, fmt.Errorf("MPC party not initialized")
	}

	// For now, return a placeholder - this would need to be implemented
	// based on the actual MPC party structure
	return nil, fmt.Errorf("public key retrieval not yet implemented")
}

// IsReady checks if the MPC service is ready for operations
func (mbs *MPCBallotService) IsReady() bool {
	return mbs.party != nil
}

// BallotLogger wraps the existing logger to provide ballot-specific logging
type BallotLogger struct {
	logger mpc.Logger
}

// NewBallotLogger creates a new ballot logger wrapper
func NewBallotLogger(logger mpc.Logger) *BallotLogger {
	return &BallotLogger{
		logger: logger,
	}
}

// Debugf logs a debug message
func (bl *BallotLogger) Debugf(format string, a ...interface{}) {
	bl.logger.Debugf("[BALLOT] "+format, a...)
}

// Warnf logs a warning message
func (bl *BallotLogger) Warnf(format string, a ...interface{}) {
	bl.logger.Warnf("[BALLOT] "+format, a...)
}

// Errorf logs an error message
func (bl *BallotLogger) Errorf(format string, a ...interface{}) {
	bl.logger.Errorf("[BALLOT] "+format, a...)
}

// Infof logs an info message
func (bl *BallotLogger) Infof(format string, a ...interface{}) {
	bl.logger.Infof("[BALLOT] "+format, a...)
}

// IntegrationManager manages the integration between ballot system and existing MPC infrastructure
type IntegrationManager struct {
	ballotService *BallotService
	mpcService    *MPCBallotService
	logger        mpc.Logger
	storage       BallotStorage
}

// NewIntegrationManager creates a new integration manager
func NewIntegrationManager(mpcParty *mpc.Party, mpcLogger mpc.Logger, storagePath string) (*IntegrationManager, error) {
	// Create storage
	storage, err := NewFileStorage(storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	// Create MPC service wrapper
	mpcService := NewMPCBallotService(mpcParty)

	// Create logger wrapper
	logger := NewBallotLogger(mpcLogger)

	// Create ballot service
	ballotService, err := NewBallotService(storage, mpcService, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create ballot service: %w", err)
	}

	return &IntegrationManager{
		ballotService: ballotService,
		mpcService:    mpcService,
		logger:        logger,
		storage:       storage,
	}, nil
}

// GetBallotService returns the ballot service
func (im *IntegrationManager) GetBallotService() *BallotService {
	return im.ballotService
}

// Start starts all integrated services
func (im *IntegrationManager) Start() error {
	im.logger.Infof("Starting ballot integration manager...")

	if err := im.ballotService.Start(); err != nil {
		return fmt.Errorf("failed to start ballot service: %w", err)
	}

	im.logger.Infof("Ballot integration manager started successfully")
	return nil
}

// Stop stops all integrated services
func (im *IntegrationManager) Stop() error {
	im.logger.Infof("Stopping ballot integration manager...")

	if err := im.ballotService.Stop(); err != nil {
		im.logger.Warnf("Error stopping ballot service: %v", err)
	}

	im.logger.Infof("Ballot integration manager stopped")
	return nil
}
