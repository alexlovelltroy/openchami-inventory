# Storage System

> 📖 **For detailed information**, this README contains the complete storage system documentation. For general development, see [Development Guide](../../docs/developer/DEVELOPMENT.md).

The OpenCHAMI Inventory storage system is flexible and pluggable, allowing different storage backends without changing application code.

## Quick Overview

**Architecture:**
- `StorageBackend` interface - Low-level backend operations
- `ResourceStorage[T]` - Type-safe resource operations
- Generated code - Automatic storage operations per resource type

**Current Backends:**
- **File Backend** (default) - JSON files in directory structure
- **Future**: Database backends (PostgreSQL, SQLite, etc.)

**Directory Structure:**
```
inventory/
├── bmcs/
│   ├── bmc-123.json
│   └── bmc-456.json
├── nodes/
└── frus/
```

## Storage Interface

```go
type StorageBackend interface {
    LoadAll(ctx context.Context, resourceType string) ([]json.RawMessage, error)
    Load(ctx context.Context, resourceType, uid string) (json.RawMessage, error)
    Save(ctx context.Context, resourceType, uid string, data json.RawMessage) error
    Delete(ctx context.Context, resourceType, uid string) error
    Exists(ctx context.Context, resourceType, uid string) (bool, error)
    List(ctx context.Context, resourceType string) ([]string, error)
    Close() error
}
```

## Type-Safe Resource Storage

```go
type ResourceStorage[T any] interface {
    LoadAll(ctx context.Context) ([]T, error)
    Load(ctx context.Context, uid string) (T, error)
    Save(ctx context.Context, resource T) error
    Delete(ctx context.Context, uid string) error
    Exists(ctx context.Context, uid string) (bool, error)
    List(ctx context.Context) ([]string, error)
}
```

## File Backend (Default)

Stores resources as JSON files:

**Features:**
- Thread-safe with file locking
- Atomic writes using temp files + rename
- Auto-creation of directories
- JSON validation
- Human-readable storage format

**Suitable for:**
- Development and testing
- Small to medium deployments
- Environments where simplicity is preferred
- Situations where human-readable storage is valuable

### Future Backends
The interface design supports additional backends:
- **DatabaseBackend**: PostgreSQL, MySQL, SQLite
- **CloudBackend**: AWS S3, Google Cloud Storage, Azure Blob
- **MemoryBackend**: In-memory storage for testing
- **CacheBackend**: Redis, Memcached with fallback

## Usage Examples

### Basic Usage (Package-Level Functions)
The simplest way to use storage is through package-level convenience functions:

```go
import "github.com/openchami/inventory/internal/storage"

// Load all BMCs
bmcs, err := storage.LoadAllBMCs()

// Load specific BMC
bmc, err := storage.LoadBMC("bmc-123")

// Save a BMC
err = storage.SaveBMC(bmc)

// Delete a BMC
err = storage.DeleteBMC("bmc-123")
```

### Advanced Usage (Direct Backend Access)
For more control, use the backend interfaces directly:

```go
// Initialize file backend
backend, err := storage.NewFileBackend("./data")
if err != nil {
    log.Fatal(err)
}
defer backend.Close()

// Get type-safe storage for BMCs
bmcStorage := storage.GetBMCStorage(backend)

// Use with context
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

bmcs, err := bmcStorage.LoadAll(ctx)
bmc, err := bmcStorage.Load(ctx, "bmc-123")
err = bmcStorage.Save(ctx, bmc)
```

### Custom Backend Configuration
Set a custom backend for package-level functions:

```go
// Use custom file location
backend, err := storage.NewFileBackend("/var/lib/openchami")
if err != nil {
    log.Fatal(err)
}
storage.SetDefaultBackend(backend)

// Now package-level functions use the custom backend
bmcs, err := storage.LoadAllBMCs()
```

## Resource Types

The storage system provides type-safe access for all resource types:

| Resource Type | Storage Function | Type |
|---------------|------------------|------|
| BMC | `GetBMCStorage()` | `*bmc.BMC` |
| Node | `GetNodeStorage()` | `*node.Node` |
| FRU | `GetFRUStorage()` | `*fru.FRU` |
| Boot Configuration | `GetBootConfigurationStorage()` | `*boot.BootConfiguration` |
| FRU Inventory Snapshot | `GetFRUInventorySnapshotStorage()` | `*fru.FRUInventorySnapshot` |

## Error Handling

The storage system provides structured error handling:

