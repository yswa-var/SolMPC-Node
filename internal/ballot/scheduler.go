package ballot

import (
	"sync"
	"time"

	mpc "tilt-valid/internal/mpc"
)

// BallotScheduler handles time-based ballot operations
type BallotScheduler struct {
	service  *BallotService
	logger   mpc.Logger
	timers   map[string]*time.Timer
	mutex    sync.RWMutex
	stopChan chan struct{}
	running  bool
}

// NewBallotScheduler creates a new ballot scheduler
func NewBallotScheduler(service *BallotService, logger mpc.Logger) *BallotScheduler {
	return &BallotScheduler{
		service:  service,
		logger:   logger,
		timers:   make(map[string]*time.Timer),
		stopChan: make(chan struct{}),
	}
}

// Start starts the scheduler
func (bs *BallotScheduler) Start() {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	if bs.running {
		return
	}

	bs.running = true
	bs.logger.Infof("Ballot scheduler started")
}

// Stop stops the scheduler and cancels all timers
func (bs *BallotScheduler) Stop() {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	if !bs.running {
		return
	}

	bs.running = false

	// Cancel all timers
	for ballotID, timer := range bs.timers {
		timer.Stop()
		delete(bs.timers, ballotID)
	}

	close(bs.stopChan)
	bs.logger.Infof("Ballot scheduler stopped")
}

// ScheduleBallotClosure schedules automatic closure of a ballot
func (bs *BallotScheduler) ScheduleBallotClosure(ballotID string, endTime time.Time) {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	if !bs.running {
		return
	}

	// Cancel existing timer if any
	if timer, exists := bs.timers[ballotID]; exists {
		timer.Stop()
		delete(bs.timers, ballotID)
	}

	// Calculate duration until end time
	duration := time.Until(endTime)
	if duration <= 0 {
		// Ballot should be closed immediately
		go func() {
			if err := bs.service.CloseBallot(ballotID); err != nil {
				bs.logger.Errorf("Failed to auto-close ballot %s: %v", ballotID, err)
			}
		}()
		return
	}

	// Create new timer
	timer := time.AfterFunc(duration, func() {
		bs.mutex.Lock()
		delete(bs.timers, ballotID)
		bs.mutex.Unlock()

		bs.logger.Infof("Auto-closing ballot: %s", ballotID)
		if err := bs.service.CloseBallot(ballotID); err != nil {
			bs.logger.Errorf("Failed to auto-close ballot %s: %v", ballotID, err)
		}
	})

	bs.timers[ballotID] = timer
	bs.logger.Infof("Scheduled ballot closure for %s at %s", ballotID, endTime.Format(time.RFC3339))
}

// CancelBallotClosure cancels a scheduled ballot closure
func (bs *BallotScheduler) CancelBallotClosure(ballotID string) {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	if timer, exists := bs.timers[ballotID]; exists {
		timer.Stop()
		delete(bs.timers, ballotID)
		bs.logger.Infof("Cancelled scheduled closure for ballot: %s", ballotID)
	}
}

// GetScheduledBallots returns a list of ballots with scheduled closures
func (bs *BallotScheduler) GetScheduledBallots() []string {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()

	var ballotIDs []string
	for ballotID := range bs.timers {
		ballotIDs = append(ballotIDs, ballotID)
	}

	return ballotIDs
}

// IsRunning returns whether the scheduler is currently running
func (bs *BallotScheduler) IsRunning() bool {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()

	return bs.running
}
