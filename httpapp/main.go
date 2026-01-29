package main

import (
	"context"
	"fmt"
	"log"
	"messagefeedapp/httpapp/handler"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

const traceIDKey string = "traceID"

func main() {
	mux := http.NewServeMux()

	handler.InitializeMessageStore()
	// Register the storemessage endpoint
	mux.HandleFunc("POST /storemessage", handler.StoreMessageHandler)
	// Register the list endpoint to get 10 messages
	mux.HandleFunc("GET /list", handler.ListMessagesHandler)
	// Static file server for /about - serves files from ./static/about/
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/about/", http.StripPrefix("/about/", fs))

	mux.HandleFunc("GET /ws/messages", handler.WsMessagesHandler)

	mux.HandleFunc("GET /ws/client", handler.WsClientHandler)

	// Apply middleware to the entire mux
	handler := traceMiddleware(mux)

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// Middleware to add TraceID to request context
func traceMiddleware(next http.Handler) http.Handler {
	counter := &atomic.Int64{}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate unique TraceID
		traceID := strconv.FormatInt(counter.Add(1), 10) + "_" + time.Now().Format("20060102_150405")

		// Add TraceID to context
		ctx := context.WithValue(r.Context(), traceIDKey, traceID)

		// Add TraceID to request headers for logging
		r = r.WithContext(ctx)
		r.Header.Set("X-Trace-ID", traceID)

		log.Printf("TraceID=%s %s %s", traceID, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
