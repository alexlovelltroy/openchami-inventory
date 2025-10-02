# OpenCHAMI Inventory Development Guide

## Architecture Overview

This project uses **code generation** and **event-driven reconciliation** to maintain consistency across resource types while enabling declarative infrastructure management. Understanding this architecture is key to effective development.

### Core Concepts

#### 1. Resource Package Structure
```
pkg/resources/
├── bmc/           # BMC-specific types and policies
├── node/          # Node-specific types and policies  
├── fru/           # FRU-specific types and policies
├── boot/          # Boot configuration types and policies
└── main.go        # Common resource interfaces and utilities
```

#### 2. Event-Driven System
```
pkg/events/        # CloudEvents-compliant event system
├── events.go      # Event interface and types
└── memory_bus.go  # In-memory event bus with pub/sub

pkg/reconcile/     # Reconciliation framework
├── reconciler.go  # Reconciler interface and base implementation
├── controller.go  # Reconciliation controller
└── workqueue.go   # Work queue with rate limiting

pkg/workflows/     # Workflow engine for complex operations
├── workflows.go   # Workflow interface
├── goworkflows.go # Embedded workflow engine
└── temporal.go    # Temporal integration (optional)

pkg/reconcilers/   # Generated reconcilers (via code generation)
├── *_reconciler_generated.go  # Per-resource reconcilers
├── registration_generated.go  # Automatic registration
└── event_handlers_generated.go # Cross-resource event handlers
```

#### 3. Generated vs. Manual Code

**Generated Code** (⚠️ Do NOT edit directly):
- `cmd/server/*_handlers_generated.go` - REST API handlers
- `cmd/server/models_generated.go` - Request/response types
- `cmd/server/routes_generated.go` - Route registration
- `cmd/server/policies_generated.go` - Auth integration
- `internal/storage/storage_generated.go` - Storage operations
- `pkg/client/` - HTTP client library
- `pkg/reconcilers/*_reconciler_generated.go` - Reconciler boilerplate
- `pkg/reconcilers/registration_generated.go` - Reconciler registration
- `pkg/reconcilers/event_handlers_generated.go` - Event handler registry

**Manual Code** (✅ Safe to edit):
- `pkg/resources/*/` - Resource type definitions
- `pkg/policies/` - Authentication/authorization logic
- `cmd/server/main.go` - Server setup and configuration
- `cmd/crawler/` - Hardware discovery logic
- `pkg/codegen/templates/` - Code generation template files
- `pkg/events/` - Event system implementation
- `pkg/reconcile/` - Reconciliation framework
- `pkg/workflows/` - Workflow engine

**Customizable Generated Code** (✅ Edit stub methods only):
- `pkg/reconcilers/*_reconciler_generated.go` - Edit `reconcile{Name}()` method
- `pkg/reconcilers/event_handlers_generated.go` - Add event handler methods

## How to Add a New Resource Type

### Step 1: Define the Resource
Create `pkg/resources/newtype/newtype.go`:
```go
package newtype

import "github.com/openchami/inventory/pkg/resources"

type NewType struct {
    resources.Resource
    Spec   NewTypeSpec   `json:"spec"`
    Status NewTypeStatus `json:"status,omitempty"`
}

type NewTypeSpec struct {
    // Your fields here
}

type NewTypeStatus struct {
    Conditions []resources.Condition `json:"conditions,omitempty"`
}
```

### Step 2: Register in Code Generator
Add to `cmd/codegen/main.go`:
```go
if err := generator.RegisterResource(&newtype.NewType{}); err != nil {
    log.Fatalf("Failed to register NewType resource: %v", err)
}
```

### Step 3: Regenerate Code
```bash
make dev
```

This automatically generates:
- REST API endpoints (`/newtypes`)
- CRUD handlers
- Client library methods
- Storage operations
- Request/response types

### Step 4: Generate Reconciler (Optional)

If your resource needs reconciliation logic:

```bash
# Generate reconciler boilerplate
go run cmd/codegen/main.go -type reconcile -output pkg/reconcilers -package reconcilers
```

This generates:
- `newtype_reconciler_generated.go` - Reconciler with stub method
- Updates `registration_generated.go` - Automatic registration
- Updates `event_handlers_generated.go` - Event handler registry

### Step 5: Implement Reconciliation Logic (Optional)

Edit the stub method in `pkg/reconcilers/newtype_reconciler_generated.go`:

```go
func (r *NewTypeReconciler) reconcileNewType(ctx context.Context, nt *newtype.NewType) error {
    // TODO: Implement reconciliation logic
    // 1. Observe actual state
    // 2. Compare with desired state (nt.Spec)
    // 3. Take actions to align actual with desired
    // 4. Update status fields (nt.Status)
    
    return nil
}
```

See [Reconciliation Guide](RECONCILIATION.md) for details.

## Event-Driven Reconciliation

The system uses an event-driven architecture where resources are automatically reconciled when they change:

### How It Works

1. **User creates/updates resource** via REST API
2. **Storage emits event** (e.g., `bmcs.created`)
3. **Event bus notifies subscribers**
4. **Reconciliation controller** receives event
5. **Reconciler processes resource** to align actual state with desired state
6. **Status updated** and reconciliation event emitted

### Event Flow Example

