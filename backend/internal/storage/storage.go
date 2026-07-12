package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/cakeru/autostock/internal/config"
)

// Storage abstracts where uploaded files live. Save returns the final,
// ready-to-serve URL that callers persist (e.g. products.image_url); the rest
// of the app depends only on this interface, so switching drivers via
// STORAGE_DRIVER never touches handlers or the database.
type Storage interface {
	// Save writes r under key and returns the public URL for it.
	Save(ctx context.Context, key string, r io.Reader, contentType string) (string, error)
	// Delete removes the object previously stored at the given public URL.
	// A missing object is not an error.
	Delete(ctx context.Context, url string) error
}

// New selects a driver from config. Defaults to local.
func New(cfg config.StorageConfig) (Storage, error) {
	switch cfg.Driver {
	case "", "local":
		return newLocalStore(cfg)
	case "r2":
		return newR2Store(cfg)
	case "s3":
		return newS3Store(cfg)
	default:
		return nil, fmt.Errorf("unknown STORAGE_DRIVER %q (expected \"local\", \"r2\" or \"s3\")", cfg.Driver)
	}
}
