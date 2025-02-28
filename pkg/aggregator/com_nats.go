package main

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

// docker run -d --name nats-server -p 4222:4222 nats:latest

const NATS_URL = "nats://localhost:4222"

func main() {
	// Connect to server
	nc, err := nats.Connect(NATS_URL, nats.Name("Validator Node"))
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// Subscribe to a subject
	_, err = nc.Subscribe("secure_channel", func(msg *nats.Msg) {
		fmt.Printf("Received message: %s\n", string(msg.Data))
	})
	if err != nil {
		log.Fatal(err)
	}

	// test
	go func() {
		for {
			msg := fmt.Sprintf("Validator %d: Secure data exchange", time.Now().Unix())
			nc.Publish("secure_channel", []byte(msg))
			time.Sleep(3 * time.Second)
		}
	}()

	select {}
}
