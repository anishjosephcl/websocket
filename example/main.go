package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// Message represents a message to be stored and broadcasted.
type Message struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Client represents a client connection.
type Client struct {
	ID      string
	Conn    net.Conn
	Channel chan Message
}

// MessageStore manages message storage and client broadcasting.
type MessageStore struct {
	mu        sync.RWMutex
	messages  []Message
	clients   map[string]*Client
	broadcast chan Message
	done      chan struct{}
}

// NewMessageStore creates a new MessageStore.
func NewMessageStore() *MessageStore {
	ms := &MessageStore{
		messages:  make([]Message, 0),
		clients:   make(map[string]*Client),
		broadcast: make(chan Message, 100),
		done:      make(chan struct{}),
	}
	go ms.broadcastLoop()
	return ms
}

// Store adds a message to the store and broadcasts it to all clients.
func (ms *MessageStore) Store(msg Message) {
	ms.messages = append(ms.messages, msg)
	ms.broadcast <- msg
}

// AddClient adds a new client with its channel.
func (ms *MessageStore) AddClient(clientID string, conn net.Conn) *Client {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ch := make(chan Message, 10)
	client := &Client{
		ID:      clientID,
		Conn:    conn,
		Channel: ch,
	}
	ms.clients[clientID] = client

	// Start goroutine to write from channel to socket
	go func() {
		defer conn.Close()
		for msg := range ch {
			data, _ := json.Marshal(msg)
			fmt.Fprintf(conn, "%s\n", data)
		}
	}()

	return client
}

// RemoveClient removes a client.
func (ms *MessageStore) RemoveClient(clientID string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if client, ok := ms.clients[clientID]; ok {
		close(client.Channel)
		delete(ms.clients, clientID)
	}
}

// GetMessages returns all stored messages.
func (ms *MessageStore) GetMessages() []Message {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return append([]Message(nil), ms.messages...)
}

// broadcastLoop reads from the broadcast channel and sends to all client channels.
func (ms *MessageStore) broadcastLoop() {
	for {
		select {
		case msg := <-ms.broadcast:
			ms.mu.RLock()
			for _, client := range ms.clients {
				select {
				case client.Channel <- msg:
				default:
					// Skip if channel is full
				}
			}
			ms.mu.RUnlock()
		case <-ms.done:
			return
		}
	}
}

// Close shuts down the MessageStore, closing all client channels.
func (ms *MessageStore) Close() {
	close(ms.done)
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, client := range ms.clients {
		close(client.Channel)
	}
}

func main() {
	ms := NewMessageStore()

	// Simulate server listening (for demo, use a dummy listener)
	listener, _ := net.Listen("tcp", ":8090")
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println(err)
				continue
			}
			clientID := fmt.Sprintf("client-%d", time.Now().UnixNano())
			ms.AddClient(clientID, conn)
			log.Printf("Added client %s", clientID)
		}
	}()

	// Simulate storing messages
	go func() {
		for i := 1; i <= 5; i++ {
			msg := Message{
				ID:        fmt.Sprintf("msg-%d", i),
				Content:   fmt.Sprintf("Hello from message %d", i),
				Timestamp: time.Now(),
			}
			ms.Store(msg)
			time.Sleep(2 * time.Second)
		}
	}()

	// Run for a while
	time.Sleep(20 * time.Second)
	ms.Close()
}
func client() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal("Connection failed:", err)
	}
	defer conn.Close()

	clientID := fmt.Sprintf("client-%d", time.Now().UnixNano())
	fmt.Println("Connected as", clientID)
	// Send client ID handshake (server expects this for AddClient)
	fmt.Fprintf(conn, "ID:%s\n", clientID)

	// Read broadcasted messages
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var msg Message
		if err := json.Unmarshal([]byte(scanner.Text()), &msg); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}
		fmt.Printf("Received: [%s] %s\n", msg.ID, msg.Content)
	}

	if err := scanner.Err(); err != nil {
		log.Println("Read error:", err)
	}
}
