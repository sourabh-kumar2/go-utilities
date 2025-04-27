package workerpool

// Task defines an interface for tasks to be executed by workers in the pool.
type Task interface {
	Execute() (interface{}, error) // Executes the task and returns the result or an error.
}
