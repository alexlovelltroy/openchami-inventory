# Reconciler Development Guide

## Quick Start

This guide shows how to work with the generated reconciler boilerplate.

## Generating Reconcilers

Generate reconciler boilerplate for all registered resources:

```bash
go run cmd/codegen/main.go -type reconcile -output pkg/reconcilers -package reconcilers
```

This creates:
- `{resource}_reconciler_generated.go` - One per resource
- `registration_generated.go` - Automatic registration
- `event_handlers_generated.go` - Cross-resource event handling

## Customizing a Reconciler

Each generated reconciler has a stub method for custom logic:

### Example: BMC Reconciler

**File:** `pkg/reconcilers/bmc_reconciler_generated.go`

**Find the stub method:**
```go
func (r *BMCReconciler) reconcileBMC(ctx context.Context, bmc *bmc.BMC) error {
    // TODO: Implement BMC-specific reconciliation logic here
    //
    // This method is called by the Reconcile() method after basic setup.
    // Implement your resource-specific logic here.
    //
    // Example:
    //   1. Connect to the BMC using bmc.Spec.Endpoint
    //   2. Query the BMC status
    //   3. Update bmc.Status fields
    //   4. Return any errors
    //
    // The Reconcile() method handles:
    //   - Event emission on success/failure
    //   - Status updates
    //   - Condition management
    //   - Error handling

    return nil
}
```

**Implement your logic:**
```go
func (r *BMCReconciler) reconcileBMC(ctx context.Context, bmc *bmc.BMC) error {
    // Connect to BMC
    client, err := redfish.Connect(bmc.Spec.Endpoint, bmc.Spec.Credentials)
    if err != nil {
        return fmt.Errorf("failed to connect: %w", err)
    }
    defer client.Logout()
    
    // Query BMC status
    system, err := client.GetSystem()
    if err != nil {
        return fmt.Errorf("failed to get system: %w", err)
    }
    
    // Update status
    bmc.Status.Connected = true
    bmc.Status.PowerState = system.PowerState
    bmc.Status.Health = system.Health
    bmc.Status.LastSeen = time.Now()
    
    return nil
}
```

## Adding Custom Fields

You can add custom fields to the reconciler struct:

```go
type BMCReconciler struct {
    reconcile.BaseReconciler
    
    // Add your custom fields
    redfishClient *redfish.Client
    timeout       time.Duration
    retryCount    int
}
```

Update the constructor:
```go
func NewDefaultBMCReconciler(client reconcile.ClientInterface, eventBus events.EventBus) *BMCReconciler {
    return &BMCReconciler{
        BaseReconciler: reconcile.BaseReconciler{
            Client:   client,
            EventBus: eventBus,
            Logger:   reconcile.NewDefaultLogger(),
        },
        timeout:    30 * time.Second,
        retryCount: 3,
    }
}
```

## Implementing Cross-Resource Event Handlers

**File:** `pkg/reconcilers/event_handlers_generated.go`

### 1. Add Handler Method

Uncomment and customize the example handler:

```go
func (r *EventHandlerRegistry) handleBMCConnected(ctx context.Context, event events.Event) error {
    r.logger.Infof("BMC connected: %s", event.ResourceUID())
    
    // Get the BMC
    bmcUID := event.ResourceUID()
    resource, err := r.client.Get(ctx, bmcUID, "BMC")
    if err != nil {
        return fmt.Errorf("failed to get BMC: %w", err)
    }
    
    bmc := resource.(*bmc.BMC)
    
    // Trigger FRU discovery
    return r.discoverFRUs(ctx, bmc)
}

func (r *EventHandlerRegistry) discoverFRUs(ctx context.Context, bmc *bmc.BMC) error {
    // Connect to BMC
    client, err := redfish.Connect(bmc.Spec.Endpoint, bmc.Spec.Credentials)
    if err != nil {
        return err
    }
    defer client.Logout()
    
    // List FRUs
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
            r.logger.Errorf("Failed to create FRU %s: %v", fruData.Id, err)
            continue
        }
        
        r.logger.Infof("Created FRU: %s", fru.UID)
    }
    
    return nil
}
```

### 2. Register Handler

Update the `RegisterEventHandlers()` method:

```go
func (r *EventHandlerRegistry) RegisterEventHandlers(eventBus events.EventBus) error {
    // Register BMC connected handler
    _, err := eventBus.Subscribe("io.openchami.inventory.bmcs.connected", r.handleBMCConnected)
    if err != nil {
        return fmt.Errorf("failed to subscribe to BMC connected events: %w", err)
    }
    
    r.logger.Infof("Event handlers registered successfully")
    return nil
}
```

### 3. Update Registered Handlers List

