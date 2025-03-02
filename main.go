package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/replit/tilt-validator/cmd"
)

func startHTTPServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `Tilt Validator Simulation
Status: Running

Services:
- HTTP Server: Running on :5000
- NATS Server: Running on :4221
`)
	})

	log.Printf("Starting HTTP server on port 5000")
	if err := http.ListenAndServe("0.0.0.0:5000", nil); err != nil {
		log.Printf("HTTP server error: %v", err)
	}
}

func main() {
	// Start HTTP server in a goroutine
	go startHTTPServer()

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
