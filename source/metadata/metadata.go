package metadata

import (
	"context"
	"file-manager/storage"
	"fmt"
	"sync"
	"time"
)

// FileMetadata represents the logical metadata for a file in our system.
// This is distinct from cloud-specific FileInfo, though it contains it.
type FileMetadata struct {
	ID          string                       `json:"id"`
	LogicalPath string                       `json:"logical_path"` // e.g., /users/john/documents/report.pdf
	FileName    string                       `json:"file_name"`
	Size        int64                        `json:"size"`
	ContentType string                       `json:"content_type"`
	UploadedAt  time.Time                    `json:"uploaded_at"`
	UploadedBy  string                       `json:"uploaded_by"`  // Server ID
	CloudCopies map[string]*storage.FileInfo `json:"cloud_copies"` // Map of cloud_provider -> FileInfo
	CustomTags  map[string]string            `json:"custom_tags,omitempty"`
}

// MetadataStore defines the interface for metadata operations.
type MetadataStore interface {
	CreateFileMetadata(ctx context.Context, meta *FileMetadata) error
	GetFileMetadata(ctx context.Context, id string) (*FileMetadata, error)
	GetFileMetadataByPath(ctx context.Context, logicalPath string) (*FileMetadata, error)
	ListFileMetadata(ctx context.Context, prefix string) ([]*FileMetadata, error)
	UpdateFileMetadata(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteFileMetadata(ctx context.Context, id string) error
}

// InMemoryMetadataStore is a simple in-memory implementation of MetadataStore.
// NOT FOR PRODUCTION USE.
type InMemoryMetadataStore struct {
	mu        sync.RWMutex
	store     map[string]*FileMetadata // map[id]*FileMetadata
	pathIndex map[string]string        // map[logicalPath]id
}

// NewInMemoryMetadataStore creates a new InMemoryMetadataStore.
func NewInMemoryMetadataStore() *InMemoryMetadataStore {
	return &InMemoryMetadataStore{
		store:     make(map[string]*FileMetadata),
		pathIndex: make(map[string]string),
	}
}

// CreateFileMetadata adds new file metadata.
func (m *InMemoryMetadataStore) CreateFileMetadata(ctx context.Context, meta *FileMetadata) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.store[meta.ID]; exists {
		return fmt.Errorf("file metadata with ID %s already exists", meta.ID)
	}
	if _, exists := m.pathIndex[meta.LogicalPath]; exists {
		return fmt.Errorf("file metadata with path %s already exists", meta.LogicalPath)
	}

	m.store[meta.ID] = meta
	m.pathIndex[meta.LogicalPath] = meta.ID
	return nil
}

// GetFileMetadata retrieves file metadata by ID.
func (m *InMemoryMetadataStore) GetFileMetadata(ctx context.Context, id string) (*FileMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	meta, ok := m.store[id]
	if !ok {
		return nil, fmt.Errorf("file metadata with ID %s not found", id)
	}
	return meta, nil
}

// GetFileMetadataByPath retrieves file metadata by logical path.
func (m *InMemoryMetadataStore) GetFileMetadataByPath(ctx context.Context, logicalPath string) (*FileMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	id, ok := m.pathIndex[logicalPath]
	if !ok {
		return nil, fmt.Errorf("file metadata with path %s not found", logicalPath)
	}
	return m.store[id], nil
}

// ListFileMetadata lists file metadata based on prefix and server ID.
func (m *InMemoryMetadataStore) ListFileMetadata(ctx context.Context, prefix string) ([]*FileMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*FileMetadata
	for _, meta := range m.store {
		if prefix == "" || (len(meta.LogicalPath) >= len(prefix) && meta.LogicalPath[:len(prefix)] == prefix) {
			results = append(results, meta)
		}
	}
	return results, nil
}

// UpdateFileMetadata updates existing file metadata.
func (m *InMemoryMetadataStore) UpdateFileMetadata(ctx context.Context, id string, updates map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	meta, ok := m.store[id]
	if !ok {
		return fmt.Errorf("file metadata with ID %s not found", id)
	}

	// Apply updates (simplified, needs robust type checking in production)
	for k, v := range updates {
		switch k {
		case "logical_path":
			if newPath, ok := v.(string); ok {
				if _, exists := m.pathIndex[newPath]; exists && m.pathIndex[newPath] != id {
					return fmt.Errorf("new logical path %s already exists for another file", newPath)
				}
				delete(m.pathIndex, meta.LogicalPath)
				meta.LogicalPath = newPath
				m.pathIndex[newPath] = id
			}
		case "file_name":
			if newName, ok := v.(string); ok {
				meta.FileName = newName
			}
		case "content_type":
			if newType, ok := v.(string); ok {
				meta.ContentType = newType
			}
		case "custom_tags":
			if tags, ok := v.(map[string]string); ok {
				meta.CustomTags = tags
			}
		case "cloud_copies":
			if copies, ok := v.(map[string]*storage.FileInfo); ok {
				meta.CloudCopies = copies
			}
			// Add more fields as needed
		}
	}
	return nil
}

// DeleteFileMetadata deletes file metadata by ID.
func (m *InMemoryMetadataStore) DeleteFileMetadata(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	meta, ok := m.store[id]
	if !ok {
		return fmt.Errorf("file metadata with ID %s not found", id)
	}

	delete(m.store, id)
	delete(m.pathIndex, meta.LogicalPath)
	return nil
}
