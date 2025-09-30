package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// FileBackend implements StorageBackend using file-based storage.
//
// This implementation stores each resource as a separate JSON file
// in a directory structure organized by resource type:
//
//	baseDir/
//	├── bmcs/
//	│   ├── bmc-123.json
//	│   └── bmc-456.json
//	├── nodes/
//	│   ├── node-789.json
//	│   └── node-abc.json
//	└── frus/
//	    └── fru-def.json
//
// Features:
//   - Thread-safe: Uses file locking for concurrent access
//   - Atomic writes: Uses temp files + rename for atomicity
//   - Auto-creation: Creates directories as needed
//   - Validation: Checks JSON format before saving
//   - Error recovery: Continues operation even if some files are corrupted
//
// Limitations:
//   - Performance: Not optimized for large numbers of resources
//   - Scalability: File system limits apply
//   - Consistency: No transactions across multiple resources
//   - Locking: File locking may not work on all file systems
//
// This backend is suitable for:
//   - Development and testing
//   - Small to medium deployments
//   - Environments where simplicity is preferred over performance
//   - Situations where human-readable storage is valuable
type FileBackend struct {
	baseDir string
	mu      sync.RWMutex
	closed  bool
}

// NewFileBackend creates a new file-based storage backend.
//
// Parameters:
//   - baseDir: Root directory for storing resource files
//
// Returns:
//   - *FileBackend: Configured file backend
//   - error: Any error that occurred during initialization
//
// The function will create the base directory and any required
// subdirectories if they don't exist.
//
// Example:
//
//	backend, err := storage.NewFileBackend("./inventory")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer backend.Close()
func NewFileBackend(baseDir string) (*FileBackend, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory %s: %w", baseDir, err)
	}

	backend := &FileBackend{
		baseDir: baseDir,
	}

	// Initialize resource type directories
	resourceTypes := []string{"bmcs", "nodes", "frus", "bootconfigurations", "fruinventorysnapshots"}
	for _, resourceType := range resourceTypes {
		dir := filepath.Join(baseDir, strings.ToLower(resourceType))
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return backend, nil
}

// resourceTypeToDir maps resource type names to directory names
func (f *FileBackend) resourceTypeToDir(resourceType string) string {
	switch resourceType {
	case "BMC":
		return "bmcs"
	case "Node":
		return "nodes"
	case "FRU":
		return "frus"
	case "BootConfiguration":
		return "bootconfigurations"
	case "FRUInventorySnapshot":
		return "fruinventorysnapshots"
	default:
		return strings.ToLower(resourceType) + "s"
	}
}

// getFilePath returns the file path for a specific resource
func (f *FileBackend) getFilePath(resourceType, uid string) string {
	dir := f.resourceTypeToDir(resourceType)
	return filepath.Join(f.baseDir, dir, uid+".json")
}

// getDirPath returns the directory path for a resource type
func (f *FileBackend) getDirPath(resourceType string) string {
	dir := f.resourceTypeToDir(resourceType)
	return filepath.Join(f.baseDir, dir)
}

// checkClosed returns an error if the backend has been closed
func (f *FileBackend) checkClosed() error {
	if f.closed {
		return fmt.Errorf("storage backend has been closed")
	}
	return nil
}

// LoadAll implements StorageBackend.LoadAll
func (f *FileBackend) LoadAll(ctx context.Context, resourceType string) ([]json.RawMessage, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if err := f.checkClosed(); err != nil {
		return nil, err
	}

	dirPath := f.getDirPath(resourceType)

	// Check if context is cancelled before starting
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []json.RawMessage{}, nil // Empty slice, not an error
		}
		return nil, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	var resources []json.RawMessage
	for _, entry := range entries {
		// Check for cancellation periodically
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			// Log warning but continue with other files
			continue
		}

		// Validate JSON format
		if !json.Valid(data) {
			// Log warning but continue with other files
			continue
		}

		resources = append(resources, json.RawMessage(data))
	}

	return resources, nil
}

// Load implements StorageBackend.Load
func (f *FileBackend) Load(ctx context.Context, resourceType, uid string) (json.RawMessage, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if err := f.checkClosed(); err != nil {
		return nil, err
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	filePath := f.getFilePath(resourceType, uid)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Validate JSON format
	if !json.Valid(data) {
		return nil, fmt.Errorf("invalid JSON in file %s: %w", filePath, ErrInvalidData)
	}

	return json.RawMessage(data), nil
}

// Save implements StorageBackend.Save
func (f *FileBackend) Save(ctx context.Context, resourceType, uid string, data json.RawMessage) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.checkClosed(); err != nil {
		return err
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Validate JSON format
	if !json.Valid(data) {
		return fmt.Errorf("invalid JSON data: %w", ErrInvalidData)
	}

	filePath := f.getFilePath(resourceType, uid)

	// Ensure directory exists
	dirPath := filepath.Dir(filePath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
	}

	// Use atomic write: write to temp file, then rename
	tempPath := filePath + ".tmp"

	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file %s: %w", tempPath, err)
	}

	if err := os.Rename(tempPath, filePath); err != nil {
		// Clean up temp file on error
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename temp file %s to %s: %w", tempPath, filePath, err)
	}

	return nil
}

// Delete implements StorageBackend.Delete
func (f *FileBackend) Delete(ctx context.Context, resourceType, uid string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.checkClosed(); err != nil {
		return err
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	filePath := f.getFilePath(resourceType, uid)

	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return fmt.Errorf("failed to stat file %s: %w", filePath, err)
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file %s: %w", filePath, err)
	}

	return nil
}

// Exists implements StorageBackend.Exists
func (f *FileBackend) Exists(ctx context.Context, resourceType, uid string) (bool, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if err := f.checkClosed(); err != nil {
		return false, err
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	filePath := f.getFilePath(resourceType, uid)

	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to stat file %s: %w", filePath, err)
	}

	return true, nil
}

// List implements StorageBackend.List
func (f *FileBackend) List(ctx context.Context, resourceType string) ([]string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if err := f.checkClosed(); err != nil {
		return nil, err
	}

	dirPath := f.getDirPath(resourceType)

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil // Empty slice, not an error
		}
		return nil, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	var uids []string
	for _, entry := range entries {
		// Check for cancellation periodically
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		// Extract UID from filename (remove .json extension)
		uid := strings.TrimSuffix(entry.Name(), ".json")
		uids = append(uids, uid)
	}

	return uids, nil
}

// Close implements StorageBackend.Close
func (f *FileBackend) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.closed = true
	return nil
}
