package ballot

import (
	"context"
	"fmt"
	"sync"
	"time"

	mpc "tilt-valid/internal/mpc"
)

// Database interface for future database integration
type Database interface {
	// Future database operations will be defined here
	Connect() error
	Close() error
	Ping() error
}

// MPCService interface for MPC operations
type MPCService interface {
	Sign(ctx context.Context, msgHash []byte) ([]byte, error)
	GetPublicKey() ([]byte, error)
	IsReady() bool
}

// BallotService manages the complete ballot lifecycle
type BallotService struct {
	storage   BallotStorage
	mpc       MPCService
	db        Database
	logger    mpc.Logger
	scheduler *BallotScheduler
	templates map[string]*BallotTemplate
	mutex     sync.RWMutex
}

// NewBallotService creates a new ballot service instance
func NewBallotService(storage BallotStorage, mpcService MPCService, logger mpc.Logger) (*BallotService, error) {
	if storage == nil {
		return nil, fmt.Errorf("storage cannot be nil")
	}

	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	service := &BallotService{
		storage:   storage,
		mpc:       mpcService,
		logger:    logger,
		templates: make(map[string]*BallotTemplate),
	}

	// Initialize scheduler
	scheduler := NewBallotScheduler(service, logger)
	service.scheduler = scheduler

	// Load default templates
	if err := service.loadDefaultTemplates(); err != nil {
		logger.Warnf("Failed to load default ballot templates: %v", err)
	}

	return service, nil
}

// CreateBallot creates a new ballot
func (bs *BallotService) CreateBallot(ballot *Ballot) error {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	// Validate ballot
	if err := ballot.Validate(); err != nil {
		return fmt.Errorf("ballot validation failed: %w", err)
	}

	// Set initial status
	ballot.Status = StatusDraft
	ballot.CreatedAt = time.Now()
	ballot.UpdatedAt = time.Now()

	// Save to storage
	if err := bs.storage.SaveBallot(ballot); err != nil {
		return fmt.Errorf("failed to save ballot: %w", err)
	}

	bs.logger.Infof("Created ballot: %s (%s)", ballot.Title, ballot.ID)
	return nil
}

// GetBallot retrieves a ballot by ID
func (bs *BallotService) GetBallot(id string) (*Ballot, error) {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()

	ballot, err := bs.storage.GetBallot(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get ballot: %w", err)
	}

	return ballot, nil
}

// UpdateBallot updates an existing ballot
func (bs *BallotService) UpdateBallot(ballot *Ballot) error {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	// Validate ballot
	if err := ballot.Validate(); err != nil {
		return fmt.Errorf("ballot validation failed: %w", err)
	}

	// Check if ballot exists and can be updated
	existing, err := bs.storage.GetBallot(ballot.ID)
	if err != nil {
		return fmt.Errorf("ballot not found: %w", err)
	}

	// Don't allow updates to closed or archived ballots
	if existing.Status == StatusClosed || existing.Status == StatusArchived {
		return fmt.Errorf("cannot update ballot in %s status", existing.Status)
	}

	ballot.UpdatedAt = time.Now()

	if err := bs.storage.UpdateBallot(ballot); err != nil {
		return fmt.Errorf("failed to update ballot: %w", err)
	}

	bs.logger.Infof("Updated ballot: %s (%s)", ballot.Title, ballot.ID)
	return nil
}

// ActivateBallot transitions a ballot from draft to active status
func (bs *BallotService) ActivateBallot(id string) error {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	ballot, err := bs.storage.GetBallot(id)
	if err != nil {
		return fmt.Errorf("ballot not found: %w", err)
	}

	if ballot.Status != StatusDraft {
		return fmt.Errorf("can only activate ballots in draft status, current status: %s", ballot.Status)
	}

	// Validate timing
	now := time.Now()
	if ballot.StartTime.Before(now) {
		ballot.StartTime = now // Auto-adjust start time if in the past
	}

	if ballot.EndTime.Before(ballot.StartTime) {
		return fmt.Errorf("end time must be after start time")
	}

	ballot.Status = StatusActive
	ballot.UpdatedAt = time.Now()

	if err := bs.storage.UpdateBallot(ballot); err != nil {
		return fmt.Errorf("failed to activate ballot: %w", err)
	}

	// Schedule automatic closure
	bs.scheduler.ScheduleBallotClosure(ballot.ID, ballot.EndTime)

	bs.logger.Infof("Activated ballot: %s (%s)", ballot.Title, ballot.ID)
	return nil
}

