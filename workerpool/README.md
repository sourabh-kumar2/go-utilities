# WorkerPool

The `WorkerPool` package provides a simple and efficient implementation of a worker pool in Go. It allows you to process tasks concurrently using a fixed number of workers, making it ideal for managing workloads in a controlled and scalable manner.

## Features

- **Task Submission**: Submit tasks to be processed by the worker pool.
- **Concurrency**: Process tasks concurrently using multiple workers.
- **Graceful Shutdown**: Stop the worker pool and ensure all tasks are processed before shutdown.
- **Result Handling**: Collect results of task execution, including errors.
- **Thread-Safe**: Designed to handle concurrent task submissions safely.

## Installation

To use the `WorkerPool` package, add it to your Go project:

```bash
go get github.com/sourabh-kumar2/go-utilities/workerpool
```

## Usage

### Import the Package

```go
import "github.com/sourabh-kumar2/go-utilities/workerpool"
```

### Define a Task

Implement the `Task` interface for your tasks:

```go
type MyTask struct {
    ID int
}

func (t *MyTask) Execute() (interface{}, error) {
    // Task logic here
    return t.ID, nil
}
```

### Create and Use the WorkerPool

```go
package main

import (
    "fmt"
    "github.com/sourabh-kumar2/go-utilities/workerpool"
)

func main() {
    // Create a worker pool with 3 workers and a buffer size of 5
    pool := workerpool.NewWorkerPool(3, 5)
    pool.Start()

    // Submit tasks
    for i := 0; i < 10; i++ {
        task := &MyTask{ID: i}
        if !pool.Submit(task) {
            fmt.Printf("Failed to submit task %d\n", i)
        }
    }

    // Collect results
    go func() {
        pool.Stop()
    }()

    for result := range pool.Results() {
        if result.Err != nil {
            fmt.Printf("Task failed: %v\n", result.Err)
        } else {
            fmt.Printf("Task succeeded: %v\n", result.Output)
        }
    }
}
```

## API Reference

### `WorkerPool`

#### Methods

- **`NewWorkerPool(workersCount, bufferSize int) *WorkerPool`**  
  Creates a new worker pool with the specified number of workers and buffer size.

- **`Start()`**  
  Starts the worker pool and launches the worker goroutines.

- **`Submit(task Task) bool`**  
  Submits a task to the pool. Returns `false` if the pool is stopped or the task channel is full.

- **`Results() <-chan Result`**  
  Returns a read-only channel to receive task results.

- **`Stop()`**  
  Stops the worker pool, ensuring all tasks are processed before shutting down.

### `Task` Interface

- **`Execute() (interface{}, error)`**  
  Defines the logic for task execution. Returns the result or an error.

### `Result`

- **`Output interface{}`**: The output of the task execution.
- **`Err error`**: Any error that occurred during task execution.
- **`Task Task`**: The task that was executed.

## Testing

Run the tests using:

```bash
go test ./workerpool
```
