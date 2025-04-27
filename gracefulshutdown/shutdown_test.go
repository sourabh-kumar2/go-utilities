package gracefulshutdown

import (
	"syscall"
	"testing"
	"time"
)

func TestGracefulShutdown_New(t *testing.T) {
	gs := New(syscall.SIGINT, syscall.SIGTERM)

	if gs == nil {
		t.Fatal("Expected GracefulShutdown instance, got nil")
	}

	if len(gs.signals) != 2 {
		t.Errorf("Expected 2 signals, got %d", len(gs.signals))
	}
}

func TestGracefulShutdown_AddCleanUpFunc(t *testing.T) {
	gs := New()
	gs.exitFunc = func(int) {}
	cleanupCalled := false

	gs.AddCleanUpFunc(func() {
		cleanupCalled = true
	})

	if len(gs.cleanupFuncs) != 1 {
		t.Errorf("Expected 1 cleanup function, got %d", len(gs.cleanupFuncs))
	}

	gs.cleanup()
	if !cleanupCalled {
		t.Error("Expected cleanup function to be called, but it was not")
	}
}

func TestGracefulShutdown_Context(t *testing.T) {
	gs := New()
	ctx := gs.Context()

	if ctx == nil {
		t.Fatal("Expected context, got nil")
	}

	if ctx.Err() != nil {
		t.Errorf("Expected context to not be canceled, got error: %v", ctx.Err())
	}
}

func TestGracefulShutdown_StartAndCleanup(t *testing.T) {
	gs := New(syscall.SIGINT)
	gs.exitFunc = func(int) {}
	cleanupCalled := false

	gs.AddCleanUpFunc(func() {
		cleanupCalled = true
	})

	go gs.Start()

	// Simulate sending a signal
	time.Sleep(100 * time.Millisecond)
	gs.signalCh <- syscall.SIGINT

	// Allow time for cleanup to complete
	time.Sleep(100 * time.Millisecond)

	if !cleanupCalled {
		t.Error("Expected cleanup function to be called, but it was not")
	}
}

func TestGracefulShutdown_WaitForShutdown(t *testing.T) {
	gs := New()
	go func() {
		time.Sleep(100 * time.Millisecond)
		gs.cancel()
	}()

	done := make(chan struct{})
	go func() {
		gs.WaitForShutdown()
		close(done)
	}()

	select {
	case <-done:
		// Test passed
	case <-time.After(200 * time.Millisecond):
		t.Error("WaitForShutdown did not return as expected")
	}
}