// CloseBallot closes an active ballot and triggers vote tallying
func (bs *BallotService) CloseBallot(id string) error {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	ballot, err := bs.storage.GetBallot(id)
	if err != nil {
		return fmt.Errorf("ballot not found: %w", err)
	}

	if ballot.Status != StatusActive {
		return fmt.Errorf("can only close active ballots, current status: %s", ballot.Status)
	}

	ballot.Status = StatusClosed
	ballot.UpdatedAt = time.Now()

	if err := bs.storage.UpdateBallot(ballot); err != nil {
		return fmt.Errorf("failed to close ballot: %w", err)
	}

	// Cancel any scheduled closure
	bs.scheduler.CancelBallotClosure(id)

	// Trigger vote tallying in background
	go func() {
		if err := bs.tallyVotes(id); err != nil {
			bs.logger.Errorf("Failed to tally votes for ballot %s: %v", id, err)
		}
	}()

	bs.logger.Infof("Closed ballot: %s (%s)", ballot.Title, ballot.ID)
	return nil
}

// ArchiveBallot moves a closed ballot to archived status
func (bs *BallotService) ArchiveBallot(id string) error {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	ballot, err := bs.storage.GetBallot(id)
	if err != nil {
		return fmt.Errorf("ballot not found: %w", err)
	}

	if ballot.Status != StatusClosed {
		return fmt.Errorf("can only archive closed ballots, current status: %s", ballot.Status)
	}

	ballot.Status = StatusArchived
	ballot.UpdatedAt = time.Now()

	if err := bs.storage.UpdateBallot(ballot); err != nil {
		return fmt.Errorf("failed to archive ballot: %w", err)
	}

	bs.logger.Infof("Archived ballot: %s (%s)", ballot.Title, ballot.ID)
	return nil
}

// ListBallots returns ballots filtered by status
func (bs *BallotService) ListBallots(status BallotStatus) ([]*Ballot, error) {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()

	ballots, err := bs.storage.ListBallots(status)
	if err != nil {
		return nil, fmt.Errorf("failed to list ballots: %w", err)
	}

	return ballots, nil
}

// CastVote records a vote for a ballot
func (bs *BallotService) CastVote(vote *Vote) error {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	// Get ballot and validate
	ballot, err := bs.storage.GetBallot(vote.BallotID)
	if err != nil {
		return fmt.Errorf("ballot not found: %w", err)
	}

	if !ballot.CanVote(vote.VoterID) {
		return fmt.Errorf("voter %s is not eligible to vote on ballot %s", vote.VoterID, vote.BallotID)
	}

	// Check for duplicate votes
	existingVotes, err := bs.storage.GetVotesByVoter(vote.VoterID)
	if err != nil {
		return fmt.Errorf("failed to check existing votes: %w", err)
	}

	for _, existing := range existingVotes {
		if existing.BallotID == vote.BallotID {
			return fmt.Errorf("voter %s has already voted on ballot %s", vote.VoterID, vote.BallotID)
		}
	}

	// Validate vote choices
	if err := bs.validateVoteChoices(ballot, vote); err != nil {
		return fmt.Errorf("invalid vote choices: %w", err)
	}

	// Set timestamp and save
	vote.Timestamp = time.Now()

	if err := bs.storage.SaveVote(vote); err != nil {
		return fmt.Errorf("failed to save vote: %w", err)
	}

	bs.logger.Infof("Vote cast by %s on ballot %s", vote.VoterID, vote.BallotID)
	return nil
}

// GetTallyResult returns the tally results for a ballot
func (bs *BallotService) GetTallyResult(ballotID string) (*TallyResult, error) {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()

	result, err := bs.storage.GetTallyResult(ballotID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tally result: %w", err)
	}

	return result, nil
}

// CreateFromTemplate creates a new ballot from a template
func (bs *BallotService) CreateFromTemplate(templateID, title, description, createdBy string, startTime, endTime time.Time) (*Ballot, error) {
	bs.mutex.RLock()
	template, exists := bs.templates[templateID]
	bs.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}

	ballot := NewBallot(title, description, template.Type, createdBy)
	ballot.Options = template.Options
	ballot.StartTime = startTime
	ballot.EndTime = endTime
	ballot.RequiresAuth = template.Settings.RequiresAuth
	ballot.AllowAnonymous = template.Settings.AllowAnonymous
	ballot.MaxVotesPerVoter = template.Settings.MaxVotesPerVoter

	if err := bs.CreateBallot(ballot); err != nil {
		return nil, fmt.Errorf("failed to create ballot from template: %w", err)
	}

	return ballot, nil
}

// Start starts the ballot service and scheduler
func (bs *BallotService) Start() error {
	bs.logger.Infof("Starting ballot service...")

	// Start scheduler
	if bs.scheduler != nil {
		bs.scheduler.Start()
	}

	// Recover any active ballots that should be scheduled
	activeballots, err := bs.ListBallots(StatusActive)
	if err != nil {
		bs.logger.Warnf("Failed to recover active ballots: %v", err)
	} else {
		for _, ballot := range activeballots {
			if ballot.EndTime.After(time.Now()) {
				bs.scheduler.ScheduleBallotClosure(ballot.ID, ballot.EndTime)
			}
		}
	}

	bs.logger.Infof("Ballot service started successfully")
	return nil
}

