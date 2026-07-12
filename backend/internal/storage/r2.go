package storage

import (
	"fmt"

	"github.com/cakeru/autostock/internal/config"
)

// newR2Store is a preset of the generic S3 driver for Cloudflare R2: the
// endpoint is derived from the account id and the region is R2's fixed "auto".
// Kept as its own driver name for backward-compatible R2_* configuration.
func newR2Store(cfg config.StorageConfig) (Storage, error) {
	if cfg.R2AccountID == "" || cfg.R2Bucket == "" || cfg.R2AccessKey == "" || cfg.R2SecretKey == "" || cfg.R2PublicURL == "" {
		return nil, fmt.Errorf("storage driver r2 requires R2_ACCOUNT_ID, R2_BUCKET, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY and R2_PUBLIC_URL")
	}
	return newS3Compatible(s3Params{
		endpoint:  fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2AccountID),
		region:    "auto",
		bucket:    cfg.R2Bucket,
		accessKey: cfg.R2AccessKey,
		secretKey: cfg.R2SecretKey,
		publicURL: cfg.R2PublicURL,
	})
}
