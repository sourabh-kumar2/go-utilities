# GracefulShutdown

The `gracefulshutdown` package provides a utility for managing graceful shutdowns in Go applications. It listens for termination signals, executes cleanup functions, and cancels the context to notify all dependent components to shut down gracefully.

## Features

- **Signal Handling**: Listens for termination signals (e.g., `SIGINT`, `SIGTERM`) to trigger shutdown.
- **Context Propagation**: Provides a context that is canceled when a shutdown is initiated.
- **Cleanup Functions**: Allows registering cleanup functions that are executed in reverse order during shutdown.
- **Panic Recovery**: Recovers from panics in cleanup functions to ensure all functions are executed.

## Installation

To use the `gracefulshutdown` package, add it to your Go project:

```bash
go get github.com/sourabh-kumar2/go-utilities/gracefulshutdown
```

## Usage

### Import the Package

```go
import "github.com/sourabh-kumar2/go-utilities/gracefulshutdown"
```

### Example

```go
package main

import (
	"fmt"
	"github.com/sourabh-kumar2/go-utilities/gracefulshutdown"
	"time"
)

func main() {
	// Create a new GracefulShutdown instance
	gs := gracefulshutdown.New()

	// Register cleanup functions
	gs.AddCleanUpFunc(func() {
		fmt.Println("Closing database connection...")
	})
	gs.AddCleanUpFunc(func() {
		fmt.Println("Stopping server...")
	})

	// Start listening for shutdown signals
	gs.Start()

	// Simulate application running
	fmt.Println("Application is running. Press Ctrl+C to exit.")
	time.Sleep(10 * time.Second)

	// Wait for shutdown to complete
	gs.WaitForShutdown()
	fmt.Println("Application has shut down gracefully.")
}
```

## API Reference

### `GracefulShutdown`

#### Methods

- **`New(signals ...os.Signal) *GracefulShutdown`**  
  Creates a new `GracefulShutdown` instance. Defaults to listening for `SIGINT` and `SIGTERM` if no signals are provided.

- **`AddCleanUpFunc(fn func())`**  
  Registers a cleanup function to be executed during shutdown.

- **`Context() context.Context`**  
  Returns a context that is canceled when a shutdown is initiated.

- **`Start()`**  
  Begins listening for termination signals and initiates the shutdown process when a signal is received.

- **`WaitForShutdown()`**  
  Blocks until the shutdown process is complete.

## Testing

To test the `gracefulshutdown` package, override the `exitFunc` to prevent `os.Exit` from being called during tests.

Run the tests using:

```bash
go test ./gracefulshutdown
```
