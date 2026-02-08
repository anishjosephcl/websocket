package main

import (
	"fmt"
	"time"
)

// CustomContext defines a struct with a timeout channel
type CustomContext struct {
	TimeoutChan chan bool
}

// worker simulates long-running work (sleep > 40s)
func worker(ctx *CustomContext) {
	fmt.Println("Worker started...")

	for {
		select {
		case <-ctx.TimeoutChan:
			fmt.Println("Timeout received, stopping worker...")
			return
		default:
			fmt.Println("Worker is running... (sleeping for 45s)")
			// Sleep longer than the main function's timeout
			time.Sleep(45 * time.Second)
		}
	}
}

func main9() {
	// Initialize custom context
	ctx := &CustomContext{TimeoutChan: make(chan bool)}

	// Start the worker in a goroutine
	go worker(ctx)

	// Run for 20 seconds, then send timeout signal
	time.Sleep(20 * time.Second)
	fmt.Println("Main: sending timeout signal.")
	ctx.TimeoutChan <- true

	// Allow worker to stop cleanly
	time.Sleep(1 * time.Second)
	fmt.Println("Main exiting.")
}
