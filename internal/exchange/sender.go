package exchange

import (
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
)

func (t *Transport) SendMsg(message []byte, broadcast bool, to uint16) {
	from := t.partyID
	if broadcast {
		for _, party_ := range t.getParties() {
			party := int(party_)
			if int(party) == from {
				continue
			}
			t.Mutex.Lock()
			file, err := os.OpenFile(t.GetReceiverFileName(strconv.Itoa(int(party))), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Println("Error opening file:", err)
				return
			}
			defer file.Close()
			writer := csv.NewWriter(file)
			defer writer.Flush()
			record := []string{
				strconv.Itoa(from),
				strconv.FormatBool(broadcast),
				strconv.Itoa(int(party)),
				hex.EncodeToString(message),
			}

			if err := writer.Write(record); err != nil {
				fmt.Println("Error writing to file:", err)
				return
			}
			t.Mutex.Unlock()

		}
	} else {
		t.Mutex.Lock()
		file, err := os.OpenFile(t.GetReceiverFileName(strconv.Itoa(int(to))), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return
		}
		defer file.Close()
		writer := csv.NewWriter(file)
		defer writer.Flush()

		record := []string{
			strconv.Itoa(from),
			strconv.FormatBool(broadcast),
			strconv.Itoa(int(to)),
			hex.EncodeToString(message),
		}

		if err := writer.Write(record); err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}
		t.Mutex.Unlock()

	}
	return

}
