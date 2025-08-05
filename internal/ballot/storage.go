package ballot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// BallotStorage defines the interface for ballot persistence
type BallotStorage interface {
	SaveBallot(ballot *Ballot) error
	GetBallot(id string) (*Ballot, error)
	ListBallots(status BallotStatus) ([]*Ballot, error)
	UpdateBallot(ballot *Ballot) error
	DeleteBallot(id string) error

	SaveVote(vote *Vote) error
	GetVote(id string) (*Vote, error)
	GetVotesByBallot(ballotID string) ([]*Vote, error)
	GetVotesByVoter(voterID string) ([]*Vote, error)

	SaveTallyResult(result *TallyResult) error
	GetTallyResult(ballotID string) (*TallyResult, error)
}

// FileStorage implements BallotStorage using the filesystem
type FileStorage struct {
	basePath string
	mutex    sync.RWMutex
}

// NewFileStorage creates a new file-based storage
func NewFileStorage(basePath string) (*FileStorage, error) {
	// Create directory structure
	dirs := []string{
		filepath.Join(basePath, "ballots"),
		filepath.Join(basePath, "votes"),
		filepath.Join(basePath, "results"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return &FileStorage{
		basePath: basePath,
	}, nil
}

// SaveBallot saves a ballot to the filesystem
func (fs *FileStorage) SaveBallot(ballot *Ballot) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	ballot.UpdatedAt = time.Now()

	data, err := ballot.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal ballot: %w", err)
	}

	filename := filepath.Join(fs.basePath, "ballots", fmt.Sprintf("%s.json", ballot.ID))
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write ballot file: %w", err)
	}

	return nil
}

// GetBallot retrieves a ballot by ID
func (fs *FileStorage) GetBallot(id string) (*Ballot, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	filename := filepath.Join(fs.basePath, "ballots", fmt.Sprintf("%s.json", id))
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("ballot not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read ballot file: %w", err)
	}

	return FromJSON(data)
}

// ListBallots returns all ballots with the specified status
func (fs *FileStorage) ListBallots(status BallotStatus) ([]*Ballot, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	ballotsDir := filepath.Join(fs.basePath, "ballots")
	files, err := ioutil.ReadDir(ballotsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read ballots directory: %w", err)
	}

	var ballots []*Ballot
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filename := filepath.Join(ballotsDir, file.Name())
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			continue // Skip corrupted files
		}

		ballot, err := FromJSON(data)
		if err != nil {
			continue // Skip corrupted files
		}

		if status == "" || ballot.Status == status {
			ballots = append(ballots, ballot)
		}
	}

	return ballots, nil
}

// UpdateBallot updates an existing ballot
func (fs *FileStorage) UpdateBallot(ballot *Ballot) error {
	// Check if ballot exists
	_, err := fs.GetBallot(ballot.ID)
	if err != nil {
		return fmt.Errorf("ballot not found for update: %w", err)
	}

	return fs.SaveBallot(ballot)
}

// DeleteBallot removes a ballot from storage
func (fs *FileStorage) DeleteBallot(id string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	filename := filepath.Join(fs.basePath, "ballots", fmt.Sprintf("%s.json", id))
	if err := os.Remove(filename); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("ballot not found: %s", id)
		}
		return fmt.Errorf("failed to delete ballot file: %w", err)
	}

	return nil
}

// SaveVote saves a vote to the filesystem
func (fs *FileStorage) SaveVote(vote *Vote) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	data, err := json.MarshalIndent(vote, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal vote: %w", err)
	}

	filename := filepath.Join(fs.basePath, "votes", fmt.Sprintf("%s.json", vote.ID))
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write vote file: %w", err)
	}

	return nil
}

// GetVote retrieves a vote by ID
func (fs *FileStorage) GetVote(id string) (*Vote, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	filename := filepath.Join(fs.basePath, "votes", fmt.Sprintf("%s.json", id))
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("vote not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read vote file: %w", err)
	}

	var vote Vote
	if err := json.Unmarshal(data, &vote); err != nil {
		return nil, fmt.Errorf("failed to unmarshal vote: %w", err)
	}

	return &vote, nil
}

// GetVotesByBallot retrieves all votes for a specific ballot
func (fs *FileStorage) GetVotesByBallot(ballotID string) ([]*Vote, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	votesDir := filepath.Join(fs.basePath, "votes")
	files, err := ioutil.ReadDir(votesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read votes directory: %w", err)
	}

	var votes []*Vote
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filename := filepath.Join(votesDir, file.Name())
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			continue // Skip corrupted files
		}

		var vote Vote
		if err := json.Unmarshal(data, &vote); err != nil {
			continue // Skip corrupted files
		}

		if vote.BallotID == ballotID {
			votes = append(votes, &vote)
		}
	}

	return votes, nil
}

// GetVotesByVoter retrieves all votes cast by a specific voter
func (fs *FileStorage) GetVotesByVoter(voterID string) ([]*Vote, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	votesDir := filepath.Join(fs.basePath, "votes")
	files, err := ioutil.ReadDir(votesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read votes directory: %w", err)
	}

	var votes []*Vote
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filename := filepath.Join(votesDir, file.Name())
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			continue // Skip corrupted files
		}

		var vote Vote
		if err := json.Unmarshal(data, &vote); err != nil {
			continue // Skip corrupted files
		}

		if vote.VoterID == voterID {
			votes = append(votes, &vote)
		}
	}

	return votes, nil
}

// SaveTallyResult saves tally results to the filesystem
func (fs *FileStorage) SaveTallyResult(result *TallyResult) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tally result: %w", err)
	}

	filename := filepath.Join(fs.basePath, "results", fmt.Sprintf("%s.json", result.BallotID))
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write tally result file: %w", err)
	}

	return nil
}

// GetTallyResult retrieves tally results for a ballot
func (fs *FileStorage) GetTallyResult(ballotID string) (*TallyResult, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	filename := filepath.Join(fs.basePath, "results", fmt.Sprintf("%s.json", ballotID))
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("tally result not found for ballot: %s", ballotID)
		}
		return nil, fmt.Errorf("failed to read tally result file: %w", err)
	}

	var result TallyResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tally result: %w", err)
	}

	return &result, nil
}

// GetStorageStats returns statistics about the storage
func (fs *FileStorage) GetStorageStats() (map[string]int, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	stats := make(map[string]int)

	// Count ballots
	ballotsDir := filepath.Join(fs.basePath, "ballots")
	if files, err := ioutil.ReadDir(ballotsDir); err == nil {
		stats["ballots"] = len(files)
	}

	// Count votes
	votesDir := filepath.Join(fs.basePath, "votes")
	if files, err := ioutil.ReadDir(votesDir); err == nil {
		stats["votes"] = len(files)
	}

	// Count results
	resultsDir := filepath.Join(fs.basePath, "results")
	if files, err := ioutil.ReadDir(resultsDir); err == nil {
		stats["results"] = len(files)
	}

	return stats, nil
}
