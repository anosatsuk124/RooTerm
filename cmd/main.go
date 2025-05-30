package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Payload represents the JSON payload to send
type Payload struct {
	Message string `json:"message"`
	Time    string `json:"time"`
}

func main() {
	// Define command-line flags
	host := flag.String("host", "127.0.0.1", "Target host")
	port := flag.Int("port", 9421, "Target port")
	msg := flag.String("msg", "", "Message to send")
	flag.Parse()

	if *msg == "" {
		fmt.Fprintln(os.Stderr, "Error: -msg is required")
		os.Exit(1)
	}

	// Prepare payload
	payload := Payload{
		Message: *msg,
		Time:    time.Now().Format(time.RFC3339),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "JSON marshal error: %v\n", err)
		os.Exit(1)
	}

	// Send HTTP POST request
	url := fmt.Sprintf("http://%s:%d/", *host, *port)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "HTTP request error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Server responded with status: %s\n", resp.Status)
		os.Exit(1)
	}

	fmt.Println("Message sent successfully.")
}