```go
import "errors"

bmc, err := storage.LoadBMC("bmc-123")
if err != nil {
    if errors.Is(err, storage.ErrNotFound) {
        // Handle missing resource
        fmt.Println("BMC not found")
    } else if errors.Is(err, storage.ErrInvalidData) {
        // Handle corrupted data
        fmt.Println("BMC data is corrupted")
    } else {
        // Handle other errors (I/O, permissions, etc.)
        fmt.Printf("Storage error: %v\n", err)
    }
}
```

## Thread Safety

All storage backends are designed to be thread-safe:

```go
// Multiple goroutines can safely access the same backend
backend := storage.GetDefaultBackend()

go func() {
    bmcStorage := storage.GetBMCStorage(backend)
    bmcs, _ := bmcStorage.LoadAll(ctx)
    // ... process bmcs
}()

go func() {
    nodeStorage := storage.GetNodeStorage(backend)
    nodes, _ := nodeStorage.LoadAll(ctx)
    // ... process nodes
}()
```

## Performance Considerations

### File Backend
- **Small datasets**: Excellent performance
- **Large datasets**: May become slow (consider database backend)
- **Concurrent access**: Uses file locking (may limit scalability)
- **Memory usage**: Loads entire resources into memory

### Optimization Tips
1. **Use context with timeouts** for long-running operations
2. **Batch operations** when loading many resources
3. **Consider List() + Load()** instead of LoadAll() for large datasets
4. **Implement custom backends** for specific performance requirements

## Adding New Storage Backends

To implement a new storage backend:

1. **Implement StorageBackend interface**:
```go
type MyBackend struct {
    // your fields
}

func (b *MyBackend) LoadAll(ctx context.Context, resourceType string) ([]json.RawMessage, error) {
    // your implementation
}
// ... implement other methods
```

2. **Add constructor function**:
```go
func NewMyBackend(config MyConfig) (*MyBackend, error) {
    // initialize and return backend
}
```

3. **Use in application**:
```go
backend, err := storage.NewMyBackend(config)
if err != nil {
    log.Fatal(err)
}
storage.SetDefaultBackend(backend)
```

## Configuration

### File Backend Configuration
```go
// Default location (./inventory)
backend, _ := storage.NewFileBackend("inventory")

// Custom location
backend, _ := storage.NewFileBackend("/var/lib/openchami/data")

// With error handling
backend, err := storage.NewFileBackend(dataDir)
if err != nil {
    log.Fatalf("Failed to initialize storage: %v", err)
}
defer backend.Close()
```

### Environment-Based Configuration
```go
dataDir := os.Getenv("INVENTORY_DATA_DIR")
if dataDir == "" {
    dataDir = "inventory" // default
}

backend, err := storage.NewFileBackend(dataDir)
if err != nil {
    log.Fatal(err)
}
storage.SetDefaultBackend(backend)
```

## Testing

The storage system is designed to be easily testable:

```go
func TestMyFunction(t *testing.T) {
    // Use temporary directory for tests
    tempDir, err := os.MkdirTemp("", "test-storage")
    require.NoError(t, err)
    defer os.RemoveAll(tempDir)
    
    // Create test backend
    backend, err := storage.NewFileBackend(tempDir)
    require.NoError(t, err)
    defer backend.Close()
    
    // Set for package-level functions
    storage.SetDefaultBackend(backend)
    
    // Run your tests
    // ...
}
```

## Migration Between Backends

When changing storage backends, you can migrate data:

```go
// Load from old backend
oldBackend, _ := storage.NewFileBackend("old-data")
bmcStorage := storage.GetBMCStorage(oldBackend)
bmcs, err := bmcStorage.LoadAll(context.Background())

// Save to new backend
newBackend, _ := storage.NewDatabaseBackend(dbConfig)
newBMCStorage := storage.GetBMCStorage(newBackend)
for _, bmc := range bmcs {
    err := newBMCStorage.Save(context.Background(), bmc)
    if err != nil {
        log.Printf("Failed to migrate BMC %s: %v", bmc.GetUID(), err)
    }
}
```

## Documentation

For more information:

- **[Development Guide](../../docs/developer/DEVELOPMENT.md)** - General architecture and development
- **[Code Generation](../../docs/developer/CODE-GENERATION.md)** - How storage code is generated
- **[User Guide](../../docs/user/USER-GUIDE.md)** - Using the system
- **[Troubleshooting](../../docs/user/TROUBLESHOOTING.md)** - Storage-related issues
## Documentation

For more information:

- **[Development Guide](../../docs/developer/DEVELOPMENT.md)** - General architecture and development
- **[Code Generation](../../docs/developer/CODE-GENERATION.md)** - How storage code is generated
- **[User Guide](../../docs/user/USER-GUIDE.md)** - Using the system
- **[Troubleshooting](../../docs/user/TROUBLESHOOTING.md)** - Storage-related issues
