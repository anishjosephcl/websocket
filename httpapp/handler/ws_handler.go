package handler

import (
	"encoding/json"
	"html"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for demo
	},
}

// WebSocket handler - sends last 10 messages then closes
func WsMessagesHandler(w http.ResponseWriter, r *http.Request) {
	traceID, _ := r.Context().Value(traceIDKey).(string)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("TraceID=%s WebSocket upgrade failed: %v", traceID, err)
		return
	}
	defer conn.Close()
	// Create some dummy messages
	recent := []struct {
		ID      int
		Content string
		Author  string
		Created time.Time
	}{
		{ID: 1, Content: "Hello World!", Author: "Alice", Created: time.Now().Add(-10 * time.Minute)},
		{ID: 2, Content: "How are you?", Author: "Bob", Created: time.Now().Add(-8 * time.Minute)},
		{ID: 3, Content: "Great, thanks!", Author: "Alice", Created: time.Now().Add(-5 * time.Minute)},
		{ID: 4, Content: "Working on websockets", Author: "Charlie", Created: time.Now().Add(-3 * time.Minute)},
		{ID: 5, Content: "Looks good!", Author: "Bob", Created: time.Now().Add(-1 * time.Minute)},
	}
	// Send messages as JSON array
	for _, msg := range recent {
		wsMsg := map[string]interface{}{
			"type":    "message",
			"id":      msg.ID,
			"content": html.EscapeString(msg.Content),
			"author":  html.EscapeString(msg.Author),
			"created": msg.Created,
		}
		data, err := json.Marshal(wsMsg)
		if err != nil {
			log.Printf("TraceID=%s WebSocket JSON marshal error: %v", traceID, err)
			return
		}

		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("TraceID=%s WebSocket write error: %v", traceID, err)
			return
		}
		log.Printf("TraceID=%s Sent message #%d via WebSocket", traceID, msg.ID)

		// Small delay between messages
		time.Sleep(100 * time.Millisecond)
	}
}
