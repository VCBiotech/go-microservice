package storage

import (
	"context"
	"io"
	"time"
)

// FileInfo represents metadata about a file/object
type FileInfo struct {
	Name           string
	Size           int64
	ContentType    string
	LastModified   time.Time
	ETag           string            // Entity tag for optimistic concurrency control
	VersionID      string            // Cloud-specific version ID (e.g., S3 version ID, GCS generation number)
	CustomMetadata map[string]string // Custom metadata from the cloud provider
	CloudProvider  string            // Which cloud provider stores this file
	StoragePath    string            // The actual path/key in the cloud storage
}

// Storage defines the generic interface for multi-cloud file operations.
type Storage interface {
	// Upload uploads a file/object to the specified bucket/container.
	// `data` is the content of the file, `size` is its length.
	// `metadata` can include content type, custom headers, etc.
	// Returns the cloud-specific object key/path and an error.
	Upload(ctx context.Context, bucket, key string, data io.Reader, size int64, metadata map[string]string) (*FileInfo, error)

	// Download retrieves a file/object from the specified bucket/container.
	// Returns an io.ReadCloser for streaming the content.
	Download(ctx context.Context, bucket, key string) (io.ReadCloser, error)

	// List lists files/objects within a specified bucket/container with an optional prefix.
	// Returns a slice of FileInfo.
	List(ctx context.Context, bucket, prefix string) ([]*FileInfo, error)

	// Delete removes a file/object from the specified bucket/container.
	Delete(ctx context.Context, bucket, key string) error

	// GetMetadata retrieves metadata for a specific file/object.
	GetMetadata(ctx context.Context, bucket, key string) (*FileInfo, error)

	// UpdateMetadata updates metadata for a specific file/object.
	UpdateMetadata(ctx context.Context, bucket, key string, metadata map[string]string) error
}
