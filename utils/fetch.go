package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// saveTiltData saves the tilt data to a file
func saveTiltData(filePath string, data map[string][]string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for tilt, recipients := range data {
		if _, err := file.WriteString(fmt.Sprintf("Tilt: %s\n", tilt)); err != nil {
			return err
		}
		for _, recipient := range recipients {
			if _, err := file.WriteString(fmt.Sprintf("Recipient: %s\n", recipient)); err != nil {
				return err
			}
		}
	}
	return nil
}

// FetchTiltData recursively fetches tilt and sub-tilt details
func FetchTiltData(t Tilt) (map[string][]string, error) {
	data := make(map[string][]string)
	if err := fetchTiltRecursive(t, data); err != nil {
		return nil, err
	}
	filePath := "/Users/apple/Documents/GitHub/tilt-validator-main/utils/current-tilt-data.txt"
	if err := saveTiltData(filePath, data); err != nil {
		return nil, err
	}
	return data, nil
}

// readTiltData reads the tilt data from a file
func ReadTiltData(filePath string) (map[string][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data := make(map[string][]string)
	var currentTilt string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Tilt: ") {
			currentTilt = strings.TrimPrefix(line, "Tilt: ")
			data[currentTilt] = []string{}
		} else if strings.HasPrefix(line, "Recipient: ") {
			recipient := strings.TrimPrefix(line, "Recipient: ")
			data[currentTilt] = append(data[currentTilt], recipient)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return data, nil
}

func fetchTiltRecursive(t Tilt, data map[string][]string) error {
	if len(t.SubTilts) == 0 && len(t.Recipients) == 0 {
		return fmt.Errorf("tilt %s has no recipients or sub-tilts", t.Name)
	}

	recipients := make([]string, len(t.Recipients))
	for i, r := range t.Recipients {
		recipients[i] = r.String()
	}
	data[t.Name] = recipients

	for _, sub := range t.SubTilts {
		if err := fetchTiltRecursive(sub, data); err != nil {
			return err
		}
	}
	return nil
}