// Stop stops the ballot service and scheduler
func (bs *BallotService) Stop() error {
	bs.logger.Infof("Stopping ballot service...")

	if bs.scheduler != nil {
		bs.scheduler.Stop()
	}

	bs.logger.Infof("Ballot service stopped")
	return nil
}

// validateVoteChoices validates that vote choices are valid for the ballot
func (bs *BallotService) validateVoteChoices(ballot *Ballot, vote *Vote) error {
	if len(vote.Choices) == 0 {
		return fmt.Errorf("vote must have at least one choice")
	}

	if len(vote.Choices) > ballot.MaxVotesPerVoter {
		return fmt.Errorf("vote has %d choices but ballot allows maximum %d", len(vote.Choices), ballot.MaxVotesPerVoter)
	}

	// Validate that all option IDs exist in the ballot
	validOptions := make(map[string]bool)
	for _, option := range ballot.Options {
		validOptions[option.ID] = true
	}

	for _, choice := range vote.Choices {
		if !validOptions[choice.OptionID] {
			return fmt.Errorf("invalid option ID: %s", choice.OptionID)
		}
	}

	return nil
}

// tallyVotes performs MPC-based vote tallying for a closed ballot
func (bs *BallotService) tallyVotes(ballotID string) error {
	bs.logger.Infof("Starting vote tally for ballot: %s", ballotID)

	// Get all votes for the ballot
	votes, err := bs.storage.GetVotesByBallot(ballotID)
	if err != nil {
		return fmt.Errorf("failed to get votes: %w", err)
	}

	// Compute tally results
	results := make(map[string]int)
	totalVotes := len(votes)

	for _, vote := range votes {
		for _, choice := range vote.Choices {
			results[choice.OptionID]++
		}
	}

	// Create tally result
	tallyResult := &TallyResult{
		BallotID:   ballotID,
		TotalVotes: totalVotes,
		Results:    results,
		ComputedAt: time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	// Generate MPC signature for the results if MPC service is available
	if bs.mpc != nil && bs.mpc.IsReady() {
		resultHash := bs.computeResultHash(tallyResult)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		signature, err := bs.mpc.Sign(ctx, resultHash)
		if err != nil {
			bs.logger.Warnf("Failed to generate MPC signature for tally: %v", err)
		} else {
			tallyResult.MPCSignature = fmt.Sprintf("%x", signature)
		}
	}

	// Save tally result
	if err := bs.storage.SaveTallyResult(tallyResult); err != nil {
		return fmt.Errorf("failed to save tally result: %w", err)
	}

	bs.logger.Infof("Vote tally completed for ballot %s: %d total votes", ballotID, totalVotes)
	return nil
}

// computeResultHash computes a hash of the tally results for signing
func (bs *BallotService) computeResultHash(result *TallyResult) []byte {
	// Create a deterministic representation of the results
	data := fmt.Sprintf("ballot:%s,total:%d,timestamp:%d",
		result.BallotID,
		result.TotalVotes,
		result.ComputedAt.Unix())

	// Add sorted results
	for optionID, count := range result.Results {
		data += fmt.Sprintf(",option:%s:%d", optionID, count)
	}

	// Use the same digest function as MPC
	return mpc.Digest([]byte(data))
}

// loadDefaultTemplates loads predefined ballot templates
func (bs *BallotService) loadDefaultTemplates() error {
	// Yes/No template
	yesNoTemplate := &BallotTemplate{
		ID:          "yes_no",
		Name:        "Yes/No Vote",
		Description: "Simple yes or no question",
		Type:        TypeYesNo,
		Options: []BallotOption{
			{ID: "yes", Text: "Yes", Order: 1},
			{ID: "no", Text: "No", Order: 2},
		},
		Settings: BallotSettings{
			Duration:         24 * time.Hour,
			RequiresAuth:     true,
			AllowAnonymous:   false,
			MaxVotesPerVoter: 1,
		},
		CreatedAt: time.Now(),
	}

	// Multiple choice template
	multipleChoiceTemplate := &BallotTemplate{
		ID:          "multiple_choice",
		Name:        "Multiple Choice",
		Description: "Choose from multiple options",
		Type:        TypeMultipleChoice,
		Options:     []BallotOption{}, // Will be customized per ballot
		Settings: BallotSettings{
			Duration:         48 * time.Hour,
			RequiresAuth:     true,
			AllowAnonymous:   false,
			MaxVotesPerVoter: 1,
		},
		CreatedAt: time.Now(),
	}

	bs.templates["yes_no"] = yesNoTemplate
	bs.templates["multiple_choice"] = multipleChoiceTemplate

	return nil
}
