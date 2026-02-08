package main

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Job represents a task to be processed.
type Job struct {
	ID      int
	Payload int // Some data to process
}

// Result holds the outcome of a processed job.
type Result struct {
	JobID    int
	Output   int
	WorkerID int
}

// Worker is a struct that can process jobs.
type Worker struct {
	ID       int
	jobQueue <-chan Job
	results  chan<- Result
	wg       *sync.WaitGroup
}

func NewWorker(id int, jobQueue <-chan Job, results chan<- Result, wg *sync.WaitGroup) *Worker {
	return &Worker{ID: id, jobQueue: jobQueue, results: results, wg: wg}
}

// Start makes the worker listen for jobs on the jobQueue.
// It should process jobs until the jobQueue is closed.
func (w *Worker) Start() {
	// TODO: Implement this method.
	// It should loop over the w.jobQueue channel.
	// For each job, it should "process" it (e.g., Job.Payload * 2),
	// create a Result, and send it to the w.results channel.
	// It must call w.wg.Done() before it exits.

	for job := range w.jobQueue {
		fmt.Println("Worker", w.ID, "processing job", job.ID)
		result := job.Payload * 2
		w.results <- Result{JobID: job.ID, Output: result, WorkerID: w.ID}
	}
	w.wg.Done()
}

// Dispatcher coordinates the workers.
type Dispatcher struct {
	jobQueue chan Job
	results  chan Result
	wg       *sync.WaitGroup
}

// NewDispatcher creates a new dispatcher with a specified number of workers.

func NewDispatcher(numWorkers, maxJobs int) *Dispatcher {
	return &Dispatcher{
		jobQueue: make(chan Job, maxJobs),
		results:  make(chan Result, maxJobs),
		wg:       &sync.WaitGroup{},
	}
}

func (d *Dispatcher) Start(numWorkers int) {
	// TODO: Implement this method.
	// It should create 'numWorkers' instances of NewWorker
	// and start each one in a separate goroutine.
	for i := 0; i < numWorkers; i++ {
		d.wg.Add(1)
		worker := NewWorker(i, d.jobQueue, d.results, d.wg)
		go worker.Start()

	}
}

// Dispatch sends a job to the worker pool.
func (d *Dispatcher) Dispatch(job Job) {
	d.jobQueue <- job
}

// Wait waits for all jobs to be processed and then stops the dispatcher.
func (d *Dispatcher) Wait() {
	// Close the job queue to signal to workers that no more jobs are coming.
	close(d.jobQueue)
	// Wait for all workers to finish their current jobs.
	d.wg.Wait()
	// Close the results channel.
	close(d.results)
}

func main() {
	numJobs := 5
	numWorkers := 5

	dispatcher := NewDispatcher(numWorkers, numJobs)
	dispatcher.Start(numWorkers)

	// Dispatch all jobs
	for i := 0; i < numJobs; i++ {
		dispatcher.Dispatch(Job{ID: i, Payload: 10})
	}

	// Wait for all jobs to complete
	dispatcher.Wait()

	// Collect results
	results := []Result{}
	for result := range dispatcher.results {
		results = append(results, result)
	}

	// Assert that the processing was correct
	sum := 0
	for _, res := range results {
		// In our simple case, output should be payload * 2

		sum += res.Output
	}

	fmt.Println("Total sum of results: Is it 100?", sum)
}

// This test helps verify concurrency. It should run faster than numJobs * sleepDuration.
func TestWorkerPoolConcurrency(t *testing.T) {
	numJobs := 10
	numWorkers := 5
	sleepDuration := 100 * time.Millisecond

	// Overwrite the worker's processing logic for this test
	jobProcessor := func(job Job) int {
		time.Sleep(sleepDuration)
		return job.Payload
	}

	jobQueue := make(chan Job, numJobs)
	results := make(chan Result, numJobs)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		worker := &Worker{
			ID: i, jobQueue: jobQueue, results: results, wg: &wg,
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range worker.jobQueue {
				output := jobProcessor(job)
				worker.results <- Result{JobID: job.ID, Output: output, WorkerID: worker.ID}
			}
		}()
	}

	startTime := time.Now()

	// Dispatch jobs
	for i := 0; i < numJobs; i++ {
		jobQueue <- Job{ID: i, Payload: i}
	}
	close(jobQueue)

	wg.Wait()
	close(results)

	elapsedTime := time.Since(startTime)
	totalSerialTime := time.Duration(numJobs) * sleepDuration

	// The key assertion: proves work was done in parallel.
	assert.Less(t, elapsedTime, totalSerialTime, "Execution time should be less than serial execution time")
	// With 5 workers and 10 jobs, it should take roughly 2 * sleepDuration
	assert.Greater(t, elapsedTime, sleepDuration, "Execution time should be more than a single job's duration")

	t.Logf("Total time for %d jobs with %d workers: %s (Serial time would be: %s)", numJobs, numWorkers, elapsedTime, totalSerialTime)
}
