// Package gracefulshutdown provides a utility for managing graceful shutdowns
// in Go applications. It listens for termination signals, executes cleanup functions,
// and cancels the context to notify all dependent components to shut down gracefully.
package gracefulshutdown

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// GracefulShutdown manages graceful shutdown behavior
type GracefulShutdown struct {
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex

	signals      []os.Signal
	signalCh     chan os.Signal
	cleanupFuncs []func()
}

// New creates a new GracefulShutdown instance.
// It listens for the specified signals (defaults to SIGINT, SIGTERM).
func New(signals ...os.Signal) *GracefulShutdown {
	ctx, cancel := context.WithCancel(context.Background())

	if len(signals) == 0 {
		signals = []os.Signal{os.Interrupt, syscall.SIGTERM}
	}

	return &GracefulShutdown{
		ctx:      ctx,
		cancel:   cancel,
		signals:  signals,
		signalCh: make(chan os.Signal, 1),
	}
}

// AddCleanUpFunc registers a function to be called during shutdown.
// These functions will be executed in reverse order (last added first).
// You can add any cleanup logic such as closing database connections, stopping servers, etc.
func (gs *GracefulShutdown) AddCleanUpFunc(fn func()) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	gs.cleanupFuncs = append(gs.cleanupFuncs, fn)
}

// Context returns the context which is canceled when a termination signal is received.
// The context can be used to propagate shutdown signals to other parts of your application.
func (gs *GracefulShutdown) Context() context.Context {
	return gs.ctx
}

// Start begins listening for shutdown signals (SIGINT, SIGTERM, etc.)
// Once a signal is received, it initiates the shutdown process, runs registered cleanup functions.
func (gs *GracefulShutdown) Start() {
	signal.Notify(gs.signalCh, gs.signals...)

	go func() {
		sig := <-gs.signalCh
		log.Printf("Received signal: %v, initiating shutdown...", sig)
		gs.cleanup()
	}()
}

// cleanup runs all registered shutdown functions in reverse order.
// the context is canceled first to notify any context-aware components.
// after that, cleanup functions are executed and any panics are recovered from.
func (gs *GracefulShutdown) cleanup() {
	// Cancel context first to notify all context-aware components
	gs.cancel()

	log.Println("Running shutdown functions...")
	// Run cleanup functions in reverse order (last registered first)
	for i := len(gs.cleanupFuncs) - 1; i >= 0; i-- {
		// Recover from panics in cleanup functions
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Panic in shutdown function: %v", r)
				}
			}()
			gs.cleanupFuncs[i]()
		}()
	}

	log.Println("Graceful shutdown complete.")
	os.Exit(0)
}

// WaitForShutdown blocks until the context is canceled, indicating that a shutdown has been triggered.
// This can be used in goroutines to wait for the application to shut down gracefully.
func (gs *GracefulShutdown) WaitForShutdown() {
	<-gs.ctx.Done()
	log.Println("Graceful shutdown triggered, context canceled.")
}