```go
func (r *EventHandlerRegistry) GetRegisteredEventHandlers() []string {
    return []string{
        "io.openchami.inventory.bmcs.connected",
    }
}
```

## Integration with Server

### 1. Initialize Event Bus and Storage

**File:** `cmd/server/main.go`

```go
func main() {
    // ... existing setup ...
    
    // Create event bus
    eventBus := events.NewInMemoryEventBus(10) // 10 workers
    eventBus.Start()
    defer eventBus.Stop()
    
    // Create storage with event bus
    storage, err := storage.NewStorage(storageDir, eventBus)
    if err != nil {
        log.Fatal(err)
    }
    
    // ... continue setup ...
}
```

### 2. Start Reconciliation Controller

```go
func main() {
    // ... after creating storage and event bus ...
    
    // Create reconciliation controller
    controller := reconcile.NewController(eventBus, storage)
    
    // Register all reconcilers
    if err := reconcilers.RegisterReconcilers(controller, storage, eventBus); err != nil {
        log.Fatalf("Failed to register reconcilers: %v", err)
    }
    
    // Register event handlers
    handlers := reconcilers.NewEventHandlerRegistry(storage, eventBus)
    if err := handlers.RegisterEventHandlers(eventBus); err != nil {
        log.Fatalf("Failed to register event handlers: %v", err)
    }
    
    // Start controller
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    go controller.Start(ctx)
    defer controller.Stop()
    
    log.Println("Reconciliation controller started")
    
    // ... start HTTP server ...
}
```

### 3. Emit Events from Storage

Update storage operations to emit events:

```go
func (s *Storage) SaveBMC(ctx context.Context, bmc *bmc.BMC) error {
    // Determine if this is create or update
    exists := s.Exists(ctx, bmc.UID, "BMC")
    
    // Save the resource
    if err := s.save(ctx, bmc); err != nil {
        return err
    }
    
    // Emit appropriate event
    eventType := "io.openchami.inventory.bmcs.created"
    if exists {
        eventType = "io.openchami.inventory.bmcs.updated"
    }
    
    event := events.NewResourceEvent(eventType, bmc.UID, "BMC", map[string]interface{}{
        "bmc": bmc,
    })
    
    return s.eventBus.Publish(ctx, event)
}

func (s *Storage) DeleteBMC(ctx context.Context, uid string) error {
    // Delete the resource
    if err := s.delete(ctx, uid, "BMC"); err != nil {
        return err
    }
    
    // Emit delete event
    event := events.NewResourceEvent(
        "io.openchami.inventory.bmcs.deleted",
        uid,
        "BMC",
        nil,
    )
    
    return s.eventBus.Publish(ctx, event)
}
```

## Testing Reconcilers

### Unit Testing

Create a test file for your reconciler:

**File:** `pkg/reconcilers/bmc_reconciler_test.go`

```go
package reconcilers

import (
    "context"
    "testing"
    
    "github.com/openchami/inventory/pkg/events"
    "github.com/openchami/inventory/pkg/reconcile"
    "github.com/openchami/inventory/pkg/resources/bmc"
)

type mockClient struct {
    reconcile.ClientInterface
    updated bool
}

func (m *mockClient) Update(ctx context.Context, resource interface{}) error {
    m.updated = true
    return nil
}

func TestBMCReconciler_Reconcile(t *testing.T) {
    // Setup
    eventBus := events.NewInMemoryEventBus(1)
    eventBus.Start()
    defer eventBus.Stop()
    
    client := &mockClient{}
    reconciler := NewDefaultBMCReconciler(client, eventBus)
    
    // Create test BMC
    testBMC := &bmc.BMC{
        Resource: resources.Resource{
            Kind: "BMC",
            UID:  "test-bmc-1",
        },
        Spec: bmc.BMCSpec{
            Endpoint: "https://bmc.example.com",
        },
    }
    
    // Test reconciliation
    err := reconciler.Reconcile(context.Background(), &testBMC.Resource)
    if err != nil {
        t.Errorf("Reconcile failed: %v", err)
    }
}
```

### Integration Testing

Test the full reconciliation flow:

```go
func TestReconciliationFlow(t *testing.T) {
    // Create components
    eventBus := events.NewInMemoryEventBus(10)
    eventBus.Start()
    defer eventBus.Stop()
    
    storage := // ... create test storage
    controller := reconcile.NewController(eventBus, storage)
    
    // Register reconcilers
    if err := RegisterReconcilers(controller, storage, eventBus); err != nil {
        t.Fatal(err)
    }
    
    // Start controller
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    go controller.Start(ctx)
    defer controller.Stop()
    
    // Create a BMC
    bmc := &bmc.BMC{
        Resource: resources.Resource{
            Kind: "BMC",
            UID:  "test-bmc-1",
        },
        Spec: bmc.BMCSpec{
            Endpoint: "https://bmc.example.com",
        },
    }
    
    // Save (triggers event)
    if err := storage.SaveBMC(ctx, bmc); err != nil {
        t.Fatal(err)
    }
    
    // Wait for reconciliation
    time.Sleep(100 * time.Millisecond)
    
    // Verify reconciliation occurred
    // ... assertions ...
}
```

