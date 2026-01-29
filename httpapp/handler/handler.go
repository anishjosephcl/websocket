package handler

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
)

const traceIDKey string = "traceID"

type Message struct {
	UserID  string `json:"userId"`
	Message string `json:"message"`
}

var MessageStoreInstance *MessageStore

func InitializeMessageStore() {
	MessageStoreInstance = &MessageStore{
		MsgChan: make(chan Message, 100),
	}
	go broadCastToRegisteredClients()

}
func StoreMessageHandler(w http.ResponseWriter, r *http.Request) {
	// Get TraceID from context
	traceID, ok := r.Context().Value(traceIDKey).(string)
	if !ok {
		traceID = "unknown"
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Parse JSON request body
	var msg Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		log.Printf("TraceID=%s JSON decode error: %v", traceID, err)
		return
	}

	// Simulate storing message (in real app, save to DB)
	log.Printf("TraceID=%s Stored message: %+v", traceID, msg)

	// Write message to MessageStore channel

	MessageStoreInstance.MsgChan <- msg
	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"traceID": traceID,
		"status":  "success",
		"message": "Message stored successfully",
		"data":    msg,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("TraceID=%s JSON encode error: %v", traceID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
func ListMessagesHandler(w http.ResponseWriter, r *http.Request) {
	// Get TraceID from context
	traceID, ok := r.Context().Value(traceIDKey).(string)
	if !ok {
		traceID = "unknown"
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Simulate retrieving messages (in real app, fetch from DB)
	messages := []Message{
		{UserID: "user1", Message: "Hello World"},
		{UserID: "user2", Message: "Test Message"},
		{UserID: "user3", Message: "Sample Data"},
		{UserID: "user4", Message: "Example Text"},
		{UserID: "user5", Message: "Demo Message"},
		{UserID: "user6", Message: "Practice Data"},
		{UserID: "user7", Message: "Trial Message"},
		{UserID: "user8", Message: "Mock Data"},
		{UserID: "user9", Message: "Placeholder Text"},
		{UserID: "user10", Message: "Template Message"},
	}

	log.Printf("TraceID=%s Retrieved %d messages", traceID, len(messages))

	t, err := template.New("messages").Parse(tmpl)
	if err != nil {
		log.Printf("TraceID=%s Template parse error: %v", traceID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		TraceID  string
		Count    int
		Messages []Message
	}{
		TraceID:  traceID,
		Count:    len(messages),
		Messages: messages,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.Execute(w, data); err != nil {
		log.Printf("TraceID=%s Template execute error: %v", traceID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
