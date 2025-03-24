package utils

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func UpdateTiltCounter(value int) error {
	filePath := "/Users/apple/Documents/GitHub/tilt-validator-main/utils/create-tilt-flag.txt"
	return os.WriteFile(filePath, []byte(strconv.Itoa(value)), 0644)
}

func ReadTiltCounter() (int, error) {
	filePath := "/Users/apple/Documents/GitHub/tilt-validator-main/utils/create-tilt-flag.txt"
	data, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %w", err)
	}
	value, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, fmt.Errorf("failed to convert value: %w", err)
	}
	return value, nil
}

func CreateTilt(id int, sender string, receiver []string, businessRules []int, subtilt []int, amount int) map[string]interface{} {
	tilt := map[string]interface{}{
		"id":             id,
		"sender":         sender,
		"receiver":       receiver,
		"business_rules": businessRules,
		"subtilt":        subtilt,
		"amount":         amount,
	}
	AppendTiltData(id, sender, receiver, businessRules, subtilt, amount)
	return tilt
}

func AppendTiltData(id int, sender string, receivers []string, businessRules []int, subtilt []int, amount int) error {
	filePath := "/Users/apple/Documents/GitHub/tilt-validator-main/utils/tiltdb.csv"
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Convert businessRules and subtilt to JSON strings
	businessRulesStr, err := json.Marshal(businessRules)
	if err != nil {
		return fmt.Errorf("failed to marshal business rules: %w", err)
	}
	subtiltStr, err := json.Marshal(subtilt)
	if err != nil {
		return fmt.Errorf("failed to marshal subtilt: %w", err)
	}

	record := []string{
		strconv.Itoa(id),
		sender,
		strings.Join(receivers, ";"),
		string(businessRulesStr),
		string(subtiltStr),
		strconv.Itoa(amount),
	}

	if err := writer.Write(record); err != nil {
		return fmt.Errorf("failed to write record to file: %w", err)
	}

	return nil
}

func ReadTiltDataByID(id int) (map[string]interface{}, error) {
	filePath := "/Users/apple/Documents/GitHub/tilt-validator-main/utils/tiltdb.csv"
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read records: %w", err)
	}

	for _, record := range records {
		recordID, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, fmt.Errorf("failed to convert id: %w", err)
		}
		if recordID == id {
			sender := record[1]
			receivers := strings.Split(record[2], ";")

			var businessRules []int
			if err := json.Unmarshal([]byte(record[3]), &businessRules); err != nil {
				return nil, fmt.Errorf("failed to parse business rules: %w", err)
			}

			var subtilt []int
			if err := json.Unmarshal([]byte(record[4]), &subtilt); err != nil {
				return nil, fmt.Errorf("failed to parse subtilt: %w", err)
			}

			amount, err := strconv.Atoi(record[5])
			if err != nil {
				return nil, fmt.Errorf("failed to convert amount: %w", err)
			}

			tilt := map[string]interface{}{
				"id":             recordID,
				"sender":         sender,
				"receiver":       receivers,
				"business_rules": businessRules,
				"subtilt":        subtilt,
				"amount":         amount,
			}
			return tilt, nil
		}
	}
	return nil, fmt.Errorf("tilt with id %d not found", id)
}

func CreateRandomRecievers() []string {
	numReceivers := 2
	var receivers []string
	for i := 0; i < numReceivers; i++ {
		receiverUUID := uuid.New()
		hash := sha256.Sum256([]byte(receiverUUID.String()))
		receivers = append(receivers, fmt.Sprintf("%x", hash))
	}
	return receivers
}

// GetTestTilt returns a tilt structure based on the specified type
func GetTestTilt(tiltType string, sender string) map[string]interface{} {

	filePath := "/Users/apple/Documents/GitHub/tilt-validator-main/utils/tiltdb.csv"
	file, err := os.OpenFile(filePath, os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return nil
	}
	file.Close()

	switch tiltType {
	case "simple":
		// A simple tilt with one recipient
		return CreateTilt(1, sender, CreateRandomRecievers(), []int{100}, nil, 100)
	case "one_subtilt":
		// Tilt with one sub-tilt
		CreateTilt(2, sender, CreateRandomRecievers(), []int{100}, nil, 100)
		return CreateTilt(1, sender, CreateRandomRecievers(), []int{80, 20}, []int{2}, 100)
	case "two_subtilts":
		// Tilt with two sub-tilts (matches original behavior)
		CreateTilt(3, sender, CreateRandomRecievers(), []int{100}, nil, 100)
		CreateTilt(2, sender, CreateRandomRecievers(), []int{100}, nil, 100)
		return CreateTilt(1, sender, CreateRandomRecievers(), []int{20, 70, 10}, []int{3, 2}, 100)
	case "nested":
		// A nested tilt structure with multiple levels
		CreateTilt(3, sender, CreateRandomRecievers(), []int{100}, nil, 100)
		CreateTilt(2, sender, CreateRandomRecievers(), []int{80, 20}, []int{3}, 100)
		return CreateTilt(1, sender, CreateRandomRecievers(), []int{80, 20}, []int{2}, 100)
	default:
		return nil
	}
}
