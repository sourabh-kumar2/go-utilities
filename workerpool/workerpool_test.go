package workerpool

import (
	"errors"
	"testing"
)

// MockTask is a mock implementation of the Task interface for testing.
type MockTask struct {
	id         int
	shouldFail bool
}

// Execute simulates task execution.
func (m *MockTask) Execute() (interface{}, error) {
	if m.shouldFail {
		return nil, errors.New("task failed")
	}
	return m.id, nil
}

func TestWorkerPool_BasicTaskSubmission(t *testing.T) {
	workersCount := 3
	bufferSize := 5
	pool := NewWorkerPool(workersCount, bufferSize)

	pool.Start()
	defer pool.Stop()

	// Submit a simple task
	task := &MockTask{id: 1, shouldFail: false}
	if !pool.Submit(task) {
		t.Errorf("Failed to submit task %d", task.id)
	}

	// Collect results
	result := <-pool.Results()

	// Verify the result
	if result.Err != nil {
		t.Errorf("Expected no error, but got: %v", result.Err)
	}
	if result.Output != task.id {
		t.Errorf("Expected output %d, but got %v", task.id, result.Output)
	}
}
