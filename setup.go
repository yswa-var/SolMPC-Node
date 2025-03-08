package exchange

import (
	"strconv"
)

type Transport struct {
	partyID int
	parties []int
}

func NewTransport(partyID int, parties []int) *Transport {
	return &Transport{partyID: partyID, parties: parties}
}

func (t *Transport) GetFileName() string {
	return "./Transport/" + strconv.Itoa(t.partyID) + ".csv"
}

func (t *Transport) GetReceiverFileName(id string) string {
	return "./Transport/" + id + ".csv"
}

func (t *Transport) getParties() []int {
	return t.parties
}

// func (t *Transport) ReadCSV(filename string) ([][]string, error) {
// 	file, err := os.Open(filename)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer file.Close()

// 	reader := csv.NewReader(file)
// 	reader.LazyQuotes = true
// 	records, err := reader.ReadAll()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return records, nil
// }

// func (t *Transport) SaveToCSV(from int, broadcast bool, to int, message string) error {
// 	file, err := os.OpenFile("transformer.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()

// 	writer := csv.NewWriter(file)
// 	defer writer.Flush()

// 	record := []string{
// 		strconv.Itoa(from),
// 		strconv.FormatBool(broadcast),
// 		strconv.Itoa(to),
// 		message,
// 	}

// 	if err := writer.Write(record); err != nil {
// 		return err
// 	}

// 	return nil
// }
