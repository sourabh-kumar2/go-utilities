package workerpool

import (
	"context"
	"sync"
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
