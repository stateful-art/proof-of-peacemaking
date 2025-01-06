package config

import (
	"fmt"
	"os"
)

// R2Config holds the configuration for a specific R2 bucket
type R2Config struct {
	// S3-compatible credentials
	S3AccessKeyID string
	S3SecretKey   string
	// Cloudflare specific
	AccountID string
	Bucket    string
	APIToken  string // Cloudflare API token, if needed for additional operations
}

// GetR2Config returns the R2 configuration for a specific bucket type
func GetR2Config(bucketType string) (*R2Config, error) {
	prefix := fmt.Sprintf("R2_%s", bucketType)

	// S3-compatible credentials
	accessKeyID := os.Getenv(prefix + "_ACCESS_KEY_ID")
	secretKey := os.Getenv(prefix + "_SECRET_KEY")

	// Cloudflare specific
	accountID := os.Getenv(prefix + "_ACCOUNT_ID")
	bucket := os.Getenv(prefix + "_BUCKET")
	apiToken := os.Getenv(prefix + "_API_TOKEN") // Optional

	// Validate required fields (API token is optional)
	if accessKeyID == "" || secretKey == "" || accountID == "" || bucket == "" {
		return nil, fmt.Errorf("missing required R2 configuration for bucket type %s", bucketType)
	}

	return &R2Config{
		S3AccessKeyID: accessKeyID,
		S3SecretKey:   secretKey,
		AccountID:     accountID,
		Bucket:        bucket,
		APIToken:      apiToken,
	}, nil
}
