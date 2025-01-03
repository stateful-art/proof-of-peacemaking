package storage

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	maxRetries      = 3
	defaultCacheAge = 3600             // 1 hour in seconds
	maxUploadSize   = 10 * 1024 * 1024 // 10MB
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
func NewR2Storage() (*R2Storage, error) {
	r2AccessKey := os.Getenv("R2_ACCESS_KEY")
	r2SecretKey := os.Getenv("R2_SECRET_KEY")
	r2AccountID := os.Getenv("R2_ACCOUNT_ID")
	r2Bucket := os.Getenv("R2_BUCKET")
	r2Endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", r2AccountID)
	r2Region := "auto"

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(r2Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(r2AccessKey, r2SecretKey, "")),
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
		bucket:   r2Bucket,
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
	var uploadOpts UploadOptions
	if len(opts) > 0 {
		uploadOpts = opts[0]
	}

	// Set default content type if not provided
	if uploadOpts.ContentType == "" {
		uploadOpts.ContentType = getContentType(key)
	}

	// Set default cache control if not provided
	if uploadOpts.CacheControl == "" {
		uploadOpts.CacheControl = fmt.Sprintf("public, max-age=%d", defaultCacheAge)
	}

	input := &s3.PutObjectInput{
		Bucket:       aws.String(s.bucket),
		Key:          aws.String(key),
		Body:         reader,
		ContentType:  aws.String(uploadOpts.ContentType),
		CacheControl: aws.String(uploadOpts.CacheControl),
	}

	// Use uploader for efficient multipart upload
	_, err := s.uploader.Upload(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}

	return nil
}

// GetFile retrieves a file from R2 storage with optimized settings
func (s *R2Storage) GetFile(ctx context.Context, key string) (io.ReadCloser, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %v", err)
	}

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
	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expires))

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}

	return request.URL, nil
}
