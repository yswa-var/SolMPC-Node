package exchange

import (
	"encoding/csv"
	"os"
	"strconv"
)

func (t *Transport) SendMsg(from int, broadcast bool, to int, message string) error {

	if broadcast {
		for _, party := range t.getParties() {
			if party == from {
				continue
			}
			file, err := os.OpenFile(t.GetReceiverFileName(strconv.Itoa(party)), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			defer file.Close()
			writer := csv.NewWriter(file)
			defer writer.Flush()

			record := []string{
				strconv.Itoa(from),
				strconv.FormatBool(broadcast),
				strconv.Itoa(party),
				message,
			}

			if err := writer.Write(record); err != nil {
				return err
			}
		}
	} else {
		file, err := os.OpenFile(t.GetReceiverFileName(strconv.Itoa(to)), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer file.Close()
		writer := csv.NewWriter(file)
		defer writer.Flush()

		record := []string{
			strconv.Itoa(from),
			strconv.FormatBool(broadcast),
			strconv.Itoa(to),
			message,
		}

		if err := writer.Write(record); err != nil {
			return err
		}
	}
	return nil

}
