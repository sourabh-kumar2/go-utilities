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

// New creates a new GracefulShutdown instance with default signals
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

// AddShutdownFunc registers a function to be called during shutdown
func (gs *GracefulShutdown) AddShutdownFunc(fn func()) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	gs.cleanupFuncs = append(gs.cleanupFuncs, fn)
}

// Context returns the context which is canceled on signal
func (gs *GracefulShutdown) Context() context.Context {
	return gs.ctx
}

// Start begins listening for shutdown signals
func (gs *GracefulShutdown) Start() {
	signal.Notify(gs.signalCh, gs.signals...)

	go func() {
		sig := <-gs.signalCh
		log.Printf("Received signal: %v, initiating shutdown...", sig)
		gs.cleanup()
	}()
}

// cleanup runs all registered shutdown functions
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

// WaitForShutdown listens for the context cancellation and can be used in goroutines
func (gs *GracefulShutdown) WaitForShutdown() {
	<-gs.ctx.Done()
	log.Println("Graceful shutdown triggered, context canceled.")
}
