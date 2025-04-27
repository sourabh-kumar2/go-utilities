package workerpool

import (
	"reflect"
	"testing"
)

func TestNewWorkerPool(t *testing.T) {
	workersCount := 5
	bufferSize := 10
	pool := NewWorkerPool(workersCount, bufferSize)

	// Check if the worker count is correct
	if pool.workersCount != workersCount {
		t.Errorf("Expected workersCount %d, but got %d", workersCount, pool.workersCount)
	}

	// Check if the tasks channel is correctly initialized
	if cap(pool.tasksChan) != bufferSize {
		t.Errorf("Expected tasks channel buffer size %d, but got %d", bufferSize, cap(pool.tasksChan))
	}

	// Check if the results channel is correctly initialized
	if cap(pool.resultsChan) != bufferSize {
		t.Errorf("Expected results channel buffer size %d, but got %d", bufferSize, cap(pool.resultsChan))
	}

	// Check if the context is not nil
	if pool.ctx == nil {
		t.Errorf("Expected a non-nil context")
	}

	// Check if cancel function is not nil
	if pool.cancel == nil {
		t.Errorf("Expected a non-nil cancel function")
	}

	// Check if the type of cancel function is correct
	cancelType := reflect.TypeOf(pool.cancel).String()
	if cancelType != "context.CancelFunc" {
		t.Errorf("Expected cancel function of type context.CancelFunc, but got %s", cancelType)
	}
}
