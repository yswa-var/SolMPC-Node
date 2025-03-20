package distribution

import (
	"fmt"
	"tilt-valid/utils"
)

// DistributionData represents the node data with float64 for amounts
type DistributionData struct {
	Amount        float64   `json:"amount"`
	BusinessRules []float64 `json:"business_rules"`
	ID            int       `json:"id"`
	Receiver      []string  `json:"receiver"`
	Sender        string    `json:"sender"`
	Subtilt       []int     `json:"subtilt"`
}

// Allocation defines the output format
type Allocation struct {
	Receiver string
	Amount   float64
}

// Distribution parses input into DistributionData
func Distribution(distBytes map[string]any) (DistributionData, error) {
	receivers, ok := distBytes["receiver"].([]string)
	if !ok {
		return DistributionData{}, fmt.Errorf("invalid receiver: expected []string")
	}
	senderStr, ok := distBytes["sender"].(string)
	if !ok {
		return DistributionData{}, fmt.Errorf("invalid sender: expected string")
	}
	amount, ok := distBytes["amount"].(float64)
	if !ok {
		// Try int conversion as a fallback
		if amtInt, ok := distBytes["amount"].(int); ok {
			amount = float64(amtInt)
		} else {
			return DistributionData{}, fmt.Errorf("invalid amount: expected float64")
		}
	}
	businessRules, ok := distBytes["business_rules"].([]float64)
	if !ok {
		// Try []int conversion as a fallback
		if brInt, ok := distBytes["business_rules"].([]int); ok {
			businessRules = make([]float64, len(brInt))
			for i, v := range brInt {
				businessRules[i] = float64(v)
			}
		} else {
			return DistributionData{}, fmt.Errorf("invalid business_rules: expected []float64")
		}
	}
	id, ok := distBytes["id"].(int)
	if !ok {
		return DistributionData{}, fmt.Errorf("invalid id: expected int")
	}
	subtilt, ok := distBytes["subtilt"].([]int)
	if !ok {
		return DistributionData{}, fmt.Errorf("invalid subtilt: expected []int")
	}

	return DistributionData{
		Amount:        amount,
		BusinessRules: businessRules,
		ID:            id,
		Receiver:      receivers,
		Sender:        senderStr,
		Subtilt:       subtilt,
	}, nil
}

// AllocateAmounts distributes amounts across the hierarchy
func AllocateAmounts(data DistributionData) ([]Allocation, error) {
	// Handle empty BusinessRules for leaf nodes
	if len(data.BusinessRules) == 0 {
		if len(data.Subtilt) == 0 && len(data.Receiver) > 0 {
			data.BusinessRules = []float64{100.0} // Default for leaf node with receivers
		} else {
			return nil, fmt.Errorf("invalid business rules: empty rules not allowed for node ID %d with %d subtilts", data.ID, len(data.Subtilt))
		}
	}

	// // Validate business rules length
	// if len(data.BusinessRules) != len(data.Subtilt)+1 {
	//     return nil, fmt.Errorf("invalid business rules: length mismatch for node ID %d, expected %d, got %d", data.ID, len(data.Subtilt)+1, len(data.BusinessRules))
	// }

	// Validate total percentage
	totalPercentage := 0.0
	for _, percentage := range data.BusinessRules {
		if percentage < 0 {
			return nil, fmt.Errorf("invalid business rules: negative percentage %f for node ID %d", percentage, data.ID)
		}
		totalPercentage += percentage
	}
	if totalPercentage != 100.0 {
		return nil, fmt.Errorf("invalid business rules: total percentage must be 100 for node ID %d, got %f", data.ID, totalPercentage)
	}

	// Validate receiver allocation
	if len(data.Receiver) == 0 && data.BusinessRules[0] != 0 {
		return nil, fmt.Errorf("invalid business rules: cannot allocate %f%% to non-existent receivers for node ID %d", data.BusinessRules[0], data.ID)
	}

	// Use a map to accumulate amounts
	resultMap := make(map[string]float64)

	// Allocate to receivers
	receiverAmount := data.Amount * (data.BusinessRules[0] / 100.0)
	if len(data.Receiver) > 0 {
		share := receiverAmount / float64(len(data.Receiver))
		for _, receiver := range data.Receiver {
			resultMap[receiver] += share
		}
	}

	// Allocate to sub-tilts
	for i, subID := range data.Subtilt {
		subAmount := data.Amount * (data.BusinessRules[i+1] / 100.0)
		subData, err := utils.ReadTiltDataByID(subID)
		if err != nil {
			return nil, fmt.Errorf("failed to read sub-tilt data for ID %d: %w", subID, err)
		}

		// Construct subDistData with type safety
		subDistData, err := Distribution(subData)
		if err != nil {
			return nil, fmt.Errorf("invalid sub-tilt data for ID %d: %w", subID, err)
		}
		subDistData.Amount += subAmount // Add parent's contribution

		// Recursive call
		subResult, err := AllocateAmounts(subDistData)
		if err != nil {
			return nil, fmt.Errorf("failed to allocate amounts for sub-tilt ID %d: %w", subID, err)
		}
		for _, alloc := range subResult {
			resultMap[alloc.Receiver] += alloc.Amount
		}
	}

	// Convert map to list
	var allocations []Allocation
	for receiver, amount := range resultMap {
		allocations = append(allocations, Allocation{Receiver: receiver, Amount: amount})
	}

	return allocations, nil
}
