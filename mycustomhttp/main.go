// main.go

package main

import (
	"fmt"
	"log"
	"mycustomhttp/myhttp"
	"time"
	// Import our completely custom myhttp package
)

// A simple handler for the root path
func rootHandler(w *myhttp.ResponseWriter, r *myhttp.Request) {
	fmt.Println("Handling request for:", r.Path)
	w.Header()["Content-Type"] = "text/html; charset=utf-8"
	responseHTML := "<h1>Welcome!</h1><p>You are at the root path.</p>"
	w.Write([]byte(responseHTML))
}

// A handler for a different path
func timeHandler(w *myhttp.ResponseWriter, r *myhttp.Request) {
	fmt.Println("Handling request for:", r.Path)
	currentTime := time.Now().Format(time.RFC1123)

	// WriteHeader is optional, defaults to 200
	w.WriteHeader(200)

	fmt.Fprintf(w, "The current time is: %s", currentTime)
}

// A simple router to direct traffic
func router(w *myhttp.ResponseWriter, r *myhttp.Request) {
	switch r.Path {
	case "/":
		rootHandler(w, r)
	case "/time":
		timeHandler(w, r)
	default:
		w.WriteHeader(404)
		fmt.Fprint(w, "404 Not Found")
	}
}

func main() {
	// The handler is now our simple router function
	handler := myhttp.HandlerFunc(router)

	log.Println("Starting completely custom server on http://localhost:8081")

	// Call ListenAndServe from our from-scratch myhttp package.
	log.Fatal(myhttp.ListenAndServe(":8081", handler))
}
