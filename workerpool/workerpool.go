package workerpool

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
