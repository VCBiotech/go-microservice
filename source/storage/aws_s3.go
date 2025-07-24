package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// AWSS3Adapter implements the Storage interface for AWS S3.
type AWSS3Adapter struct {
	client     *s3.Client
	uploader   *manager.Uploader
	downloader *manager.Downloader
	region     string
}

// NewAWSS3Adapter creates a new AWSS3Adapter instance.
func NewAWSS3Adapter(region, accessKeyID, secretAccessKey string) (*AWSS3Adapter, error) {
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithRegion(region),
		awsConfig.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     accessKeyID,
				SecretAccessKey: secretAccessKey,
			}, nil
		})),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS SDK config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)
	downloader := manager.NewDownloader(client)

	return &AWSS3Adapter{
		client:     client,
		uploader:   uploader,
		downloader: downloader,
		region:     region,
	}, nil
}

// Upload implements the Storage.Upload method for AWS S3.
func (a *AWSS3Adapter) Upload(ctx context.Context, bucket, key string, data io.Reader, size int64, metadata map[string]string) (*FileInfo, error) {
	// Convert custom metadata to S3-compatible format (x-amz-meta-)
	s3Metadata := make(map[string]string)
	for k, v := range metadata {
		s3Metadata["x-amz-meta-"+k] = v
	}

	uploadInput := &s3.PutObjectInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		Body:     data,
		Metadata: s3Metadata,
	}

	// For large files, use the uploader which handles multipart uploads automatically
	uploadOutput, err := a.uploader.Upload(ctx, uploadInput)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to S3: %w", err)
	}

	fileInfo := &FileInfo{
		Name:           key,
		Size:           size,                     // Note: S3 PutObjectInput doesn't directly take size, but uploader handles it.
		ContentType:    metadata["Content-Type"], // Assuming Content-Type is passed in metadata
		ETag:           *uploadOutput.ETag,
		VersionID:      aws.ToString(uploadOutput.VersionID), // Will be nil if versioning is not enabled
		CustomMetadata: metadata,
		CloudProvider:  "aws",
		StoragePath:    fmt.Sprintf("s3://%s/%s", bucket, key),
	}
	return fileInfo, nil
}

// Download implements the Storage.Download method for AWS S3.
func (a *AWSS3Adapter) Download(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	buff := manager.NewWriteAtBuffer([]byte{})
	_, err := a.downloader.Download(ctx, buff, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file from S3: %w", err)
	}
	return io.NopCloser(bytes.NewReader(buff.Bytes())), nil
}

// List implements the Storage.List method for AWS S3.
func (a *AWSS3Adapter) List(ctx context.Context, bucket, prefix string) ([]*FileInfo, error) {
	var files []*FileInfo
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}

	paginator := s3.NewListObjectsV2Paginator(a.client, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects in S3: %w", err)
		}
		for _, obj := range page.Contents {
			// Fetch full metadata for each object (ListObjectsV2 doesn't return all metadata)
			headOutput, err := a.client.HeadObject(ctx, &s3.HeadObjectInput{
				Bucket: aws.String(bucket),
				Key:    obj.Key,
			})
			if err != nil {
				log.Printf("Warning: Failed to get metadata for S3 object %s/%s: %v", bucket, *obj.Key, err)
				continue
			}

			customMeta := make(map[string]string)
			for k, v := range headOutput.Metadata {
				customMeta[k] = v // S3 returns custom metadata with original keys
			}

			files = append(files, &FileInfo{
				Name:           *obj.Key,
				Size:           *obj.Size,
				ContentType:    aws.ToString(headOutput.ContentType),
				LastModified:   aws.ToTime(obj.LastModified),
				ETag:           aws.ToString(obj.ETag),
				VersionID:      aws.ToString(headOutput.VersionId),
				CustomMetadata: customMeta,
				CloudProvider:  "aws",
				StoragePath:    fmt.Sprintf("s3://%s/%s", bucket, *obj.Key),
			})
		}
	}
	return files, nil
}

// Delete implements the Storage.Delete method for AWS S3.
func (a *AWSS3Adapter) Delete(ctx context.Context, bucket, key string) error {
	_, err := a.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}
	return nil
}

// GetMetadata implements the Storage.GetMetadata method for AWS S3.
func (a *AWSS3Adapter) GetMetadata(ctx context.Context, bucket, key string) (*FileInfo, error) {
	headOutput, err := a.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get S3 object metadata: %w", err)
	}

	customMeta := make(map[string]string)
	for k, v := range headOutput.Metadata {
		customMeta[k] = v
	}

	return &FileInfo{
		Name:           key,
		Size:           *headOutput.ContentLength,
		ContentType:    aws.ToString(headOutput.ContentType),
		LastModified:   aws.ToTime(headOutput.LastModified),
		ETag:           aws.ToString(headOutput.ETag),
		VersionID:      aws.ToString(headOutput.VersionId),
		CustomMetadata: customMeta,
		CloudProvider:  "aws",
		StoragePath:    fmt.Sprintf("s3://%s/%s", bucket, key),
	}, nil
}

// UpdateMetadata implements the Storage.UpdateMetadata method for AWS S3.
// Note: S3 does not allow direct update of metadata without rewriting the object.
// This implementation performs a copy operation to update metadata.
func (a *AWSS3Adapter) UpdateMetadata(ctx context.Context, bucket, key string, metadata map[string]string) error {
	// Get existing metadata
	existingMeta, err := a.GetMetadata(ctx, bucket, key)
	if err != nil {
		return fmt.Errorf("failed to get existing S3 object metadata for update: %w", err)
	}

	// Merge existing custom metadata with new metadata
	newS3Metadata := make(map[string]string)
	for k, v := range existingMeta.CustomMetadata {
		newS3Metadata[k] = v
	}
	for k, v := range metadata {
		newS3Metadata[k] = v
	}

	// S3 requires a copy operation to update metadata
	_, err = a.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:            aws.String(bucket),
		Key:               aws.String(key),
		CopySource:        aws.String(fmt.Sprintf("/%s/%s", bucket, key)),
		Metadata:          newS3Metadata,
		MetadataDirective: types.MetadataDirectiveReplace, // Replace existing metadata
	})
	if err != nil {
		return fmt.Errorf("failed to update S3 object metadata via copy: %w", err)
	}
	return nil
}
