package ballot

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// BallotStatus represents the current state of a ballot
type BallotStatus string

const (
	StatusDraft    BallotStatus = "draft"
	StatusActive   BallotStatus = "active"
	StatusClosed   BallotStatus = "closed"
	StatusArchived BallotStatus = "archived"
)

// BallotType represents different voting patterns
type BallotType string

const (
	TypeYesNo          BallotType = "yes_no"
	TypeMultipleChoice BallotType = "multiple_choice"
	TypeRanked         BallotType = "ranked"
	TypeCustom         BallotType = "custom"
)

// Ballot represents a voting ballot with comprehensive metadata
type Ballot struct {
	ID               string                 `json:"id"`
	Title            string                 `json:"title"`
	Description      string                 `json:"description"`
	Type             BallotType             `json:"type"`
	Options          []BallotOption         `json:"options"`
	StartTime        time.Time              `json:"start_time"`
	EndTime          time.Time              `json:"end_time"`
	EligibleVoters   []string               `json:"eligible_voters"` // Wallet addresses or voter IDs
	Status           BallotStatus           `json:"status"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	CreatedBy        string                 `json:"created_by"`          // Creator's wallet address
	RequiresAuth     bool                   `json:"requires_auth"`       // Whether voters need authentication
	AllowAnonymous   bool                   `json:"allow_anonymous"`     // Whether to allow anonymous voting
	MaxVotesPerVoter int                    `json:"max_votes_per_voter"` // For multiple choice ballots
	Metadata         map[string]interface{} `json:"metadata,omitempty"`  // Additional custom data
}

// BallotOption represents a single voting option
type BallotOption struct {
	ID          string `json:"id"`
	Text        string `json:"text"`
	Description string `json:"description,omitempty"`
	Order       int    `json:"order"` // For ranked voting
}

// Vote represents a single vote cast by a voter
type Vote struct {
	ID          string    `json:"id"`
	BallotID    string    `json:"ballot_id"`
	VoterID     string    `json:"voter_id"` // Wallet address or anonymous ID
	Choices     []Choice  `json:"choices"`  // Support multiple choices/rankings
	Timestamp   time.Time `json:"timestamp"`
	Signature   string    `json:"signature"` // Cryptographic signature for verification
	IsAnonymous bool      `json:"is_anonymous"`
}

// Choice represents a single choice within a vote
type Choice struct {
	OptionID string `json:"option_id"`
	Rank     int    `json:"rank,omitempty"`   // For ranked voting (1 = highest preference)
	Weight   int    `json:"weight,omitempty"` // For weighted voting
}

// VoterEligibility manages voter access control
type VoterEligibility struct {
	Whitelist    []string `json:"whitelist,omitempty"`     // Allowed voter addresses
	Blacklist    []string `json:"blacklist,omitempty"`     // Blocked voter addresses
	MinBalance   uint64   `json:"min_balance,omitempty"`   // Minimum token balance required
	TokenAddress string   `json:"token_address,omitempty"` // Token contract for balance check
}

// TallyResult represents the final results of a ballot
type TallyResult struct {
	BallotID     string                 `json:"ballot_id"`
	TotalVotes   int                    `json:"total_votes"`
	Results      map[string]int         `json:"results"`            // OptionID -> vote count
	Rankings     []RankingResult        `json:"rankings,omitempty"` // For ranked voting
	ComputedAt   time.Time              `json:"computed_at"`
	MPCSignature string                 `json:"mpc_signature"` // MPC signature for result verification
	ProofHash    string                 `json:"proof_hash"`    // Hash of cryptographic proof
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// RankingResult represents ranking results for ranked choice voting
type RankingResult struct {
	OptionID    string  `json:"option_id"`
	Points      int     `json:"points"`       // Total ranking points
	FirstChoice int     `json:"first_choice"` // Number of first-place votes
	AverageRank float64 `json:"average_rank"` // Average ranking position
}

// BallotTemplate represents a reusable ballot configuration
type BallotTemplate struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Type        BallotType     `json:"type"`
	Options     []BallotOption `json:"options"`
	Settings    BallotSettings `json:"settings"`
	CreatedAt   time.Time      `json:"created_at"`
}

// BallotSettings contains configurable ballot parameters
type BallotSettings struct {
	Duration         time.Duration    `json:"duration"` // Default ballot duration
	RequiresAuth     bool             `json:"requires_auth"`
	AllowAnonymous   bool             `json:"allow_anonymous"`
	MaxVotesPerVoter int              `json:"max_votes_per_voter"`
	Eligibility      VoterEligibility `json:"eligibility"`
}

// NewBallot creates a new ballot with default values
func NewBallot(title, description string, ballotType BallotType, createdBy string) *Ballot {
	return &Ballot{
		ID:               uuid.New().String(),
		Title:            title,
		Description:      description,
		Type:             ballotType,
		Options:          make([]BallotOption, 0),
		EligibleVoters:   make([]string, 0),
		Status:           StatusDraft,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		CreatedBy:        createdBy,
		RequiresAuth:     true,
		AllowAnonymous:   false,
		MaxVotesPerVoter: 1,
		Metadata:         make(map[string]interface{}),
	}
}

// NewVote creates a new vote
func NewVote(ballotID, voterID string, choices []Choice) *Vote {
	return &Vote{
		ID:          uuid.New().String(),
		BallotID:    ballotID,
		VoterID:     voterID,
		Choices:     choices,
		Timestamp:   time.Now(),
		IsAnonymous: false,
	}
}

// IsActive checks if the ballot is currently active for voting
func (b *Ballot) IsActive() bool {
	now := time.Now()
	return b.Status == StatusActive &&
		now.After(b.StartTime) &&
		now.Before(b.EndTime)
}

// CanVote checks if a voter is eligible to vote on this ballot
func (b *Ballot) CanVote(voterID string) bool {
	if !b.IsActive() {
		return false
	}

	// Check whitelist if it exists
	if len(b.EligibleVoters) > 0 {
		for _, eligible := range b.EligibleVoters {
			if eligible == voterID {
				return true
			}
		}
		return false
	}

	// If no whitelist, allow all voters (unless there's a blacklist in the future)
	return true
}

// AddOption adds a new option to the ballot
func (b *Ballot) AddOption(text, description string) {
	option := BallotOption{
		ID:          uuid.New().String(),
		Text:        text,
		Description: description,
		Order:       len(b.Options) + 1,
	}
	b.Options = append(b.Options, option)
	b.UpdatedAt = time.Now()
}

// ToJSON converts the ballot to JSON
func (b *Ballot) ToJSON() ([]byte, error) {
	return json.MarshalIndent(b, "", "  ")
}

// FromJSON creates a ballot from JSON data
func FromJSON(data []byte) (*Ballot, error) {
	var ballot Ballot
	err := json.Unmarshal(data, &ballot)
	return &ballot, err
}

// Validate checks if the ballot configuration is valid
func (b *Ballot) Validate() error {
	if b.Title == "" {
		return fmt.Errorf("ballot title is required")
	}

	if len(b.Options) < 2 && b.Type != TypeCustom {
		return fmt.Errorf("ballot must have at least 2 options")
	}

	if b.EndTime.Before(b.StartTime) {
		return fmt.Errorf("end time must be after start time")
	}

	if b.MaxVotesPerVoter < 1 {
		return fmt.Errorf("max votes per voter must be at least 1")
	}

	return nil
}
