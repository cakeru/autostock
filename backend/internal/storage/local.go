package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cakeru/autostock/internal/config"
)

// localStore writes files to a directory on disk (a mounted volume in
// production) and serves them under a URL prefix.
type localStore struct {
	dir       string
	publicURL string
}

func newLocalStore(cfg config.StorageConfig) (Storage, error) {
	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}
	return &localStore{
		dir:       cfg.UploadDir,
		publicURL: strings.TrimRight(cfg.PublicURL, "/"),
	}, nil
}

func (s *localStore) Save(_ context.Context, key string, r io.Reader, _ string) (string, error) {
	dest := filepath.Join(s.dir, filepath.Clean("/"+key))
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", fmt.Errorf("mkdir: %w", err)
	}
	f, err := os.Create(dest)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}
	return s.publicURL + "/" + key, nil
}

func (s *localStore) Delete(_ context.Context, url string) error {
	prefix := s.publicURL + "/"
	if !strings.HasPrefix(url, prefix) {
		return nil // not one of ours (e.g. an external seed URL)
	}
	key := strings.TrimPrefix(url, prefix)
	err := os.Remove(filepath.Join(s.dir, filepath.Clean("/"+key)))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
