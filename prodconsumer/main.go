package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"
)

// MessageStore is a simple in-memory store that returns a random message.
type MessageStore struct {
	mu       sync.RWMutex
	messages []string
}

func NewMessageStore() *MessageStore {
	return &MessageStore{
		messages: []string{
			"alpha",
			"beta",
			"gamma",
			"delta",
			"epsilon",
		},
	}
}

func (s *MessageStore) GetRandom() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.messages) == 0 {
		return ""
	}
	i := rand.Intn(len(s.messages))
	return s.messages[i]
}

// Result is what producers send to the consumer.
type Result struct {
	ProducerID int
	Raw        string
	Processed  string
	At         time.Time
}

func producer(id int, store *MessageStore, out chan<- Result) {
	for {
		time.Sleep(10 * time.Second) // run every 10 seconds

		raw := store.GetRandom()
		if raw == "" {
			continue
		}

		// "Processing" step – customize as needed.
		processed := fmt.Sprintf("processed(%s)", raw)

		res := Result{
			ProducerID: id,
			Raw:        raw,
			Processed:  processed,
			At:         time.Now(),
		}

		out <- res
	}
}

func consumer(filePath string, in <-chan Result) error {
	f, err := os.OpenFile(filePath,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	for res := range in {
		line := fmt.Sprintf("%s [producer=%d] raw=%s processed=%s\n",
			res.At.Format(time.RFC3339),
			res.ProducerID,
			res.Raw,
			res.Processed,
		)
		if _, err := f.WriteString(line); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	store := NewMessageStore()
	results := make(chan Result, 10) // buffered to decouple producers/consumer

	const numProducers = 3

	// Start producers.
	for i := 1; i <= numProducers; i++ {
		go producer(i, store, results)
	}

	// Start consumer (writer).
	go func() {
		if err := consumer("results.log", results); err != nil {
			fmt.Println("consumer error:", err)
		}
	}()

	// Block forever for this demo.
	select {}
}
