package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	// Command line flags
	serverURL := flag.String("url", "ws://localhost:8080/ws/client", "WebSocket server URL")
	flag.Parse()

	log.Println("Connecting to WebSocket server:", *serverURL)

	// Channel for graceful shutdown

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Just Test with multiple clients
	go connectAndGetUpdates(*serverURL)
	go connectAndGetUpdates(*serverURL)

	// Wait for interrupt signal
	<-interrupt
	log.Println("Interrupt received, shutting down...")

}
func connectAndGetUpdates(serverURL string) {
	log.Printf("Connecting to %s\n", serverURL)

	u, err := url.Parse(serverURL)
	if err != nil {
		log.Printf("Invalid URL: %v\n", err)
		return
	}

	dialer := &websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		log.Printf("WebSocket dial error: %v\n", err)
		return
	}
	defer conn.Close()

	log.Printf("Connected to WebSocket server\n")

	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v\n", err)
			return
		}

		fmt.Printf("\nMESSAGE (Type: %d):\n", mt)
		fmt.Printf("Raw: %s\n", message)

		var pretty interface{}
		if json.Unmarshal(message, &pretty) == nil {
			prettyBytes, _ := json.MarshalIndent(pretty, "", "  ")
			fmt.Printf("JSON:\n%s\n", prettyBytes)
		}

		fmt.Println(strings.Repeat("─", 60))
	}
}
