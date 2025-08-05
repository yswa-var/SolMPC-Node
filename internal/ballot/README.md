# Ballot Management System

## Overview

The Ballot Management System is a comprehensive voting infrastructure built for the SolMPC-Node project. It provides production-grade voting mechanics with MPC (Multi-Party Computation) integration for secure, verifiable voting on the Solana blockchain.

## Features

### ✅ Implemented Features

- **Complete Ballot Lifecycle Management**: Draft → Active → Closed → Archived
- **Time-based Ballot Control**: Automatic opening and closing of ballots
- **Voter Eligibility Management**: Whitelist/blacklist support with flexible access control
- **Multiple Voting Patterns**: Yes/No, Multiple Choice, Ranked Choice, and Custom ballots
- **Ballot Templates**: Pre-configured templates for common voting scenarios
- **MPC Integration**: Secure vote tallying with threshold signatures
- **File-based Storage**: Production-ready storage system with JSON persistence
- **Comprehensive Audit Trail**: Full logging of all voting actions
- **Thread-safe Operations**: Concurrent access protection with proper locking

### 🔄 Database Schema

Complete PostgreSQL schema with:
- Ballot management tables
- Vote storage with cryptographic verification
- Tally results with MPC signatures
- Audit logging for transparency
- Validator and MPC session tracking
- Performance-optimized indexes and triggers

## Architecture

```
internal/ballot/
├── models.go           # Core data structures (Ballot, Vote, TallyResult)
├── storage.go          # Storage interface and file-based implementation
├── service.go          # Main BallotService with business logic
├── scheduler.go        # Time-based ballot control
├── templates.go        # Ballot template management
├── integration.go      # MPC and logger integration wrappers
├── example.go          # Usage examples and demos
├── schema.sql          # Complete database schema
└── README.md           # This documentation
```

## Quick Start

### 1. Basic Usage

```go
package main

import (
    "time"
    "tilt-valid/internal/ballot"
    "tilt-valid/utils"
)

func main() {
    // Create logger
    logger := utils.Logger("ballot-demo", "voting")
    
    // Create storage
    storage, err := ballot.NewFileStorage("/tmp/ballot-data")
    if err != nil {
        panic(err)
    }
    
    // Create ballot service
    ballotService, err := ballot.NewBallotService(storage, nil, logger)
    if err != nil {
        panic(err)
    }
    
    // Start the service
    ballotService.Start()
    defer ballotService.Stop()
    
    // Create a ballot
    ballot := ballot.NewBallot(
        "Should we implement feature X?",
        "Community vote on new feature implementation",
        ballot.TypeYesNo,
        "creator_wallet_address",
    )
    
    ballot.AddOption("Yes", "Implement the feature")
    ballot.AddOption("No", "Do not implement")
    ballot.StartTime = time.Now()
    ballot.EndTime = time.Now().Add(24 * time.Hour)
    
    // Create and activate ballot
    ballotService.CreateBallot(ballot)
    ballotService.ActivateBallot(ballot.ID)
    
    // Cast votes
    vote := ballot.NewVote(ballot.ID, "voter_wallet", []ballot.Choice{
        {OptionID: ballot.Options[0].ID, Rank: 1},
    })
    ballotService.CastVote(vote)
    
    // Results will be automatically computed when ballot closes
}
```

### 2. Using Templates

```go
// Create template manager
templateManager := ballot.NewTemplateManager()

// Create ballot from template
ballot, err := templateManager.CreateBallotFromTemplate(
    "yes_no",
    "Protocol Upgrade Vote",
    "Should we upgrade to version 2.0?",
    "admin_wallet",
    time.Now().Add(1*time.Hour),
    time.Now().Add(25*time.Hour),
)

// Use with ballot service
ballotService.CreateBallot(ballot)
```

### 3. MPC Integration

```go
// Create MPC-integrated ballot service
mpcParty := // ... your existing MPC party
logger := utils.Logger("ballot", "mpc-voting")

integrationManager, err := ballot.NewIntegrationManager(
    mpcParty,
    logger,
    "/path/to/ballot/storage",
)

integrationManager.Start()
defer integrationManager.Stop()

ballotService := integrationManager.GetBallotService()
// Use ballotService normally - MPC signatures will be automatically generated
```

## Data Models

### Ballot

```go
type Ballot struct {
    ID               string                 `json:"id"`
    Title            string                 `json:"title"`
    Description      string                 `json:"description"`
    Type             BallotType             `json:"type"`
    Options          []BallotOption         `json:"options"`
    StartTime        time.Time              `json:"start_time"`
    EndTime          time.Time              `json:"end_time"`
    EligibleVoters   []string               `json:"eligible_voters"`
    Status           BallotStatus           `json:"status"`
    CreatedAt        time.Time              `json:"created_at"`
    UpdatedAt        time.Time              `json:"updated_at"`
    CreatedBy        string                 `json:"created_by"`
    RequiresAuth     bool                   `json:"requires_auth"`
    AllowAnonymous   bool                   `json:"allow_anonymous"`
    MaxVotesPerVoter int                    `json:"max_votes_per_voter"`
    Metadata         map[string]interface{} `json:"metadata,omitempty"`
}
```

### Vote

```go
type Vote struct {
    ID          string    `json:"id"`
    BallotID    string    `json:"ballot_id"`
    VoterID     string    `json:"voter_id"`
    Choices     []Choice  `json:"choices"`
    Timestamp   time.Time `json:"timestamp"`
    Signature   string    `json:"signature"`
    IsAnonymous bool      `json:"is_anonymous"`
}
```

### Tally Result

