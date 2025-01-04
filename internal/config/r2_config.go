package config

import (
	"fmt"
	"os"
)

// R2Config holds the configuration for a specific R2 bucket
type R2Config struct {
	AccessKey string
	SecretKey string
	AccountID string
	Bucket    string
}

// GetR2Config returns the R2 configuration for a specific bucket type
func GetR2Config(bucketType string) (*R2Config, error) {
	prefix := fmt.Sprintf("R2_%s", bucketType)
	accessKey := os.Getenv(prefix + "_ACCESS_KEY")
	secretKey := os.Getenv(prefix + "_SECRET_KEY")
	accountID := os.Getenv(prefix + "_ACCOUNT_ID")
	bucket := os.Getenv(prefix + "_BUCKET")

	// Validate required fields
	if accessKey == "" || secretKey == "" || accountID == "" || bucket == "" {
		return nil, fmt.Errorf("missing required R2 configuration for bucket type %s", bucketType)
	}

	return &R2Config{
		AccessKey: accessKey,
		SecretKey: secretKey,
		AccountID: accountID,
		Bucket:    bucket,
	}, nil
}