## Debugging

### Enable Debug Logging

```go
func NewDefaultBMCReconciler(client reconcile.ClientInterface, eventBus events.EventBus) *BMCReconciler {
    logger := reconcile.NewDefaultLogger()
    logger.SetLevel(reconcile.LogLevelDebug) // Enable debug logs
    
    return &BMCReconciler{
        BaseReconciler: reconcile.BaseReconciler{
            Client:   client,
            EventBus: eventBus,
            Logger:   logger,
        },
    }
}
```

### Add Custom Logging

```go
func (r *BMCReconciler) reconcileBMC(ctx context.Context, bmc *bmc.BMC) error {
    r.Logger.Infof("Starting BMC reconciliation: %s", bmc.UID)
    r.Logger.Debugf("BMC endpoint: %s", bmc.Spec.Endpoint)
    
    // ... reconciliation logic ...
    
    r.Logger.Infof("BMC reconciliation complete: %s", bmc.UID)
    return nil
}
```

## Common Patterns

### Periodic Reconciliation

Reconcilers run periodically by default. Customize the interval:

```go
controller := reconcile.NewController(eventBus, storage)
controller.SetReconcileInterval(10 * time.Minute)
```

### Conditional Reconciliation

Skip reconciliation based on conditions:

```go
func (r *BMCReconciler) reconcileBMC(ctx context.Context, bmc *bmc.BMC) error {
    // Skip if already processing
    if bmc.Status.Processing {
        r.Logger.Debugf("BMC %s already processing, skipping", bmc.UID)
        return nil
    }
    
    // Mark as processing
    bmc.Status.Processing = true
    defer func() { bmc.Status.Processing = false }()
    
    // ... reconciliation logic ...
}
```

### Error Handling with Retry

```go
func (r *BMCReconciler) reconcileBMC(ctx context.Context, bmc *bmc.BMC) error {
    var lastErr error
    
    for i := 0; i < r.retryCount; i++ {
        if err := r.tryReconcile(ctx, bmc); err == nil {
            return nil
        } else {
            lastErr = err
            r.Logger.Warnf("Reconciliation attempt %d failed: %v", i+1, err)
            time.Sleep(time.Duration(i+1) * time.Second)
        }
    }
    
    return fmt.Errorf("reconciliation failed after %d attempts: %w", r.retryCount, lastErr)
}
```

## Best Practices

1. **Keep reconciliation idempotent** - Running multiple times should be safe
2. **Handle missing resources gracefully** - Resources may be deleted
3. **Use contexts for cancellation** - Respect context timeouts
4. **Log important events** - Use structured logging
5. **Emit events for state changes** - Enable reactive behavior
6. **Update status fields** - Reflect observed state
7. **Set conditions** - Track reconciliation health
8. **Return errors** - Let framework handle retry logic
9. **Test thoroughly** - Both unit and integration tests
10. **Document custom logic** - Help future maintainers

## Event Types Reference

Standard event types:
```
io.openchami.inventory.{resource}.created
io.openchami.inventory.{resource}.updated
io.openchami.inventory.{resource}.deleted
io.openchami.inventory.{resource}.reconciled
io.openchami.inventory.{resource}.error
```

Custom event types (add as needed):
```
io.openchami.inventory.bmcs.connected
io.openchami.inventory.bmcs.disconnected
io.openchami.inventory.nodes.provisioned
io.openchami.inventory.nodes.deprovisioned
io.openchami.inventory.frus.discovered
```

## Next Steps

1. Review the [RECONCILIATION-PROPOSAL.md](../RECONCILIATION-PROPOSAL.md) for architecture details
2. Check [IMPLEMENTATION-STATUS.md](../IMPLEMENTATION-STATUS.md) for current status
3. See [PHASE2-COMPLETION.md](PHASE2-COMPLETION.md) for Phase 2 summary
4. Implement your reconciler logic in the stub methods
5. Add tests for your reconciliation logic
6. Wire up the controller in the server
7. Test the full integration

## Getting Help

- Review generated code comments - they contain usage examples
- Check existing reconcilers for patterns
- Look at the base reconciler utilities in `pkg/reconcile/reconciler.go`
- Examine event bus patterns in `pkg/events/`
