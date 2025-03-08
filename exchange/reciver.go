package exchange

import (
	"crypto/md5"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"strconv"
	"time"
)

type Msg struct {
	From      int
	Broadcast bool
	To        int
	Message   []byte
}

func (t *Transport) ReadMsg() ([][]string, error) {
	fileName := t.GetFileName()
	t.Mutex.Lock()
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	record, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	t.Mutex.Unlock()

	if len(record) > 0 {
		record = record[:]
	}

	return record, nil
}

func (t *Transport) ReadMsgToChannel(ch chan<- []byte) error {
	records, err := t.ReadMsg()
	if err != nil {
		return err
	}

	for _, record := range records {
		from, _ := strconv.Atoi(record[0])
		broadcast, _ := strconv.ParseBool(record[1])
		to, _ := strconv.Atoi(record[2])
		msg_, _ := hex.DecodeString(record[3])
		msg := Msg{
			From:      from,
			Broadcast: broadcast,
			To:        to,
			Message:   msg_,
		}
		byteMsg, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		ch <- byteMsg
	}
	return nil
}

func (t *Transport) getFileHash() (string, error) {
	fileName := t.GetFileName()
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Create a new hash
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return string(hash.Sum(nil)), nil
}

func (t *Transport) WatchFile(interval time.Duration, ch chan<- []byte) {
	var lastHash string

	// Infinite loop to check the file hash
	for {
		currentHash, err := t.getFileHash()
		if err != nil {
			time.Sleep(interval)
			continue
		}

		// If the hash is different, read the message to the channel
		if currentHash != lastHash {
			lastHash = currentHash
			err := t.ReadMsgToChannel(ch)
			if err != nil {
				time.Sleep(interval)
				continue
			}
		}

		time.Sleep(interval)
	}
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
