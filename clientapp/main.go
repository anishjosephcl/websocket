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

func main1() {
	// Command line flags
	serverURL := flag.String("url", "ws://localhost:8080/ws/messages", "WebSocket server URL")
	flag.Parse()

	log.Println("Connecting to WebSocket server:", *serverURL)

	// Parse URL
	u, err := url.Parse(*serverURL)
	if err != nil {
		log.Fatal("Invalid URL:", err)
	}

	// WebSocket dialer with reasonable timeouts
	dialer := &websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	// Connect to server
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("WebSocket dial error:", err)
	}
	defer conn.Close()

	log.Println(" Connected to WebSocket server")

	// Channel for graceful shutdown
	done := make(chan struct{})
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Handle incoming messages
	go func() {
		defer close(done)
		for {
			// Read message
			mt, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				return
			}

			// Print message details
			fmt.Printf("\n MESSAGE #%d (Type: %s):\n", mt, mt)
			fmt.Printf(" Raw: %s\n", message)

			// Try to parse as JSON for pretty printing
			var pretty interface{}
			if json.Unmarshal(message, &pretty) == nil {
				prettyBytes, _ := json.MarshalIndent(pretty, "", "  ")
				fmt.Printf(" JSON:\n%s\n", prettyBytes)
			}

			fmt.Println("─" + strings.Repeat("─", 60))
		}
	}()

	// Wait for interrupt signal
	select {
	case <-done:
		log.Println("WebSocket connection closed by server")
	case <-interrupt:
		log.Println(" Interrupt signal received, closing...")
		err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Printf("Close message error: %v", err)
		}
		<-done
	}
}
