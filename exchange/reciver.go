package exchange

import (
	"encoding/csv"
	"os"
)

func (t *Transport) ReadMsg() ([][]string, error) {
	fileName := t.GetFileName()
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// After reading the line, delete it
	if len(records) > 0 {
		records = records[1:]
	}

	return records, nil
}

// func main() {
// 	filename := "/Users/apple/Documents/GitHub/new_tss/transformer.csv"

// 	// partyID := 1
// 	// t := Transport{partyID}

// 	records, err := ReadCSV(filename)
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		return
// 	}

// 	for _, record := range records {
// 		fmt.Println(record)
// 	}
// }