```go
type TallyResult struct {
    BallotID      string                 `json:"ballot_id"`
    TotalVotes    int                    `json:"total_votes"`
    Results       map[string]int         `json:"results"`
    Rankings      []RankingResult        `json:"rankings,omitempty"`
    ComputedAt    time.Time              `json:"computed_at"`
    MPCSignature  string                 `json:"mpc_signature"`
    ProofHash     string                 `json:"proof_hash"`
    Metadata      map[string]interface{} `json:"metadata,omitempty"`
}
```

## Ballot Types

### 1. Yes/No Voting
Simple binary choice voting.

```go
ballot := ballot.NewBallot("Question?", "Description", ballot.TypeYesNo, "creator")
ballot.AddOption("Yes", "Approve")
ballot.AddOption("No", "Reject")
```

### 2. Multiple Choice
Choose one option from many.

```go
ballot := ballot.NewBallot("Choose option", "Description", ballot.TypeMultipleChoice, "creator")
ballot.AddOption("Option A", "First choice")
ballot.AddOption("Option B", "Second choice")
ballot.AddOption("Option C", "Third choice")
ballot.MaxVotesPerVoter = 1
```

### 3. Ranked Choice
Rank options in order of preference.

```go
ballot := ballot.NewBallot("Rank candidates", "Description", ballot.TypeRanked, "creator")
ballot.AddOption("Candidate 1", "First candidate")
ballot.AddOption("Candidate 2", "Second candidate")
ballot.AddOption("Candidate 3", "Third candidate")
ballot.MaxVotesPerVoter = 3 // Can rank up to 3
```

### 4. Approval Voting
Select all options you approve of.

```go
ballot := ballot.NewBallot("Approve proposals", "Description", ballot.TypeMultipleChoice, "creator")
ballot.AddOption("Proposal A", "First proposal")
ballot.AddOption("Proposal B", "Second proposal")
ballot.AddOption("Proposal C", "Third proposal")
ballot.MaxVotesPerVoter = 3 // Can approve multiple
```

## Security Features

### 1. Voter Authentication
- Wallet-based authentication
- Signature verification
- Whitelist/blacklist management

### 2. Vote Integrity
- Cryptographic signatures on votes
- MPC-based result verification
- Immutable audit trail

### 3. Privacy Protection
- Anonymous voting support
- Zero-knowledge proof compatibility (future)
- Encrypted vote storage (future)

## Storage System

### File-based Storage (Current)
- JSON-based persistence
- Thread-safe operations
- Automatic directory structure creation
- Efficient querying and filtering

### Database Storage (Future)
- Complete PostgreSQL schema provided
- Optimized indexes for performance
- Automatic audit logging
- ACID compliance

## Integration Points

### MPC Service Integration
The ballot system integrates with the existing MPC infrastructure:

```go
type MPCService interface {
    Sign(ctx context.Context, msgHash []byte) ([]byte, error)
    GetPublicKey() ([]byte, error)
    IsReady() bool
}
```

### Logger Integration
Compatible with existing logging infrastructure:

```go
type Logger interface {
    Debugf(format string, a ...interface{})
    Warnf(format string, a ...interface{})
    Errorf(format string, a ...interface{})
    Infof(format string, a ...interface{})
}
```

## Configuration

### Environment Variables
```bash
BALLOT_STORAGE_PATH=/path/to/ballot/storage
BALLOT_AUTO_CLOSE_ENABLED=true
BALLOT_MPC_TIMEOUT=30s
```

### Service Configuration
```go
config := &BallotConfig{
    StoragePath:        "/var/lib/ballot",
    AutoCloseEnabled:   true,
    MPCTimeoutSeconds: 30,
    MaxBallotDuration: 30 * 24 * time.Hour, // 30 days
}
```

## API Reference

### BallotService Methods

#### Ballot Management
- `CreateBallot(ballot *Ballot) error`
- `GetBallot(id string) (*Ballot, error)`
- `UpdateBallot(ballot *Ballot) error`
- `ActivateBallot(id string) error`
- `CloseBallot(id string) error`
- `ArchiveBallot(id string) error`
- `ListBallots(status BallotStatus) ([]*Ballot, error)`

#### Voting Operations
- `CastVote(vote *Vote) error`
- `GetTallyResult(ballotID string) (*TallyResult, error)`

#### Template Management
- `CreateFromTemplate(templateID, title, description, createdBy string, startTime, endTime time.Time) (*Ballot, error)`

#### Service Control
- `Start() error`
- `Stop() error`

### Storage Interface

```go
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
```

## Testing

### Running Examples
```bash
cd internal/ballot
go run example.go
```

### Unit Tests (Future)
```bash
go test ./internal/ballot/...
```

## Roadmap

### Phase 2 Completed ✅
- [x] Core ballot management system
- [x] File-based storage implementation
- [x] Time-based ballot control
- [x] Voter eligibility management
- [x] Ballot templates
- [x] MPC integration interfaces
- [x] Complete database schema

### Phase 3 Planned 🔄
- [ ] Database storage implementation
- [ ] REST API endpoints
- [ ] WebSocket real-time updates
- [ ] Advanced cryptographic proofs
- [ ] Performance optimizations
- [ ] Comprehensive test suite

### Phase 4 Future 🔮
- [ ] Zero-knowledge privacy features
- [ ] Advanced analytics dashboard
- [ ] Multi-chain support
- [ ] Governance token integration
- [ ] Mobile app support

## Contributing

1. Follow existing code patterns
2. Add comprehensive tests for new features
3. Update documentation for API changes
4. Ensure thread safety for concurrent operations
5. Maintain backward compatibility

## License

This ballot management system is part of the SolMPC-Node project and follows the same licensing terms.