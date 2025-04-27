package gracefulshutdown

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

// GracefulShutdown manages graceful shutdown behavior
type GracefulShutdown struct {
	signals      []os.Signal
	signalCh     chan os.Signal
	cleanupFuncs []func()
}

// New creates a new GracefulShutdown instance with default signals
func New(signals ...os.Signal) *GracefulShutdown {
	if len(signals) == 0 {
		signals = []os.Signal{os.Interrupt, syscall.SIGTERM}
	}

	return &GracefulShutdown{
		signals:  signals,
		signalCh: make(chan os.Signal, 1),
	}
}

// AddShutdownFunc registers a function to be called during shutdown
func (gs *GracefulShutdown) AddShutdownFunc(fn func()) {
	gs.cleanupFuncs = append(gs.cleanupFuncs, fn)
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
	log.Println("Running shutdown functions...")
	for _, fn := range gs.cleanupFuncs {
		fn()
	}
	log.Println("Graceful shutdown complete.")
	os.Exit(0)
}
