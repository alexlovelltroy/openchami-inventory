package workflows
// Package workflows provides workflow execution abstraction.
//
// This package abstracts workflow execution to support both embedded (go-workflows)
// and distributed (Temporal) workflow engines. This allows the system to start
// simple with embedded workflows and scale to distributed execution as needed.
//
// Implementations:
//   - GoWorkflowsManager: Embedded workflow engine using SQLite backend
//   - TemporalManager: Distributed workflow engine using Temporal
//
// Usage:
//
//	// Create workflow manager from config
//	mgr, err := workflows.NewWorkflowManager(viper.GetViper())
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer mgr.Close()
//
//	// Define a workflow
//	workflow := &MyWorkflow{...}
//
//	// Execute workflow
//	execution, err := mgr.ExecuteWorkflow(ctx, workflow, input)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Wait for result
//	result, err := execution.Result()
package workflows

import (
	"context"
	"fmt"
)

// WorkflowManager abstracts workflow execution.
//
// This interface allows the system to use different workflow engines
// (embedded or distributed) without changing application code.
type WorkflowManager interface {
	// ExecuteWorkflow starts a workflow execution
	//
	// Parameters:
	//   - ctx: Context for cancellation
	//   - workflow: Workflow definition to execute
	//   - input: Input data for the workflow
	//
	// Returns:
	//   - WorkflowExecution: Handle to the running workflow
	//   - error: If workflow failed to start
	ExecuteWorkflow(ctx context.Context, workflow Workflow, input interface{}) (WorkflowExecution, error)

	// GetExecution retrieves an existing workflow execution
	//
	// Parameters:
	//   - ctx: Context for cancellation
	//   - id: Workflow execution ID
	//
	// Returns:
	//   - WorkflowExecution: Handle to the workflow
	//   - error: If workflow not found
	GetExecution(ctx context.Context, id string) (WorkflowExecution, error)

	// CancelExecution cancels a running workflow
	//
	// Parameters:
	//   - ctx: Context for cancellation
	//   - id: Workflow execution ID
	//
	// Returns:
	//   - error: If cancellation failed
	CancelExecution(ctx context.Context, id string) error

	// Close shuts down the workflow manager
	//
	// This should be called when the application exits to clean up resources.
	Close() error
}

// Workflow represents a workflow definition.
//
// Implementations should be idempotent and handle retries gracefully.
type Workflow interface {
	// Name returns the workflow name
	//
	// This is used for identification and routing.
	Name() string

	// Execute runs the workflow logic
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - input: Input data (type-specific)
	//
	// Returns:
	//   - interface{}: Result data
	//   - error: If workflow failed
	Execute(ctx context.Context, input interface{}) (interface{}, error)
}

// WorkflowExecution represents a running workflow.
//
// This provides a handle to query status and results.
type WorkflowExecution interface {
	// ID returns the unique workflow execution ID
	ID() string

	// Status returns the current execution status
	Status() ExecutionStatus

	// Result blocks until the workflow completes and returns the result
	//
	// Returns:
	//   - interface{}: Workflow result
	//   - error: If workflow failed
	Result() (interface{}, error)

	// Cancel requests cancellation of the workflow
	Cancel() error
}

// ExecutionStatus represents the status of a workflow execution.
type ExecutionStatus string

const (
	// StatusRunning indicates the workflow is currently executing
	StatusRunning ExecutionStatus = "Running"

	// StatusCompleted indicates the workflow finished successfully
	StatusCompleted ExecutionStatus = "Completed"

	// StatusFailed indicates the workflow failed with an error
	StatusFailed ExecutionStatus = "Failed"

	// StatusCancelled indicates the workflow was cancelled
	StatusCancelled ExecutionStatus = "Cancelled"
)

// String returns the string representation of the status
func (s ExecutionStatus) String() string {
	return string(s)
}

// IsTerminal returns true if the status is terminal (completed, failed, or cancelled)
func (s ExecutionStatus) IsTerminal() bool {
	return s == StatusCompleted || s == StatusFailed || s == StatusCancelled
}

// Config holds workflow manager configuration
type Config struct {
	// Engine specifies which workflow engine to use ("go-workflows" or "temporal")
	Engine string

	// GoWorkflows configuration
	GoWorkflows GoWorkflowsConfig

	// Temporal configuration
	Temporal TemporalConfig
}

// GoWorkflowsConfig holds go-workflows specific configuration
type GoWorkflowsConfig struct {
	// DBPath is the path to the SQLite database file
	DBPath string

	// WorkerCount is the number of worker goroutines
	WorkerCount int

	// MaxWorkflowRuntime is the maximum time a workflow can run
	MaxWorkflowRuntime string
}

// TemporalConfig holds Temporal specific configuration
type TemporalConfig struct {
	// HostPort is the Temporal server address
	HostPort string

	// Namespace is the Temporal namespace to use
	Namespace string

	// TaskQueue is the task queue name
	TaskQueue string

	// WorkerCount is the number of worker goroutines
	WorkerCount int

	// MaxConcurrentWorkflows is the max number of concurrent workflows
	MaxConcurrentWorkflows int
}

// NewWorkflowManager creates a workflow manager from configuration.
//
// This factory function creates the appropriate workflow manager based
// on the engine specified in the configuration.
//
// Parameters:
//   - config: Workflow manager configuration
//
// Returns:
//   - WorkflowManager: Initialized workflow manager
//   - error: If initialization failed
//
// Example:
//
//	config := workflows.Config{
//	    Engine: "go-workflows",
//	    GoWorkflows: workflows.GoWorkflowsConfig{
//	        DBPath: "./workflows.db",
//	        WorkerCount: 10,
//	    },
//	}
//	mgr, err := workflows.NewWorkflowManager(config)
func NewWorkflowManager(config Config) (WorkflowManager, error) {
	switch config.Engine {
	case "go-workflows":
		return NewGoWorkflowsManager(config.GoWorkflows)
	case "temporal":
		return NewTemporalManager(config.Temporal)
	default:
		return nil, fmt.Errorf("unknown workflow engine: %s (supported: go-workflows, temporal)", config.Engine)
	}
}

// BaseWorkflow provides common functionality for workflows.
//
// Concrete workflows can embed this to get helper methods.
type BaseWorkflow struct {
	name string
}

// NewBaseWorkflow creates a new base workflow
func NewBaseWorkflow(name string) BaseWorkflow {
	return BaseWorkflow{name: name}
}

// Name returns the workflow name
func (w BaseWorkflow) Name() string {
	return w.name
}
