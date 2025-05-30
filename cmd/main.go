package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

// Payload represents the JSON payload to send
type Payload struct {
	Message string `json:"message"`
	Time    string `json:"time"`
}

func main() {
	// ホストとポートのみフラグで受け取る
	host := flag.String("host", "127.0.0.1", "Target host")
	port := flag.Int("port", 9421, "Target port")
	flag.Parse()

	// WebSocket サーバに接続
	url := fmt.Sprintf("ws://%s:%d/", *host, *port)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WebSocket dial error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	// 対話モードで標準入力を受け付け
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Connected to", url)
	fmt.Println("Enter messages to send. Ctrl+D to exit.")
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		text := scanner.Text()
		payload := Payload{Message: text, Time: time.Now().Format(time.RFC3339)}
		data, err := json.Marshal(payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "JSON marshal error: %v\n", err)
			continue
		}
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			fmt.Fprintf(os.Stderr, "Write error: %v\n", err)
			break
		}
		_, resp, err := conn.ReadMessage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Read error: %v\n", err)
			break
		}
		fmt.Printf("Server response: %s\n", resp)
	}
	fmt.Println("Disconnected.")
}
