// Package workerpool provides a simple and efficient implementation of a worker pool in Go.
// It allows concurrent processing of tasks using a fixed number of workers, making it ideal
// for managing workloads in a controlled and scalable manner.
package workerpool

import (
	"context"
	"log"
	"sync"
	"time"
)

// Task defines an interface for tasks to be executed by workers in the pool.
type Task interface {
	Execute() (interface{}, error) // Executes the task and returns the result or an error.
}

// Result stores the output of a task execution along with any error and the task itself.
type Result struct {
	Output interface{} // The output of the task execution.
	Err    error       // Any error that occurred during the task execution.
	Task   Task        // The task that was executed.
}

// WorkerPool manages a pool of workers and task processing.
type WorkerPool struct {
	workersCount int                // The number of workers in the pool.
	tasksChan    chan Task          // Channel to send tasks to workers.
	resultsChan  chan Result        // Channel to receive results from workers.
	wg           sync.WaitGroup     // WaitGroup to wait for all workers to finish.
	ctx          context.Context    // Context to control cancellation of worker operations.
	cancel       context.CancelFunc // Function to cancel the context and stop the workers.
	mu           sync.Mutex         // Mutex to protect shared resources.
	stopped      bool               // Flag to indicate if the worker pool has been stopped.
	resultsWg    sync.WaitGroup     // WaitGroup to track results processing
}

// NewWorkerPool creates and returns a new WorkerPool with the specified number of workers and buffer size.
func NewWorkerPool(workersCount, bufferSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background()) // Creating a cancelable context for managing worker lifecycle.
	return &WorkerPool{
		workersCount: workersCount,
		tasksChan:    make(chan Task, bufferSize),   // Buffer to store tasks until workers can process them.
		resultsChan:  make(chan Result, bufferSize), // Buffer to store results of task executions.
		ctx:          ctx,                           // Assign the created context.
		cancel:       cancel,                        // Assign the cancel function to stop the workers.
	}
}

// a worker processes tasks from tasksChan and sends results to resultsChan.
func (wp *WorkerPool) worker() {
	defer wp.wg.Done() // Ensure the WaitGroup counter is decremented when the worker finishes.

	for {
		select {
		case <-wp.ctx.Done(): // If the context is canceled, stop the worker.
			return
		case task, ok := <-wp.tasksChan: // Get a task from the tasks channel.
			if !ok { // If the tasks channel is closed, stop the worker.
				return
			}
			// Execute the task and get the result.
			output, err := task.Execute()
			result := Result{
				Output: output,
				Err:    err,
				Task:   task,
			}

			// Increment result counter before sending
			wp.resultsWg.Add(1)

			select {
			case <-wp.ctx.Done(): // If the context is canceled during result handling, stop.
				wp.resultsWg.Done() // Decrement since we won't send this result
				return
			case wp.resultsChan <- result: // Send the result to the results channel.
			}
		}
	}
}

// Start launches the worker goroutines to process tasks concurrently.
func (wp *WorkerPool) Start() {
	// Start the worker goroutines.
	for i := 0; i < wp.workersCount; i++ {
		wp.wg.Add(1)   // Add to the WaitGroup to track the number of workers.
		go wp.worker() // Launch each worker in a new goroutine.
	}
}

// Submit adds a task to the tasks channel to be processed by workers.
func (wp *WorkerPool) Submit(task Task) bool {
	wp.mu.Lock()
	defer wp.mu.Unlock() // Lock to ensure thread-safe access to the worker pool state.

	if wp.stopped {
		return false // If the pool is stopped, do not accept new tasks.
	}

	select {
	case <-wp.ctx.Done(): // If the context is canceled, return false to indicate failure.
		return false
	case wp.tasksChan <- task: // If the task channel is not full, submit the task.
		return true
	}
}

// Results returns a read-only channel for receiving the results of processed tasks.
func (wp *WorkerPool) Results() <-chan Result {
	return wp.resultsChan
}

// markResultProcessed decrements the resultsWg counter to track processed results
func (wp *WorkerPool) markResultProcessed() {
	wp.resultsWg.Done()
}

// Stop stops the worker pool, cancels the context, and waits for all workers to finish.
func (wp *WorkerPool) Stop() {
	wp.mu.Lock()
	if wp.stopped {
		wp.mu.Unlock()
		return
	}
	wp.stopped = true
	close(wp.tasksChan) // Close task channel to signal workers to stop
	wp.mu.Unlock()

	// Use a timeout for waiting for workers to finish
	done := make(chan struct{})
	go func() {
		wp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Workers finished gracefully
		log.Println("All workers finished gracefully")
	case <-time.After(5 * time.Second):
		log.Println("WARNING: Timed out waiting for workers to finish")
		wp.cancel() // Force cancel if timed out
	}

	// Wait for all results to be processed with timeout
	resultsDone := make(chan struct{})
	go func() {
		wp.resultsWg.Wait()
		close(resultsDone)
	}()

	select {
	case <-resultsDone:
		// All results processed
		log.Println("All results processed")
	case <-time.After(5 * time.Second):
		log.Println("WARNING: Timed out waiting for results to be processed")
	}

	// Only now close the results channel
	close(wp.resultsChan)
	wp.cancel() // Ensure context is canceled
}

// ProcessResult Helper function to mark a result as processed
// This should be called by the consumer after processing each result
func ProcessResult(wp *WorkerPool, result Result) {
	wp.markResultProcessed()
}
