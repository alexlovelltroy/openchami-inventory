# Reconciliation & Events System

Complete guide to the event-driven reconciliation system in OpenCHAMI Inventory.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Event System](#event-system)
- [Reconciliation Framework](#reconciliation-framework)
- [Workflow Engine](#workflow-engine)
- [Developing Reconcilers](#developing-reconcilers)
- [Event Handlers](#event-handlers)
- [Testing](#testing)
- [Best Practices](#best-practices)

## Overview

The OpenCHAMI Inventory system uses an **event-driven reconciliation pattern** to manage infrastructure resources. This enables:

- **Declarative Infrastructure** - Define desired state, system converges to it
- **Reactive Behavior** - Resources react to changes in other resources
- **Eventual Consistency** - Changes propagate through the system automatically
- **Complex Workflows** - Multi-step operations with rollback and retry
- **Observable Operations** - Track what the system is doing and why

### Key Concepts

**Events**: Notifications about state changes (e.g., "BMC created", "Node provisioned")

**Reconciliation**: Process of aligning actual state with desired state

**Workflows**: Multi-step operations that coordinate resource changes

**Reconcilers**: Controllers that implement reconciliation logic for resource types

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      REST API Layer                          │
│              (Receives user requests)                        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     Storage Layer                            │
│              (Saves resources, emits events)                 │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│            Event Bus (In-Memory/Distributed)                 │
│  • CloudEvents-compliant event structure                     │
│  • Wildcard pattern subscriptions                            │
│  • Worker pool for async event processing                    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              Reconciliation Controller                       │
│  • Watches resource change events                            │
│  • Manages work queue with deduplication                     │
│  • Dispatches to registered reconcilers                      │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              Resource Reconcilers                            │
│  • BMCReconciler - Manages BMC connections                   │
│  • FRUReconciler - Discovers hardware inventory              │
│  • NodeReconciler - Monitors node health                     │
│  • BootConfigReconciler - Manages boot configuration         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│            Workflow Engine (Optional)                        │
│  • Complex multi-step operations                             │
│  • Automatic retry and error handling                        │
│  • Status tracking and cancellation                          │
└─────────────────────────────────────────────────────────────┘
```

## Event System

### CloudEvents Standard

Events follow the [CloudEvents](https://cloudevents.io/) specification:

```go
type Event interface {
    ID() string              // Unique event identifier
    Type() string            // Event type (e.g., "bmcs.created")
    Source() string          // Event source
    Time() time.Time         // When event occurred
    Data() interface{}       // Event payload
    ResourceUID() string     // Associated resource UID
    ResourceKind() string    // Resource type
}
```

### Event Types

Standard event types follow the pattern: `io.openchami.inventory.{resource}.{action}`

| Event Type | Emitted When | Example |
|------------|--------------|---------|
| `*.created` | Resource created | `io.openchami.inventory.bmcs.created` |
| `*.updated` | Resource updated | `io.openchami.inventory.nodes.updated` |
| `*.deleted` | Resource deleted | `io.openchami.inventory.frus.deleted` |
| `*.reconciled` | Reconciliation succeeded | `io.openchami.inventory.bmcs.reconciled` |
| `*.error` | Reconciliation failed | `io.openchami.inventory.nodes.error` |

Custom events for domain-specific actions:
- `io.openchami.inventory.bmcs.connected` - BMC connection established
- `io.openchami.inventory.bmcs.disconnected` - BMC connection lost
- `io.openchami.inventory.nodes.provisioned` - Node provisioning complete
- `io.openchami.inventory.frus.discovered` - New FRUs discovered

### Event Bus

The event bus provides pub/sub messaging with wildcard subscriptions:

```go
// Create event bus with worker pool
eventBus := events.NewInMemoryEventBus(10) // 10 workers
eventBus.Start()
defer eventBus.Stop()

// Subscribe to specific events
eventBus.Subscribe("io.openchami.inventory.bmcs.created", handleBMCCreated)

// Subscribe with wildcards
eventBus.Subscribe("io.openchami.inventory.bmcs.*", handleAnyBMCEvent)
eventBus.Subscribe("io.openchami.inventory.*.created", handleAnyCreateEvent)

// Publish events
event := events.NewResourceEvent(
    "io.openchami.inventory.bmcs.created",
    bmc.UID,
    "BMC",
    map[string]interface{}{"bmc": bmc},
)
eventBus.Publish(ctx, event)
```

### Event Emission from Storage

Storage operations automatically emit events:

```go
func (s *Storage) SaveBMC(ctx context.Context, bmc *bmc.BMC) error {
    // Determine operation type
    exists := s.Exists(ctx, bmc.UID, "BMC")
    
    // Save resource
    if err := s.save(ctx, bmc); err != nil {
        return err
    }
    
    // Emit event
    eventType := "io.openchami.inventory.bmcs.created"
    if exists {
        eventType = "io.openchami.inventory.bmcs.updated"
    }
    
    event := events.NewResourceEvent(eventType, bmc.UID, "BMC", 
        map[string]interface{}{"bmc": bmc})
    return s.eventBus.Publish(ctx, event)
}
```

## Reconciliation Framework

### Reconciler Interface

Every resource type implements the `Reconciler` interface:

```go
type Reconciler interface {
    // Reconcile brings resource to desired state
    Reconcile(ctx context.Context, resource *resources.Resource) error
    
    // GetResourceKind returns the resource type this reconciler handles
    GetResourceKind() string
}
```

### Base Reconciler

`BaseReconciler` provides common functionality:

```go
type BaseReconciler struct {
    Client   ClientInterface  // Storage access
    EventBus events.EventBus  // Event publishing
    Logger   Logger           // Structured logging
}

// Helper methods
func (r *BaseReconciler) UpdateStatus(ctx, resource, status) error
func (r *BaseReconciler) EmitEvent(ctx, eventType, resource, data) error
func (r *BaseReconciler) SetCondition(resource, condition) error
```

### Reconciliation Controller

The controller manages the reconciliation loop:

```go
// Create controller
controller := reconcile.NewController(eventBus, storage)

// Register reconcilers
controller.RegisterReconciler(bmcReconciler)
controller.RegisterReconciler(fruReconciler)
controller.RegisterReconciler(nodeReconciler)

// Start controller
ctx := context.Background()
go controller.Start(ctx)
defer controller.Stop()
```

**Controller Features:**
- **Event-driven**: Reconciles when resources change
- **Periodic**: Reconciles all resources on schedule (default: 5 minutes)
- **Deduplication**: Prevents duplicate reconciliation requests
- **Rate limiting**: Exponential backoff for failing resources
- **Concurrent**: Processes multiple resources in parallel

### Work Queue

The work queue manages reconciliation requests:

```go
type WorkQueue interface {
    Add(request ReconcileRequest)           // Add to queue
    Get() (ReconcileRequest, bool)          // Get next item
    Done(request ReconcileRequest)          // Mark complete
    Shutdown()                              // Stop processing
}
```

**Features:**
- Deduplication - Only one request per resource at a time
- Rate limiting - Backs off on failures
- Priority - Recent changes processed first
- Thread-safe - Concurrent access from multiple goroutines

## Workflow Engine

For complex multi-step operations, use the workflow engine:

```go
// Create workflow manager
workflowMgr := workflows.NewGoWorkflowsManager(10) // 10 workers
defer workflowMgr.Close()

// Define workflow
workflow := func(ctx context.Context, input interface{}) (interface{}, error) {
    // Step 1: Connect to BMC
    bmc := input.(*bmc.BMC)
    client, err := connectBMC(ctx, bmc)
    if err != nil {
        return nil, err
    }
    
    // Step 2: Discover FRUs
    frus, err := discoverFRUs(ctx, client)
    if err != nil {
        return nil, err
    }
    
    // Step 3: Create FRU resources
    for _, fru := range frus {
        if err := createFRU(ctx, fru); err != nil {
            return nil, err
        }
    }
    
    return len(frus), nil
}

// Execute workflow
execution, err := workflowMgr.ExecuteWorkflow(ctx, "discover-frus", workflow, bmc)
if err != nil {
    log.Fatal(err)
}

// Check status
status := execution.Status()
fmt.Printf("Workflow %s: %s\n", execution.ID(), status.State)
```

**Workflow Features:**
- Automatic retry on failures
- Cancellation support
- Status tracking
- Result capture
- Error handling

## Developing Reconcilers

### Generate Reconciler Boilerplate

```bash
# Generate reconcilers for all resources
go run cmd/codegen/main.go -type reconcile -output pkg/reconcilers -package reconcilers
```

This creates:
- `{resource}_reconciler_generated.go` - Reconciler for each resource
- `registration_generated.go` - Automatic registration
- `event_handlers_generated.go` - Event handler registry

### Implement Reconciliation Logic

Edit the generated stub method:

```go
// File: pkg/reconcilers/bmc_reconciler_generated.go

func (r *BMCReconciler) reconcileBMC(ctx context.Context, bmc *bmc.BMC) error {
    // 1. Connect to BMC
    client, err := redfish.Connect(bmc.Spec.Endpoint, bmc.Spec.Credentials)
    if err != nil {
        return fmt.Errorf("failed to connect: %w", err)
    }
    defer client.Logout()
    
    // 2. Query BMC status
    system, err := client.GetSystem()
    if err != nil {
        return fmt.Errorf("failed to get system: %w", err)
    }
    
    // 3. Update status
    bmc.Status.Connected = true
    bmc.Status.PowerState = system.PowerState
    bmc.Status.Health = system.Health
    bmc.Status.LastSeen = time.Now()
    
    // Status is automatically saved and events emitted by framework
    return nil
}
```

### Reconciler Lifecycle

1. **Trigger**: Event or periodic timer triggers reconciliation
2. **Dequeue**: Controller gets request from work queue
3. **Dispatch**: Controller calls reconciler's `Reconcile()` method
4. **Execute**: Reconciler observes actual state and takes actions
5. **Update**: Reconciler updates resource status
6. **Emit**: Framework emits success/error events
7. **Requeue**: On error, request is rate-limited and requeued

### Error Handling

```go
func (r *BMCReconciler) reconcileBMC(ctx context.Context, bmc *bmc.BMC) error {
    // Transient errors - will be retried
    if err := r.connect(bmc); err != nil {
        return fmt.Errorf("connection failed: %w", err)
    }
    
    // Permanent errors - set condition and don't retry
    if !r.isValid(bmc) {
        r.SetCondition(&bmc.Resource, resources.Condition{
            Type:    "Invalid",
            Status:  "True",
            Reason:  "ValidationFailed",
            Message: "BMC endpoint is invalid",
        })
        return nil // Don't retry
    }
    
    return nil
}
```

## Event Handlers

### Cross-Resource Reactions

Event handlers enable one resource type to react to another:

```go
// File: pkg/reconcilers/event_handlers_generated.go

func (r *EventHandlerRegistry) handleBMCConnected(ctx context.Context, event events.Event) error {
    bmcUID := event.ResourceUID()
    
    // Get BMC resource
    resource, err := r.client.Get(ctx, bmcUID, "BMC")
    if err != nil {
        return err
    }
    bmc := resource.(*bmc.BMC)
    
    // Connect and discover FRUs
    client, err := redfish.Connect(bmc.Spec.Endpoint, bmc.Spec.Credentials)
    if err != nil {
        return err
    }
    defer client.Logout()
    
    frus, err := client.ListFRUs()
    if err != nil {
        return err
    }
    
    // Create FRU resources
    for _, fruData := range frus {
        fru := &fru.FRU{
            Resource: resources.Resource{
                Kind: "FRU",
                UID:  fmt.Sprintf("%s-fru-%s", bmc.UID, fruData.Id),
            },
            Spec: fru.FRUSpec{
                BMCUID: bmc.UID,
                FRUId:  fruData.Id,
            },
        }
        
        if err := r.client.Create(ctx, fru); err != nil {
            r.logger.Errorf("Failed to create FRU: %v", err)
            continue
        }
    }
    
    return nil
}
```

### Register Event Handlers

```go
func (r *EventHandlerRegistry) RegisterEventHandlers(eventBus events.EventBus) error {
    // BMC events
    eventBus.Subscribe("io.openchami.inventory.bmcs.connected", r.handleBMCConnected)
    eventBus.Subscribe("io.openchami.inventory.bmcs.disconnected", r.handleBMCDisconnected)
    
    // Node events
    eventBus.Subscribe("io.openchami.inventory.nodes.created", r.handleNodeCreated)
    
    return nil
}
```

## Testing

### Unit Testing Reconcilers

```go
func TestBMCReconciler(t *testing.T) {
    // Create mock components
    eventBus := events.NewInMemoryEventBus(1)
    eventBus.Start()
    defer eventBus.Stop()
    
    storage := &mockStorage{}
    reconciler := NewDefaultBMCReconciler(storage, eventBus)
    
    // Create test BMC
    bmc := &bmc.BMC{
        Resource: resources.Resource{
            Kind: "BMC",
            UID:  "test-bmc",
        },
        Spec: bmc.BMCSpec{
            Endpoint: "https://test.bmc.local",
        },
    }
    
    // Test reconciliation
    err := reconciler.Reconcile(context.Background(), &bmc.Resource)
    if err != nil {
        t.Errorf("Reconcile failed: %v", err)
    }
    
    // Verify status updated
    if !bmc.Status.Connected {
        t.Error("Expected BMC to be connected")
    }
}
```

### Integration Testing

```go
func TestReconciliationFlow(t *testing.T) {
    // Setup
    eventBus := events.NewInMemoryEventBus(10)
    eventBus.Start()
    defer eventBus.Stop()
    
    storage := createTestStorage(t)
    controller := reconcile.NewController(eventBus, storage)
    
    // Register reconcilers
    RegisterReconcilers(controller, storage, eventBus)
    
    // Start controller
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    go controller.Start(ctx)
    defer controller.Stop()
    
    // Create BMC (triggers reconciliation)
    bmc := createTestBMC()
    err := storage.SaveBMC(ctx, bmc)
    require.NoError(t, err)
    
    // Wait for reconciliation
    time.Sleep(500 * time.Millisecond)
    
    // Verify reconciliation occurred
    updated, err := storage.GetBMC(ctx, bmc.UID)
    require.NoError(t, err)
    assert.True(t, updated.Status.Reconciled)
}
```

## Best Practices

### Reconciler Design

✅ **DO:**
- Keep reconciliation idempotent (safe to run multiple times)
- Update status to reflect observed state
- Return errors for transient failures (will be retried)
- Set conditions for permanent failures
- Use structured logging
- Respect context cancellation
- Emit events for significant state changes

❌ **DON'T:**
- Store state in reconciler struct (it's shared)
- Block indefinitely (respect timeouts)
- Retry internally (let the framework handle it)
- Panic on errors
- Modify Spec fields (only Status)

### Event Design

✅ **DO:**
- Use descriptive event types
- Follow naming conventions
- Include relevant data in event payload
- Document custom event types
- Use wildcards for broad subscriptions

❌ **DON'T:**
- Create events for every field change
- Include sensitive data in events
- Use events for synchronous operations
- Create circular event chains

### Workflow Design

✅ **DO:**
- Use workflows for multi-step operations
- Make workflow steps idempotent
- Handle partial failures gracefully
- Include rollback logic
- Track workflow progress

❌ **DON'T:**
- Use workflows for simple operations
- Block indefinitely in workflow steps
- Ignore errors
- Nest workflows too deeply

## Configuration

### Server Configuration

Enable reconciliation in `cmd/server/main.go`:

```go
var (
    reconcileEnabled  = flag.Bool("reconcile", true, "Enable reconciliation")
    reconcileInterval = flag.Duration("reconcile-interval", 5*time.Minute, 
                                      "Reconciliation interval")
    reconcileWorkers  = flag.Int("reconcile-workers", 10, 
                                 "Number of reconciliation workers")
)

func main() {
    flag.Parse()
    
    if *reconcileEnabled {
        // Create event bus
        eventBus := events.NewInMemoryEventBus(*reconcileWorkers)
        eventBus.Start()
        defer eventBus.Stop()
        
        // Create controller
        controller := reconcile.NewController(eventBus, storage)
        controller.SetReconcileInterval(*reconcileInterval)
        
        // Register reconcilers
        reconcilers.RegisterReconcilers(controller, storage, eventBus)
        
        // Start controller
        go controller.Start(ctx)
        defer controller.Stop()
    }
}
```

### CLI Flags

```bash
# Enable reconciliation (default)
./server --reconcile=true

# Disable reconciliation
./server --reconcile=false

# Set reconciliation interval
./server --reconcile-interval=10m

# Set worker count
./server --reconcile-workers=20
```

## Monitoring

### Metrics

The reconciliation system exposes metrics (when enabled):

- `reconcile_requests_total` - Total reconciliation requests
- `reconcile_errors_total` - Total reconciliation errors
- `reconcile_duration_seconds` - Reconciliation duration histogram
- `event_publish_total` - Total events published
- `event_subscribe_total` - Total event subscriptions

### Logging

Enable debug logging to see reconciliation activity:

```go
logger := reconcile.NewDefaultLogger()
logger.SetLevel(reconcile.LogLevelDebug)
```

Log output:
```
[INFO] Starting reconciliation controller
[INFO] Registered reconciler: BMC
[INFO] Registered reconciler: FRU
[DEBUG] Reconciling BMC: bmc-001
[DEBUG] BMC connection successful: bmc-001
[INFO] Reconciliation complete: BMC/bmc-001
[DEBUG] Emitting event: io.openchami.inventory.bmcs.reconciled
```

## Troubleshooting

### Reconciliation Not Triggering

**Check:**
1. Is reconciliation enabled? (`--reconcile=true`)
2. Are reconcilers registered?
3. Are events being emitted from storage?
4. Check logs for errors

### Reconciliation Failing

**Check:**
1. Review error logs
2. Check resource status conditions
3. Verify external dependencies (BMC reachable?)
4. Check for rate limiting (backoff)

### Events Not Received

**Check:**
1. Is event bus started?
2. Are subscriptions registered correctly?
3. Check event type patterns (wildcards?)
4. Review event bus logs

## See Also

- [Reconciler Development Guide](../RECONCILER-GUIDE.md) - Detailed reconciler customization
- [Phase 2 Completion](../PHASE2-COMPLETION.md) - Code generation implementation
- [Reconciliation Proposal](../RECONCILIATION-PROPOSAL.md) - Original design document
- [Implementation Status](../../IMPLEMENTATION-STATUS.md) - Current status

## References

- [CloudEvents Specification](https://cloudevents.io/)
- [Kubernetes Controller Pattern](https://kubernetes.io/docs/concepts/architecture/controller/)
- [Go Workflows](https://github.com/cschleiden/go-workflows)
- [Temporal Workflows](https://temporal.io/)
