package utils

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gagliardetto/solana-go"
)

func UpdateTiltCounter(value int) error {
	filePath := "/Users/yash/Downloads/exercises/tilt-validator/utils/create-tilt-flag.txt"
	return os.WriteFile(filePath, []byte(strconv.Itoa(value)), 0644)
}

func ReadTiltCounter() (int, error) {
	filePath := "/Users/yash/Downloads/exercises/tilt-validator/utils/create-tilt-flag.txt"
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

func CreateTilt(filepath string, id int, receiver []string, businessRules []int, subtilt []int, amount int) (map[string]any, error) {
	tilt := map[string]any{
		"id":             id,
		"receiver":       receiver,
		"business_rules": businessRules,
		"subtilt":        subtilt,
		"amount":         amount,
	}
	file, err := os.OpenFile("/Users/apple/Documents/GitHub/tv-solana_int/utils/tiltdb.csv", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write to file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Convert businessRules and subtilt to JSON strings
	businessRulesStr, err := json.Marshal(businessRules)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	subtiltStr, err := json.Marshal(subtilt)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	record := []string{
		strconv.Itoa(id),
		strings.Join(receiver, ";"),
		string(businessRulesStr),
		string(subtiltStr),
		strconv.Itoa(amount),
	}

	if err := writer.Write(record); err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return tilt, nil
}

func AppendTiltData(filePath string, id int, receivers []string, businessRules []int, subtilt []int, amount int) error {

	return nil
}

func ReadTiltData(filePath string) (map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV data: %w", err)
	}

	data := make(map[string]interface{})
	for _, record := range records {
		if len(record) < 5 {
			continue
		}

		id := record[0]
		receivers := strings.Split(record[1], ";")
		var businessRules []int
		if err := json.Unmarshal([]byte(record[2]), &businessRules); err != nil {
			return nil, fmt.Errorf("failed to parse business rules: %w", err)
		}
		var subtilt []int
		if record[3] != "null" {
			if err := json.Unmarshal([]byte(record[3]), &subtilt); err != nil {
				return nil, fmt.Errorf("failed to parse subtilt: %w", err)
			}
		}
		amount, err := strconv.Atoi(record[4])
		if err != nil {
			return nil, fmt.Errorf("failed to parse amount: %w", err)
		}

		data[id] = map[string]interface{}{
			"receivers":      receivers,
			"business_rules": businessRules,
			"subtilt":        subtilt,
			"amount":         amount,
		}
	}

	return data, nil
}

func CreateRandomRecievers() []string {
	numReceivers := 2
	receivers := make([]string, numReceivers)
	for i := range receivers {
		// Generate a new Solana wallet (keypair)
		wallet := solana.NewWallet()
		// Get the public key and encode it as Base58 string
		receivers[i] = wallet.PublicKey().String() // Base58-encoded string
	}
	return receivers
}

func DeleteTiltDBFile(filepath string) error {
	if err := os.Remove(filepath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func EnsureFileExists(filepath string) error {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		file, err := os.Create(filepath)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		file.Close()
	}
	return nil
}

// GetTestTilt returns a tilt structure based on the specified type
func GetTestTilt(filepath string, tiltType string) map[string]interface{} {
	DeleteTiltDBFile(filepath)
	EnsureFileExists(filepath)
	fmt.Println("Creating test tilt")
	file, err := os.OpenFile(filepath, os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return nil
	}
	file.Close()

	switch tiltType {
	case "simple":
		// A simple tilt with one recipient
		tilt, err := CreateTilt(filepath, 1, CreateRandomRecievers(), []int{100}, nil, 100)
		if err != nil {
			fmt.Printf("Error creating tilt: %v\n", err)
			return nil
		}
		return tilt
	case "one_subtilt":
		// Tilt with one sub-tilt
		CreateTilt(filepath, 2, CreateRandomRecievers(), []int{100}, nil, 100)
		tilt, err := CreateTilt(filepath, 1, CreateRandomRecievers(), []int{80, 20}, []int{2}, 100)
		if err != nil {
			fmt.Printf("Error creating tilt: %v\n", err)
			return nil
		}
		return tilt
	case "two_subtilts":
		// Tilt with two sub-tilts (matches original behavior)
		_, err := CreateTilt(filepath, 3, CreateRandomRecievers(), []int{100}, nil, 100)
		if err != nil {
			fmt.Printf("Error creating tilt: %v\n", err)
			return nil
		}
		_, err = CreateTilt(filepath, 2, CreateRandomRecievers(), []int{100}, nil, 100)
		if err != nil {
			fmt.Printf("Error creating tilt: %v\n", err)
			return nil
		}
		tilt, err := CreateTilt(filepath, 1, CreateRandomRecievers(), []int{20, 70, 10}, []int{3, 2}, 100)
		if err != nil {
			fmt.Printf("Error creating tilt: %v\n", err)
			return nil
		}
		return tilt
	case "nested":
		// A nested tilt structure with multiple levels
		_, err := CreateTilt(filepath, 3, CreateRandomRecievers(), []int{100}, nil, 100)
		if err != nil {
			fmt.Printf("Error creating tilt: %v\n", err)
			return nil
		}
		_, err = CreateTilt(filepath, 2, CreateRandomRecievers(), []int{80, 20}, []int{3}, 100)
		if err != nil {
			fmt.Printf("Error creating tilt: %v\n", err)
			return nil
		}
		tilt, err := CreateTilt(filepath, 1, CreateRandomRecievers(), []int{80, 20}, []int{2}, 100)
		if err != nil {
			fmt.Printf("Error creating tilt: %v\n", err)
			return nil
		}
		return tilt
	default:
		return nil
	}
}
