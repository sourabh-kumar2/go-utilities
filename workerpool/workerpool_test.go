package workerpool

import (
	"errors"
	"sync"
	"testing"
	"time"
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

// 1. Basic Task Submission and Execution
func TestWorkerPool_BasicTaskSubmission(t *testing.T) {
	pool := NewWorkerPool(3, 5)
	pool.Start()
	defer pool.Stop()

	task := &MockTask{id: 1, shouldFail: false}
	if !pool.Submit(task) {
		t.Errorf("Failed to submit task %d", task.id)
	}

	result := <-pool.Results()
	if result.Err != nil || result.Output != task.id {
		t.Errorf("Expected output %d, got %v (error: %v)", task.id, result.Output, result.Err)
	}
}

// 2. Task Execution Failure
func TestWorkerPool_TaskExecutionFailure(t *testing.T) {
	pool := NewWorkerPool(3, 5)
	pool.Start()
	defer pool.Stop()

	task := &MockTask{id: 1, shouldFail: true}
	if !pool.Submit(task) {
		t.Errorf("Failed to submit task %d", task.id)
	}

	result := <-pool.Results()
	if result.Err == nil {
		t.Errorf("Expected error for task %d, but got none", task.id)
	}
}

// 3. Pool Shutdown (Context Cancelation)
func TestWorkerPool_PoolShutdown(t *testing.T) {
	pool := NewWorkerPool(3, 5)
	pool.Start()

	task := &MockTask{id: 1, shouldFail: false}
	if !pool.Submit(task) {
		t.Errorf("Failed to submit task %d", task.id)
	}

	pool.Stop()

	if pool.Submit(&MockTask{id: 2, shouldFail: false}) {
		t.Errorf("Task submitted after pool shutdown")
	}
}

// 4. Multiple Tasks Submission
func TestWorkerPool_MultipleTasksSubmission(t *testing.T) {
	pool := NewWorkerPool(3, 5)
	pool.Start()
	defer pool.Stop()

	taskCount := 10
	for i := 0; i < taskCount; i++ {
		task := &MockTask{id: i, shouldFail: false}
		if !pool.Submit(task) {
			t.Errorf("Failed to submit task %d", i)
		}
	}

	results := make([]Result, 0)
	for i := 0; i < taskCount; i++ {
		results = append(results, <-pool.Results())
	}

	if len(results) != taskCount {
		t.Errorf("Expected %d results, got %d", taskCount, len(results))
	}
}

// 5. Submit Task After Pool Shutdown
func TestWorkerPool_SubmitAfterShutdown(t *testing.T) {
	pool := NewWorkerPool(3, 5)
	pool.Start()
	pool.Stop()

	task := &MockTask{id: 1, shouldFail: false}
	if pool.Submit(task) {
		t.Errorf("Task submitted after pool shutdown")
	}
}

// 6. Concurrency Test
func TestWorkerPool_Concurrency(t *testing.T) {
	pool := NewWorkerPool(3, 5)
	pool.Start()
	defer pool.Stop()

	var wg sync.WaitGroup
	taskCount := 10
	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			task := &MockTask{id: id, shouldFail: false}
			if !pool.Submit(task) {
				t.Errorf("Failed to submit task %d", id)
			}
		}(i)
	}
	wg.Wait()

	results := make([]Result, 0)
	for i := 0; i < taskCount; i++ {
		results = append(results, <-pool.Results())
	}

	if len(results) != taskCount {
		t.Errorf("Expected %d results, got %d", taskCount, len(results))
	}
}

// 7. Buffer Overflow (Channel Capacity)
func TestWorkerPool_BufferOverflow(t *testing.T) {
	pool := NewWorkerPool(3, 2) // Small buffer size
	pool.Start()
	defer pool.Stop()

	taskCount := 5
	for i := 0; i < taskCount; i++ {
		task := &MockTask{id: i, shouldFail: false}
		if !pool.Submit(task) {
			t.Logf("Task %d rejected due to buffer overflow", i)
		}
	}
}

// 8. Pool Capacity with Multiple Workers
func TestWorkerPool_MultipleWorkers(t *testing.T) {
	pool := NewWorkerPool(5, 10)
	pool.Start()
	defer pool.Stop()

	taskCount := 20
	for i := 0; i < taskCount; i++ {
		task := &MockTask{id: i, shouldFail: false}
		if !pool.Submit(task) {
			t.Errorf("Failed to submit task %d", i)
		}
	}

	results := make([]Result, 0)
	for i := 0; i < taskCount; i++ {
		results = append(results, <-pool.Results())
	}

	if len(results) != taskCount {
		t.Errorf("Expected %d results, got %d", taskCount, len(results))
	}
}

// 9. Task Submission Before Shutdown
func TestWorkerPool_SubmitBeforeShutdown(t *testing.T) {
	pool := NewWorkerPool(3, 5)
	pool.Start()

	task := &MockTask{id: 1, shouldFail: false}
	if !pool.Submit(task) {
		t.Errorf("Failed to submit task %d", task.id)
	}

	var wg sync.WaitGroup

	results := make([]Result, 0)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for result := range pool.Results() {
			results = append(results, result)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	pool.Stop()
	wg.Wait()

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
		return
	}

	result := results[0]
	if result.Err != nil || result.Output != task.id {
		t.Errorf("Expected output %d, got %v (error: %v)", task.id, result.Output, result.Err)
	}
}

// 10. Worker Pool Stop Mechanism
func TestWorkerPool_StopMechanism(t *testing.T) {
	pool := NewWorkerPool(3, 5)
	pool.Start()

	task := &MockTask{id: 1, shouldFail: false}
	if !pool.Submit(task) {
		t.Errorf("Failed to submit task %d", task.id)
	}

	pool.Stop()

	if pool.Submit(&MockTask{id: 2, shouldFail: false}) {
		t.Errorf("Task submitted after pool shutdown")
	}
}
