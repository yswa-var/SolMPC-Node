package distribution

import (
	"fmt"
	"sort"
	"strconv"
)

// DistributionData represents the node data with float64 for amounts
type DistributionData struct {
	Amount        float64   `json:"amount"`
	BusinessRules []float64 `json:"business_rules"`
	ID            int       `json:"id"`
	Receiver      []string  `json:"receiver"`
	Subtilt       []int     `json:"subtilt"`
}

// Allocation defines the output format
type Allocation struct {
	Receiver string
	Amount   float64
}

// AllocateAmounts distributes amounts across the hierarchy starting from rootID
func AllocateAmounts(tiltData map[string]map[string]interface{}, rootID string) ([]Allocation, error) {
	allocations := make(map[string]int) // Receiver ID -> total amount
	err := allocateRecursive(tiltData, rootID, 0, allocations)
	if err != nil {
		return nil, err
	}

	// Convert map to slice
	var result []Allocation
	for receiver, amount := range allocations {
		result = append(result, Allocation{Receiver: receiver, Amount: float64(amount)})
	}

	// Sort by receiver ID for consistent output (optional)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Receiver < result[j].Receiver
	})

	return result, nil
}

// allocateRecursive processes a node and its subtree
func allocateRecursive(tiltData map[string]map[string]interface{}, currentID string, receivedAmount int, allocations map[string]int) error {
	nodeData, ok := tiltData[currentID]
	if !ok {
		return fmt.Errorf("node ID %s not found in data", currentID)
	}

	// Extract node data
	ownAmount := nodeData["amount"].(int)
	businessRules := nodeData["business_rules"].([]int)
	receivers := nodeData["receivers"].([]string)
	subtilts := nodeData["subtilt"].([]int)

	// Validate business rules
	if len(businessRules) != len(subtilts)+1 {
		return fmt.Errorf("node %s: business rules length (%d) must be subtilts length (%d) + 1", currentID, len(businessRules), len(subtilts))
	}
	sum := 0
	for _, p := range businessRules {
		sum += p
	}
	if sum != 100 {
		return fmt.Errorf("node %s: business rules must sum to 100, got %d", currentID, sum)
	}

	// Compute total amount to distribute
	totalAmount := ownAmount + receivedAmount

	// Calculate amounts for local receivers and subtilts
	amounts := make([]int, len(businessRules))
	for i, percentage := range businessRules {
		amounts[i] = (percentage * totalAmount) / 100
	}

	// Adjust for remainder due to integer division
	totalAllocated := 0
	for _, amt := range amounts {
		totalAllocated += amt
	}
	remainder := totalAmount - totalAllocated
	for i := 0; remainder > 0; i++ {
		amounts[i%len(amounts)]++
		remainder--
	}

	// Distribute local amount to receivers
	localAmount := amounts[0]
	numReceivers := len(receivers)
	if numReceivers > 0 {
		amountPerReceiver := localAmount / numReceivers
		remainder := localAmount % numReceivers
		for i, receiver := range receivers {
			alloc := amountPerReceiver
			if i < remainder {
				alloc++
			}
			allocations[receiver] += alloc
		}
	}

	// Distribute to subtilts recursively
	for i, subtiltID := range subtilts {
		subAmount := amounts[i+1]
		subtiltKey := strconv.Itoa(subtiltID) // Convert int ID to string key
		if err := allocateRecursive(tiltData, subtiltKey, subAmount, allocations); err != nil {
			return err
		}
	}

	return nil
}
