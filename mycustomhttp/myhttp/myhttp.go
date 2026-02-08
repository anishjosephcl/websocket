// myhttp/myhttp.go

package myhttp

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

// Request holds the parsed data from an incoming HTTP request.
type Request struct {
	Method  string
	Path    string
	Proto   string
	Headers map[string]string
	Body    io.Reader
	Conn    net.Conn // The underlying connection
}

// ResponseWriter is a helper to construct a response.
// For this simple example, it holds the data and writes it back
// to the connection when the handler is done.
type ResponseWriter struct {
	conn       io.Writer
	headers    map[string]string
	statusCode int
	body       []byte
}

// Header returns the map of headers.
func (rw *ResponseWriter) Header() map[string]string {
	return rw.headers
}

// Write appends data to the response body.
func (rw *ResponseWriter) Write(data []byte) (int, error) {
	rw.body = append(rw.body, data...)
	return len(data), nil
}

// WriteHeader sets the HTTP status code for the response.
func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
}

// flush writes the entire response to the connection.
func (rw *ResponseWriter) flush() {
	// Default to 200 OK if not set
	if rw.statusCode == 0 {
		rw.statusCode = 200
	}

	// Write the status line
	fmt.Fprintf(rw.conn, "HTTP/1.1 %d %s\r\n", rw.statusCode, statusText(rw.statusCode))

	// Set a default content type if not present
	if _, ok := rw.headers["Content-Type"]; !ok {
		rw.headers["Content-Type"] = "text/plain; charset=utf-8"
	}

	// Add content length header
	rw.headers["Content-Length"] = strconv.Itoa(len(rw.body))

	// Write headers
	for key, value := range rw.headers {
		fmt.Fprintf(rw.conn, "%s: %s\r\n", key, value)
	}

	// End of headers
	fmt.Fprintf(rw.conn, "\r\n")

	// Write the body
	rw.conn.Write(rw.body)
}

// Handler is an interface that objects can implement to be an HTTP handler.
type Handler interface {
	ServeHTTP(*ResponseWriter, *Request)
}

// HandlerFunc is a function type that acts as an adapter.
type HandlerFunc func(*ResponseWriter, *Request)

// ServeHTTP calls the underlying function, making HandlerFunc satisfy the Handler interface.
func (f HandlerFunc) ServeHTTP(w *ResponseWriter, r *Request) {
	f(w, r)
}

// ListenAndServe starts our custom server.
func ListenAndServe(addr string, handler Handler) error {
	// 1. Listen for incoming TCP connections
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Printf("Listening on %s", addr)

	for {
		// 2. Accept a new connection
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		// 3. Handle the new connection in a new goroutine
		go handleConnection(conn, handler)
	}
}

// handleConnection reads and parses the request, calls the handler, and writes the response.
func handleConnection(conn net.Conn, handler Handler) {
	defer conn.Close()

	// Create a buffered reader for the connection
	reader := bufio.NewReader(conn)

	// --- Parse the Request ---
	// 1. Read the request line (e.g., "GET /hello HTTP/1.1")
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Error reading request line: %v", err)
		return
	}
	parts := strings.Split(strings.TrimSpace(requestLine), " ")
	if len(parts) != 3 {
		log.Printf("Malformed request line: %s", requestLine)
		return
	}

	req := &Request{
		Method:  parts[0],
		Path:    parts[1],
		Proto:   parts[2],
		Headers: make(map[string]string),
		Conn:    conn,
	}

	// 2. Read headers until we get an empty line
	for {
		line, err := reader.ReadString('\n')
		if err != nil || strings.TrimSpace(line) == "" {
			break // End of headers
		}
		headerParts := strings.SplitN(line, ":", 2)
		if len(headerParts) == 2 {
			req.Headers[strings.TrimSpace(headerParts[0])] = strings.TrimSpace(headerParts[1])
		}
	}

	// For simplicity, we are not parsing the body. A full implementation
	// would need to read 'Content-Length' bytes from the reader.
	req.Body = reader

	// --- Create a Response Writer ---
	res := &ResponseWriter{
		conn:    conn,
		headers: make(map[string]string),
	}

	// --- Call the User's Handler ---
	handler.ServeHTTP(res, req)

	// --- Write the Response ---
	res.flush()
}

// statusText returns a text for the HTTP status code.
// A very simplified version.
func statusText(code int) string {
	switch code {
	case 200:
		return "OK"
	case 404:
		return "Not Found"
	case 500:
		return "Internal Server Error"
	}
	return ""
}
