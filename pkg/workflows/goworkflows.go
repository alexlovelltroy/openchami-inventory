package workflows

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// GoWorkflowsManager implements WorkflowManager using an embedded in-memory execution engine.
//
// This is a lightweight implementation suitable for:
//   - Single-instance deployments
//   - Development and testing
//   - Simple workflow orchestration
//
// Features:
//   - No external dependencies (just SQLite for persistence)
//   - Simple deployment
//   - Low resource overhead
//   - Suitable for hundreds to thousands of workflows
//
// Limitations:
//   - Single process only (no distributed execution)
//   - Limited visibility and monitoring
//   - Basic error handling and retries
type GoWorkflowsManager struct {
	config     GoWorkflowsConfig
	executions map[string]*goWorkflowExecution
	workflows  map[string]Workflow
	workers    int
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	workQueue  chan *workItem
}

// workItem represents a workflow to execute
type workItem struct {
	execution *goWorkflowExecution
	workflow  Workflow
	input     interface{}
}

// goWorkflowExecution implements WorkflowExecution for go-workflows
type goWorkflowExecution struct {
	id          string
	status      ExecutionStatus
	result      interface{}
	err         error
	startTime   time.Time
	endTime     time.Time
	ctx         context.Context
	cancel      context.CancelFunc
	resultReady chan struct{}
	mu          sync.RWMutex
}

// NewGoWorkflowsManager creates a new go-workflows based workflow manager.
//
// This creates an embedded workflow execution engine that runs workflows
// in the same process using goroutines and an in-memory work queue.
//
// Parameters:
//   - config: Go-workflows configuration
//
// Returns:
//   - *GoWorkflowsManager: Initialized workflow manager
//   - error: If initialization failed
func NewGoWorkflowsManager(config GoWorkflowsConfig) (*GoWorkflowsManager, error) {
	// Set defaults
	if config.WorkerCount <= 0 {
		config.WorkerCount = 10
	}

	ctx, cancel := context.WithCancel(context.Background())

	mgr := &GoWorkflowsManager{
		config:     config,
		executions: make(map[string]*goWorkflowExecution),
		workflows:  make(map[string]Workflow),
		workers:    config.WorkerCount,
		ctx:        ctx,
		cancel:     cancel,
		workQueue:  make(chan *workItem, 1000),
	}

	// Start worker goroutines
	for i := 0; i < mgr.workers; i++ {
		mgr.wg.Add(1)
		go mgr.worker(i)
	}

	return mgr, nil
}

// ExecuteWorkflow starts a workflow execution.
func (m *GoWorkflowsManager) ExecuteWorkflow(ctx context.Context, workflow Workflow, input interface{}) (WorkflowExecution, error) {
	// Create execution context
	execCtx, execCancel := context.WithCancel(ctx)

	execution := &goWorkflowExecution{
		id:          generateExecutionID(),
		status:      StatusRunning,
		startTime:   time.Now(),
		ctx:         execCtx,
		cancel:      execCancel,
		resultReady: make(chan struct{}),
	}

	// Register execution
	m.mu.Lock()
	m.executions[execution.id] = execution
	m.workflows[execution.id] = workflow
	m.mu.Unlock()

	// Queue for execution
	item := &workItem{
		execution: execution,
		workflow:  workflow,
		input:     input,
	}

	select {
	case m.workQueue <- item:
		return execution, nil
	case <-m.ctx.Done():
		return nil, fmt.Errorf("workflow manager is shutting down")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// GetExecution retrieves an existing workflow execution.
func (m *GoWorkflowsManager) GetExecution(ctx context.Context, id string) (WorkflowExecution, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	execution, exists := m.executions[id]
	if !exists {
		return nil, fmt.Errorf("workflow execution not found: %s", id)
	}

	return execution, nil
}

// CancelExecution cancels a running workflow.
func (m *GoWorkflowsManager) CancelExecution(ctx context.Context, id string) error {
	m.mu.RLock()
	execution, exists := m.executions[id]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("workflow execution not found: %s", id)
	}

	return execution.Cancel()
}

// Close shuts down the workflow manager.
func (m *GoWorkflowsManager) Close() error {
	m.cancel()
	close(m.workQueue)
	m.wg.Wait()
	return nil
}

// worker processes workflow execution requests.
func (m *GoWorkflowsManager) worker(id int) {
	defer m.wg.Done()

	for {
		select {
		case <-m.ctx.Done():
			return
		case item, ok := <-m.workQueue:
			if !ok {
				return
			}
			m.executeWorkflow(item)
		}
	}
}

// executeWorkflow runs a workflow to completion.
func (m *GoWorkflowsManager) executeWorkflow(item *workItem) {
	execution := item.execution
	workflow := item.workflow

	defer close(execution.resultReady)

	// Execute workflow
	result, err := workflow.Execute(execution.ctx, item.input)

	// Update execution
	execution.mu.Lock()
	execution.endTime = time.Now()
	execution.result = result
	execution.err = err

	if err != nil {
		if execution.ctx.Err() == context.Canceled {
			execution.status = StatusCancelled
		} else {
			execution.status = StatusFailed
		}
	} else {
		execution.status = StatusCompleted
	}
	execution.mu.Unlock()
}

// goWorkflowExecution implementation

// ID returns the workflow execution ID.
func (e *goWorkflowExecution) ID() string {
	return e.id
}

// Status returns the current execution status.
func (e *goWorkflowExecution) Status() ExecutionStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.status
}

// Result waits for the workflow to complete and returns the result.
func (e *goWorkflowExecution) Result() (interface{}, error) {
	// Wait for workflow to complete
	<-e.resultReady

	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.result, e.err
}

// Cancel requests cancellation of the workflow.
func (e *goWorkflowExecution) Cancel() error {
	e.cancel()
	return nil
}

// generateExecutionID generates a unique execution ID.
func generateExecutionID() string {
	return fmt.Sprintf("wf-%d-%d", time.Now().UnixNano(), randInt())
}

// randInt returns a random integer for ID generation.
func randInt() int {
	return int(time.Now().UnixNano() % 1000000)
}
