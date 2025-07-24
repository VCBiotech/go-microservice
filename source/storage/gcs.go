package storage

import (
	"context"
	"fmt"
	"io"
	"strconv"

	gcs "cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// GCSAdapter implements the Storage interface for Google Cloud Storage.
type GCSAdapter struct {
	client    *gcs.Client
	projectID string
}

// NewGCSAdapter creates a new GCSAdapter instance.
func NewGCSAdapter(projectID, credentialsFile string) (*GCSAdapter, error) {
	var opts []option.ClientOption
	if credentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(credentialsFile))
	} else {
		// Use Application Default Credentials if no file is provided
		opts = append(opts, option.WithScopes(gcs.ScopeFullControl))
	}

	client, err := gcs.NewClient(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &GCSAdapter{
		client:    client,
		projectID: projectID,
	}, nil
}

// Upload implements the Storage.Upload method for GCS.
func (a *GCSAdapter) Upload(ctx context.Context, bucket, key string, data io.Reader, size int64, metadata map[string]string) (*FileInfo, error) {
	obj := a.client.Bucket(bucket).Object(key)
	wc := obj.NewWriter(ctx)

	// Set content type and custom metadata
	wc.ContentType = metadata["Content-Type"]
	if wc.ContentType == "" {
		wc.ContentType = "application/octet-stream" // Default if not provided
	}
	wc.Metadata = metadata // GCS directly supports custom metadata

	if _, err := io.Copy(wc, data); err != nil {
		wc.Close() // Ensure writer is closed on error
		return nil, fmt.Errorf("failed to write data to GCS: %w", err)
	}
	if err := wc.Close(); err != nil {
		return nil, fmt.Errorf("failed to close GCS writer: %w", err)
	}

	attrs := wc.Attrs()
	fileInfo := &FileInfo{
		Name:           attrs.Name,
		Size:           attrs.Size,
		ContentType:    attrs.ContentType,
		LastModified:   attrs.Updated,
		ETag:           attrs.Etag,
		VersionID:      strconv.FormatInt(attrs.Generation, 10), // GCS uses generation numbers for versions
		CustomMetadata: attrs.Metadata,
		CloudProvider:  "gcp",
		StoragePath:    fmt.Sprintf("gs://%s/%s", bucket, key),
	}
	return fileInfo, nil
}

// Download implements the Storage.Download method for GCS.
func (a *GCSAdapter) Download(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	rc, err := a.client.Bucket(bucket).Object(key).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS reader: %w", err)
	}
	return rc, nil
}

// List implements the Storage.List method for GCS.
func (a *GCSAdapter) List(ctx context.Context, bucket, prefix string) ([]*FileInfo, error) {
	var files []*FileInfo
	it := a.client.Bucket(bucket).Objects(ctx, &gcs.Query{
		Prefix: prefix,
	})

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list objects in GCS: %w", err)
		}

		files = append(files, &FileInfo{
			Name:           attrs.Name,
			Size:           attrs.Size,
			ContentType:    attrs.ContentType,
			LastModified:   attrs.Updated,
			ETag:           attrs.Etag,
			VersionID:      strconv.FormatInt(attrs.Generation, 10),
			CustomMetadata: attrs.Metadata,
			CloudProvider:  "gcp",
			StoragePath:    fmt.Sprintf("gs://%s/%s", bucket, attrs.Name),
		})
	}
	return files, nil
}

// Delete implements the Storage.Delete method for GCS.
func (a *GCSAdapter) Delete(ctx context.Context, bucket, key string) error {
	err := a.client.Bucket(bucket).Object(key).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete file from GCS: %w", err)
	}
	return nil
}

// GetMetadata implements the Storage.GetMetadata method for GCS.
func (a *GCSAdapter) GetMetadata(ctx context.Context, bucket, key string) (*FileInfo, error) {
	attrs, err := a.client.Bucket(bucket).Object(key).Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get GCS object metadata: %w", err)
	}

	return &FileInfo{
		Name:           attrs.Name,
		Size:           attrs.Size,
		ContentType:    attrs.ContentType,
		LastModified:   attrs.Updated,
		ETag:           attrs.Etag,
		VersionID:      strconv.FormatInt(attrs.Generation, 10),
		CustomMetadata: attrs.Metadata,
		CloudProvider:  "gcp",
		StoragePath:    fmt.Sprintf("gs://%s/%s", bucket, key),
	}, nil
}

// UpdateMetadata implements the Storage.UpdateMetadata method for GCS.
func (a *GCSAdapter) UpdateMetadata(ctx context.Context, bucket, key string, metadata map[string]string) error {
	// GCS allows direct update of metadata
	_, err := a.client.Bucket(bucket).Object(key).Update(ctx, gcs.ObjectAttrsToUpdate{
		ContentType: metadata, // Update content type if present
		Metadata:    metadata, // Update custom metadata
	})
	if err != nil {
		return fmt.Errorf("failed to update GCS object metadata: %w", err)
	}
	return nil
}
