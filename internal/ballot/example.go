package ballot

import (
	"fmt"
	"time"

	"tilt-valid/utils"
)

// ExampleUsage demonstrates how to use the ballot management system
func ExampleUsage() {
	fmt.Println("=== Ballot Management System Example ===")

	// 1. Create logger (using existing utils.Logger function)
	logger := utils.Logger("ballot-example", "ballot-demo")

	// 2. Create file storage
	storage, err := NewFileStorage("/tmp/ballot-storage")
	if err != nil {
		logger.Errorf("Failed to create storage: %v", err)
		return
	}

	// 3. Create MPC service (nil for demo - would use actual MPC party in production)
	var mpcService MPCService = nil

	// 4. Create ballot service
	ballotService, err := NewBallotService(storage, mpcService, logger)
	if err != nil {
		logger.Errorf("Failed to create ballot service: %v", err)
		return
	}

	// 5. Start the service
	if err := ballotService.Start(); err != nil {
		logger.Errorf("Failed to start ballot service: %v", err)
		return
	}
	defer ballotService.Stop()

	// 6. Create a simple Yes/No ballot
	ballot := NewBallot(
		"Should we implement feature X?",
		"This ballot asks the community whether we should implement the new feature X",
		TypeYesNo,
		"creator_wallet_address",
	)

	// Add options
	ballot.AddOption("Yes", "Implement the feature")
	ballot.AddOption("No", "Do not implement the feature")

	// Set timing (start now, end in 1 hour)
	ballot.StartTime = time.Now()
	ballot.EndTime = time.Now().Add(1 * time.Hour)

	// Add eligible voters
	ballot.EligibleVoters = []string{
		"voter1_wallet_address",
		"voter2_wallet_address",
		"voter3_wallet_address",
	}

	// 7. Create the ballot
	if err := ballotService.CreateBallot(ballot); err != nil {
		logger.Errorf("Failed to create ballot: %v", err)
		return
	}

	logger.Infof("Created ballot: %s", ballot.ID)

	// 8. Activate the ballot
	if err := ballotService.ActivateBallot(ballot.ID); err != nil {
		logger.Errorf("Failed to activate ballot: %v", err)
		return
	}

	logger.Infof("Activated ballot: %s", ballot.ID)

	// 9. Cast some votes
	votes := []*Vote{
		NewVote(ballot.ID, "voter1_wallet_address", []Choice{
			{OptionID: ballot.Options[0].ID, Rank: 1}, // Yes
		}),
		NewVote(ballot.ID, "voter2_wallet_address", []Choice{
			{OptionID: ballot.Options[1].ID, Rank: 1}, // No
		}),
		NewVote(ballot.ID, "voter3_wallet_address", []Choice{
			{OptionID: ballot.Options[0].ID, Rank: 1}, // Yes
		}),
	}

	for _, vote := range votes {
		if err := ballotService.CastVote(vote); err != nil {
			logger.Errorf("Failed to cast vote: %v", err)
		} else {
			logger.Infof("Vote cast by %s", vote.VoterID)
		}
	}

	// 10. Close the ballot manually (normally would happen automatically)
	if err := ballotService.CloseBallot(ballot.ID); err != nil {
		logger.Errorf("Failed to close ballot: %v", err)
		return
	}

	logger.Infof("Closed ballot: %s", ballot.ID)

	// 11. Wait a moment for tally to complete, then get results
	time.Sleep(1 * time.Second)

	result, err := ballotService.GetTallyResult(ballot.ID)
	if err != nil {
		logger.Errorf("Failed to get tally result: %v", err)
		return
	}

	logger.Infof("Tally Results for ballot %s:", ballot.ID)
	logger.Infof("Total votes: %d", result.TotalVotes)
	for optionID, count := range result.Results {
		// Find option text
		var optionText string
		for _, option := range ballot.Options {
			if option.ID == optionID {
				optionText = option.Text
				break
			}
		}
		logger.Infof("%s: %d votes", optionText, count)
	}

	// 12. List all ballots
	allBallots, err := ballotService.ListBallots("")
	if err != nil {
		logger.Errorf("Failed to list ballots: %v", err)
		return
	}

	logger.Infof("Total ballots in system: %d", len(allBallots))

	fmt.Println("=== Example completed successfully ===")
}

// CreateTemplateExample demonstrates ballot template usage
func CreateTemplateExample() {
	fmt.Println("=== Ballot Template Example ===")

	// Create template manager
	templateManager := NewTemplateManager()

	// List available templates
	templates := templateManager.ListTemplates()
	fmt.Printf("Available templates: %d\n", len(templates))

	for _, template := range templates {
		fmt.Printf("- %s: %s (%s)\n", template.ID, template.Name, template.Type)
	}

	// Create ballot from template
	ballot, err := templateManager.CreateBallotFromTemplate(
		"yes_no",
		"Community Vote: Should we upgrade the protocol?",
		"This vote determines whether we should proceed with the protocol upgrade",
		"admin_wallet",
		time.Now().Add(1*time.Hour),  // Start in 1 hour
		time.Now().Add(25*time.Hour), // End in 25 hours
	)

	if err != nil {
		fmt.Printf("Failed to create ballot from template: %v\n", err)
		return
	}

	fmt.Printf("Created ballot from template: %s\n", ballot.ID)
	fmt.Printf("Ballot type: %s\n", ballot.Type)
	fmt.Printf("Options: %d\n", len(ballot.Options))

	for _, option := range ballot.Options {
		fmt.Printf("- %s: %s\n", option.ID, option.Text)
	}

	fmt.Println("=== Template example completed ===")
}
