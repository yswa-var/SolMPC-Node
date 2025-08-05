package ballot

import (
	"fmt"
	"time"
)

// TemplateManager manages ballot templates
type TemplateManager struct {
	templates map[string]*BallotTemplate
}

// NewTemplateManager creates a new template manager
func NewTemplateManager() *TemplateManager {
	tm := &TemplateManager{
		templates: make(map[string]*BallotTemplate),
	}

	// Load default templates
	tm.loadDefaultTemplates()

	return tm
}

// GetTemplate retrieves a template by ID
func (tm *TemplateManager) GetTemplate(id string) (*BallotTemplate, error) {
	template, exists := tm.templates[id]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", id)
	}

	return template, nil
}

// ListTemplates returns all available templates
func (tm *TemplateManager) ListTemplates() []*BallotTemplate {
	templates := make([]*BallotTemplate, 0, len(tm.templates))
	for _, template := range tm.templates {
		templates = append(templates, template)
	}

	return templates
}

// AddTemplate adds a new template
func (tm *TemplateManager) AddTemplate(template *BallotTemplate) error {
	if template.ID == "" {
		return fmt.Errorf("template ID cannot be empty")
	}

	tm.templates[template.ID] = template
	return nil
}

// RemoveTemplate removes a template
func (tm *TemplateManager) RemoveTemplate(id string) error {
	if _, exists := tm.templates[id]; !exists {
		return fmt.Errorf("template not found: %s", id)
	}

	delete(tm.templates, id)
	return nil
}

// CreateBallotFromTemplate creates a ballot from a template
func (tm *TemplateManager) CreateBallotFromTemplate(templateID, title, description, createdBy string, startTime, endTime time.Time) (*Ballot, error) {
	template, err := tm.GetTemplate(templateID)
	if err != nil {
		return nil, err
	}

	ballot := NewBallot(title, description, template.Type, createdBy)
	ballot.Options = make([]BallotOption, len(template.Options))
	copy(ballot.Options, template.Options)
	ballot.StartTime = startTime
	ballot.EndTime = endTime
	ballot.RequiresAuth = template.Settings.RequiresAuth
	ballot.AllowAnonymous = template.Settings.AllowAnonymous
	ballot.MaxVotesPerVoter = template.Settings.MaxVotesPerVoter

	return ballot, nil
}

// loadDefaultTemplates loads predefined ballot templates
func (tm *TemplateManager) loadDefaultTemplates() {
	// Yes/No template
	yesNoTemplate := &BallotTemplate{
		ID:          "yes_no",
		Name:        "Yes/No Vote",
		Description: "Simple yes or no question",
		Type:        TypeYesNo,
		Options: []BallotOption{
			{ID: "yes", Text: "Yes", Description: "Vote in favor", Order: 1},
			{ID: "no", Text: "No", Description: "Vote against", Order: 2},
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
		Description: "Choose one option from multiple choices",
		Type:        TypeMultipleChoice,
		Options: []BallotOption{
			{ID: "option_a", Text: "Option A", Order: 1},
			{ID: "option_b", Text: "Option B", Order: 2},
			{ID: "option_c", Text: "Option C", Order: 3},
		},
		Settings: BallotSettings{
			Duration:         48 * time.Hour,
			RequiresAuth:     true,
			AllowAnonymous:   false,
			MaxVotesPerVoter: 1,
		},
		CreatedAt: time.Now(),
	}

	// Ranked choice template
	rankedTemplate := &BallotTemplate{
		ID:          "ranked_choice",
		Name:        "Ranked Choice Voting",
		Description: "Rank options in order of preference",
		Type:        TypeRanked,
		Options: []BallotOption{
			{ID: "candidate_1", Text: "Candidate 1", Order: 1},
			{ID: "candidate_2", Text: "Candidate 2", Order: 2},
			{ID: "candidate_3", Text: "Candidate 3", Order: 3},
		},
		Settings: BallotSettings{
			Duration:         72 * time.Hour,
			RequiresAuth:     true,
			AllowAnonymous:   false,
			MaxVotesPerVoter: 3, // Can rank up to 3 choices
		},
		CreatedAt: time.Now(),
	}

	// Approval voting template (multiple selections allowed)
	approvalTemplate := &BallotTemplate{
		ID:          "approval_voting",
		Name:        "Approval Voting",
		Description: "Select all options you approve of",
		Type:        TypeMultipleChoice,
		Options: []BallotOption{
			{ID: "proposal_a", Text: "Proposal A", Order: 1},
			{ID: "proposal_b", Text: "Proposal B", Order: 2},
			{ID: "proposal_c", Text: "Proposal C", Order: 3},
			{ID: "proposal_d", Text: "Proposal D", Order: 4},
		},
		Settings: BallotSettings{
			Duration:         48 * time.Hour,
			RequiresAuth:     true,
			AllowAnonymous:   false,
			MaxVotesPerVoter: 4, // Can approve multiple options
		},
		CreatedAt: time.Now(),
	}

	// Anonymous voting template
	anonymousTemplate := &BallotTemplate{
		ID:          "anonymous_poll",
		Name:        "Anonymous Poll",
		Description: "Anonymous voting with privacy protection",
		Type:        TypeYesNo,
		Options: []BallotOption{
			{ID: "yes", Text: "Yes", Order: 1},
			{ID: "no", Text: "No", Order: 2},
		},
		Settings: BallotSettings{
			Duration:         24 * time.Hour,
			RequiresAuth:     false,
			AllowAnonymous:   true,
			MaxVotesPerVoter: 1,
		},
		CreatedAt: time.Now(),
	}

	// Store templates
	tm.templates["yes_no"] = yesNoTemplate
	tm.templates["multiple_choice"] = multipleChoiceTemplate
	tm.templates["ranked_choice"] = rankedTemplate
	tm.templates["approval_voting"] = approvalTemplate
	tm.templates["anonymous_poll"] = anonymousTemplate
}