```
POST /bmcs (Create BMC)
    ↓
Storage.SaveBMC()
    ↓
Emit: io.openchami.inventory.bmcs.created
    ↓
Controller receives event
    ↓
BMCReconciler.Reconcile()
    ↓
Connect to BMC, update status
    ↓
Emit: io.openchami.inventory.bmcs.reconciled
```

### Reconciler Pattern

Reconcilers implement the pattern:

```go
func (r *ResourceReconciler) Reconcile(ctx context.Context, resource *Resource) error {
    // 1. Get desired state from Spec
    desired := resource.Spec
    
    // 2. Observe actual state
    actual := r.observeActualState(resource)
    
    // 3. Compare and take actions
    if !reflect.DeepEqual(actual, desired) {
        if err := r.alignState(actual, desired); err != nil {
            return err // Will be retried
        }
    }
    
    // 4. Update status
    resource.Status = actual
    
    return nil
}
```

### Cross-Resource Reactions

Event handlers enable resources to react to other resources:

```go
// When BMC connects, discover FRUs
eventBus.Subscribe("io.openchami.inventory.bmcs.connected", func(event Event) {
    bmc := event.Data()
    frus := discoverFRUs(bmc)
    for _, fru := range frus {
        storage.SaveFRU(fru) // Triggers FRU reconciliation
    }
})
```

See [Reconciliation Guide](RECONCILIATION.md) for complete details.

## Understanding Generated Code

### API Handlers
Each resource gets 5 standard endpoints:
- `GET /resources` - List all
- `GET /resources/{uid}` - Get by ID  
- `POST /resources` - Create new
- `PUT /resources/{uid}` - Update existing
- `DELETE /resources/{uid}` - Delete

### Storage Layer
Generated storage provides:
- `LoadAllResourceTypes()` - Load all from disk
- `LoadResourceType(uid)` - Load specific resource
- `SaveResourceType(resource)` - Save to disk
- `DeleteResourceType(uid)` - Remove from disk

### Client Library
Generated client provides type-safe methods:
```go
client := inventory.NewClient("http://localhost:9999")
bmcs, err := client.GetBMCs(ctx)
bmc, err := client.GetBMC(ctx, "uid-123")
```

## Customization Points

### 1. Resource Validation
Add validation to your resource spec:
```go
type MyResourceSpec struct {
    Name string `json:"name" validate:"required"`
    // Add other validation tags
}
```

### 2. Authorization Policies
Create resource-specific policies in `pkg/resources/myresource/policy.go`:
```go
func NewDefaultMyResourcePolicy() policies.ResourcePolicy {
    return &myResourcePolicy{}
}

type myResourcePolicy struct{}

func (p *myResourcePolicy) CanList(ctx context.Context, auth *policies.AuthContext, req *http.Request) policies.Decision {
    // Custom authorization logic
}
```

### 3. Custom Storage Logic
Override generated storage by implementing the interface in `internal/storage/`:
```go
func (s *Storage) CustomLoadLogic() error {
    // Your custom storage logic
}
```

## Debugging Generated Code

### 1. Inspect Templates
Look at `pkg/codegen/templates/` files:
- `handlersTemplate` - API handler logic
- `storageTemplate` - Storage operations  
- `clientTemplate` - Client library
- `modelsTemplate` - Request/response types

### 2. Generate Specific Components
```bash
# Generate only storage
go run cmd/codegen/main.go -type=storage

# Generate only client
go run cmd/codegen/main.go -type=client

# Generate only server components
go run cmd/codegen/main.go -type=server
```

### 3. View Generated Output
Generated files have clear headers indicating they're auto-generated:
```go
// Code generated by codegen. DO NOT EDIT.
```

## Common Workflows

### Adding a Field to Existing Resource
1. Edit `pkg/resources/resourcetype/resourcetype.go`
2. Run `make dev`
3. Generated handlers/storage automatically include new field

### Customizing API Behavior
1. Modify templates in `pkg/codegen/templates/`
2. Run `make dev`
3. All resources get the updated behavior

### Adding Authentication
1. Call `generator.EnableAuthForResource("ResourceName")` in `cmd/codegen/main.go`
2. Implement policy in `pkg/resources/resourcetype/policy.go`
3. Register policy in `cmd/server/main.go`

## Best Practices

### ✅ DO
- Edit resource definitions in `pkg/resources/*/`
- Modify templates for cross-cutting changes
- Use the type system to catch errors at compile time
- Add comprehensive tests for your resource types

### ❌ DON'T  
- Edit generated files directly (changes will be lost)
- Bypass the resource interfaces
- Add business logic to generated handlers
- Forget to run `make dev` after changes

## Troubleshooting

### Build Failures
```bash
# Clean and rebuild everything
make clean
make dev
```

### Missing Endpoints
Check that your resource is registered in `cmd/codegen/main.go`

### Import Errors
Ensure your resource package follows the expected structure and naming conventions

## IDE Integration

### VS Code
The generated files are marked as auto-generated, so VS Code will:
- Show them as read-only
- Exclude them from search by default
- Warn when attempting to edit

### GoLand/IntelliJ
Generated files appear with a special icon indicating they're auto-generated.

## Contributing

When contributing:
1. Focus changes on `pkg/resources/` and templates
2. Include tests for new resource types
3. Update this documentation for new patterns
4. Run `make dev` before committing to ensure everything generates correctly

---

**Key Takeaway**: The generated code provides consistency and reduces boilerplate, but the real development happens in the resource definitions and templates. Think of generation as "smart copying" - you define the pattern once, and it gets applied everywhere.