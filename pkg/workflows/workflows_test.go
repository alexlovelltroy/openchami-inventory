package workflows

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestWorkflow is a simple test workflow
type TestWorkflow struct {
	BaseWorkflow
	delay  time.Duration
	result string
	err    error
}

func NewTestWorkflow(name string, delay time.Duration, result string, err error) *TestWorkflow {
	return &TestWorkflow{
		BaseWorkflow: NewBaseWorkflow(name),
		delay:        delay,
		result:       result,
		err:          err,
	}
}

func (w *TestWorkflow) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	if w.delay > 0 {
		select {
		case <-time.After(w.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if w.err != nil {
		return nil, w.err
	}

	return w.result, nil
}

func TestNewGoWorkflowsManager(t *testing.T) {
	config := GoWorkflowsConfig{
		WorkerCount: 5,
	}

	mgr, err := NewGoWorkflowsManager(config)
	if err != nil {
		t.Fatalf("Failed to create workflow manager: %v", err)
	}
	defer mgr.Close()

	if mgr == nil {
		t.Fatal("Expected non-nil workflow manager")
	}

	if mgr.workers != 5 {
		t.Errorf("Expected 5 workers, got %d", mgr.workers)
	}
}

func TestNewGoWorkflowsManager_Defaults(t *testing.T) {
	config := GoWorkflowsConfig{}

	mgr, err := NewGoWorkflowsManager(config)
	if err != nil {
		t.Fatalf("Failed to create workflow manager: %v", err)
	}
	defer mgr.Close()

	if mgr.workers != 10 {
		t.Errorf("Expected default 10 workers, got %d", mgr.workers)
	}
}

func TestGoWorkflowsManager_ExecuteWorkflow(t *testing.T) {
	mgr, err := NewGoWorkflowsManager(GoWorkflowsConfig{})
	if err != nil {
		t.Fatalf("Failed to create workflow manager: %v", err)
	}
	defer mgr.Close()

	workflow := NewTestWorkflow("test", 0, "success", nil)

	execution, err := mgr.ExecuteWorkflow(context.Background(), workflow, nil)
	if err != nil {
		t.Fatalf("Failed to execute workflow: %v", err)
	}

	if execution.ID() == "" {
		t.Error("Expected non-empty execution ID")
	}

	// Wait for result
	result, err := execution.Result()
	if err != nil {
		t.Fatalf("Workflow failed: %v", err)
	}

	if result != "success" {
		t.Errorf("Expected result 'success', got %v", result)
	}

	if execution.Status() != StatusCompleted {
		t.Errorf("Expected status Completed, got %s", execution.Status())
	}
}

func TestGoWorkflowsManager_ExecuteWorkflowError(t *testing.T) {
	mgr, err := NewGoWorkflowsManager(GoWorkflowsConfig{})
	if err != nil {
		t.Fatalf("Failed to create workflow manager: %v", err)
	}
	defer mgr.Close()

	expectedErr := errors.New("workflow error")
	workflow := NewTestWorkflow("test", 0, "", expectedErr)

	execution, err := mgr.ExecuteWorkflow(context.Background(), workflow, nil)
	if err != nil {
		t.Fatalf("Failed to execute workflow: %v", err)
	}

	// Wait for result
	_, err = execution.Result()
	if err == nil {
		t.Fatal("Expected workflow to fail")
	}

	if err.Error() != expectedErr.Error() {
		t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
	}

	if execution.Status() != StatusFailed {
		t.Errorf("Expected status Failed, got %s", execution.Status())
	}
}

func TestGoWorkflowsManager_CancelWorkflow(t *testing.T) {
	mgr, err := NewGoWorkflowsManager(GoWorkflowsConfig{})
	if err != nil {
		t.Fatalf("Failed to create workflow manager: %v", err)
	}
	defer mgr.Close()

	// Create workflow with delay
	workflow := NewTestWorkflow("test", 5*time.Second, "success", nil)

	execution, err := mgr.ExecuteWorkflow(context.Background(), workflow, nil)
	if err != nil {
		t.Fatalf("Failed to execute workflow: %v", err)
	}

	// Cancel workflow
	if err := execution.Cancel(); err != nil {
		t.Fatalf("Failed to cancel workflow: %v", err)
	}

	// Wait for result (should be cancelled)
	_, err = execution.Result()
	if err == nil {
		t.Fatal("Expected workflow to be cancelled")
	}

	if execution.Status() != StatusCancelled {
		t.Errorf("Expected status Cancelled, got %s", execution.Status())
	}
}

func TestGoWorkflowsManager_GetExecution(t *testing.T) {
	mgr, err := NewGoWorkflowsManager(GoWorkflowsConfig{})
	if err != nil {
		t.Fatalf("Failed to create workflow manager: %v", err)
	}
	defer mgr.Close()

	workflow := NewTestWorkflow("test", 100*time.Millisecond, "success", nil)

	execution, err := mgr.ExecuteWorkflow(context.Background(), workflow, nil)
	if err != nil {
		t.Fatalf("Failed to execute workflow: %v", err)
	}

	// Get execution by ID
	retrieved, err := mgr.GetExecution(context.Background(), execution.ID())
	if err != nil {
		t.Fatalf("Failed to get execution: %v", err)
	}

	if retrieved.ID() != execution.ID() {
		t.Errorf("Expected execution ID %s, got %s", execution.ID(), retrieved.ID())
	}

	// Wait for completion
	execution.Result()
}

func TestGoWorkflowsManager_GetExecution_NotFound(t *testing.T) {
	mgr, err := NewGoWorkflowsManager(GoWorkflowsConfig{})
	if err != nil {
		t.Fatalf("Failed to create workflow manager: %v", err)
	}
	defer mgr.Close()

	_, err = mgr.GetExecution(context.Background(), "non-existent")
	if err == nil {
		t.Fatal("Expected error for non-existent execution")
	}
}

func TestGoWorkflowsManager_ConcurrentWorkflows(t *testing.T) {
	mgr, err := NewGoWorkflowsManager(GoWorkflowsConfig{
		WorkerCount: 5,
	})
	if err != nil {
		t.Fatalf("Failed to create workflow manager: %v", err)
	}
	defer mgr.Close()

	// Execute multiple workflows concurrently
	count := 10
	executions := make([]WorkflowExecution, count)

	for i := 0; i < count; i++ {
		workflow := NewTestWorkflow("test", 10*time.Millisecond, "success", nil)
		execution, err := mgr.ExecuteWorkflow(context.Background(), workflow, nil)
		if err != nil {
			t.Fatalf("Failed to execute workflow %d: %v", i, err)
		}
		executions[i] = execution
	}

	// Wait for all to complete
	for i, execution := range executions {
		result, err := execution.Result()
		if err != nil {
			t.Errorf("Workflow %d failed: %v", i, err)
		}
		if result != "success" {
			t.Errorf("Workflow %d: expected 'success', got %v", i, result)
		}
	}
}

func TestNewWorkflowManager(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			name: "go-workflows",
			config: Config{
				Engine: "go-workflows",
				GoWorkflows: GoWorkflowsConfig{
					WorkerCount: 5,
				},
			},
			expectError: false,
		},
		{
			name: "temporal (not implemented)",
			config: Config{
				Engine: "temporal",
				Temporal: TemporalConfig{
					HostPort: "localhost:7233",
				},
			},
			expectError: true, // Expected since Temporal is not fully implemented
		},
		{
			name: "unknown engine",
			config: Config{
				Engine: "unknown",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := NewWorkflowManager(tt.config)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if mgr != nil {
					mgr.Close()
				}
			}
		})
	}
}

func TestExecutionStatus(t *testing.T) {
	tests := []struct {
		status     ExecutionStatus
		isTerminal bool
	}{
		{StatusRunning, false},
		{StatusCompleted, true},
		{StatusFailed, true},
		{StatusCancelled, true},
	}

	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			if tt.status.IsTerminal() != tt.isTerminal {
				t.Errorf("Expected IsTerminal() = %v for status %s", tt.isTerminal, tt.status)
			}
		})
	}
}
