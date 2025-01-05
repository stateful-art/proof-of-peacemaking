package storage

import (
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"time"

	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	maxRetries      = 3
	defaultCacheAge = 3600              // 1 hour in seconds
	maxUploadSize   = 100 * 1024 * 1024 // 10MB
)

type R2Storage struct {
	client   *s3.Client
	bucket   string
	uploader *manager.Uploader
}

type UploadOptions struct {
	CacheControl string
	ContentType  string
	Compress     bool
}

// NewR2Storage creates a new instance of R2Storage with optimized configuration
func NewR2Storage(accessKey, secretKey, accountID, bucket string) (*R2Storage, error) {
	r2Endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)
	r2Region := "auto"

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(r2Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load R2 configuration: %v", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(r2Endpoint)
		o.Region = r2Region
		o.UsePathStyle = true
	})

	// Initialize uploader with optimized concurrency
	uploader := manager.NewUploader(client, func(u *manager.Uploader) {
		u.Concurrency = 3
		u.PartSize = 5 * 1024 * 1024 // 5MB part size
	})

	return &R2Storage{
		client:   client,
		bucket:   bucket,
		uploader: uploader,
	}, nil
}

// getContentType determines the content type of a file
func getContentType(filename string) string {
	ext := filepath.Ext(filename)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		return "application/octet-stream"
	}
	return contentType
}

// shouldCompress determines if a file should be compressed based on its content type
func shouldCompress(contentType string) bool {
	compressibleTypes := map[string]bool{
		"text/plain":             true,
		"text/html":              true,
		"text/css":               true,
		"text/javascript":        true,
		"application/javascript": true,
		"application/json":       true,
		"application/xml":        true,
	}
	return compressibleTypes[contentType]
}

// UploadFile uploads a file to R2 storage with optimizations
func (s *R2Storage) UploadFile(ctx context.Context, key string, reader io.Reader, opts ...UploadOptions) error {
	log.Printf("[R2] Starting upload for key: %s", key)

	var uploadOpts UploadOptions
	if len(opts) > 0 {
		uploadOpts = opts[0]
		log.Printf("[R2] Upload options - ContentType: %s, CacheControl: %s", uploadOpts.ContentType, uploadOpts.CacheControl)
	}

	// Set default content type if not provided
	if uploadOpts.ContentType == "" {
		uploadOpts.ContentType = getContentType(key)
		log.Printf("[R2] Using default content type: %s", uploadOpts.ContentType)
	}

	// Set default cache control if not provided
	if uploadOpts.CacheControl == "" {
		uploadOpts.CacheControl = fmt.Sprintf("public, max-age=%d", defaultCacheAge)
		log.Printf("[R2] Using default cache control: %s", uploadOpts.CacheControl)
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),

		Body:         reader,
		ContentType:  aws.String(uploadOpts.ContentType),
		CacheControl: aws.String(uploadOpts.CacheControl),
	}

	// Use uploader for efficient multipart upload
	log.Printf("[R2] Starting multipart upload to bucket: %s", s.bucket)
	result, err := s.uploader.Upload(ctx, input)
	if err != nil {
		log.Printf("[R2] Upload failed for key %s: %v", key, err)
		return fmt.Errorf("failed to upload file: %v", err)
	}
	log.Printf("[R2] Successfully uploaded file. Key: %s, ETag: %s", key, *result.ETag)

	return nil
}

// GetFile retrieves a file from R2 storage with optimized settings
func (s *R2Storage) GetFile(ctx context.Context, key string) (io.ReadCloser, error) {
	log.Printf("[R2] Retrieving file with key: %s", key)

	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		log.Printf("[R2] Failed to get file with key %s: %v", key, err)
		return nil, fmt.Errorf("failed to get file: %v", err)
	}
	log.Printf("[R2] Successfully retrieved file. Key: %s, Size: %d", key, result.ContentLength)

	return result.Body, nil
}

// DeleteFile removes a file from R2 storage
func (s *R2Storage) DeleteFile(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}
	return nil
}

// ListFiles lists all files in the bucket with the given prefix
func (s *R2Storage) ListFiles(ctx context.Context, prefix string) ([]string, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	}

	var files []string
	paginator := s3.NewListObjectsV2Paginator(s.client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list files: %v", err)
		}
		for _, obj := range page.Contents {
			files = append(files, *obj.Key)
		}
	}
	return files, nil
}

// GetPresignedURL generates a presigned URL for direct client access
func (s *R2Storage) GetPresignedURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	log.Printf("[R2] Generating presigned URL for key: %s with expiry: %v", key, expires)

	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expires))

	if err != nil {
		log.Printf("[R2] Failed to generate presigned URL for key %s: %v", key, err)
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}
	log.Printf("[R2] Successfully generated presigned URL for key: %s", key)

	return request.URL, nil
}
