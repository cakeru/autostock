package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	JWTExpiry   time.Duration

	LogLevel string
	AppEnv   string

	TelegramEnabled bool
	TelegramToken   string
	TelegramChatID  string

	// BackupDir is the host folder the nightly database dumps land in (mounted
	// read-only into the backend at /backups); the Settings page streams the
	// latest one for download.
	BackupDir string

	Storage StorageConfig
}

type StorageConfig struct {
	Driver    string // "local" (default), "r2", or "s3"
	UploadDir string // local: filesystem dir for uploads
	PublicURL string // local: URL prefix stored image paths are served under

	R2AccountID string
	R2Bucket    string
	R2AccessKey string
	R2SecretKey string
	R2PublicURL string // public/CDN base URL for the bucket

	// Generic S3-compatible driver — works with AWS S3, Backblaze B2,
	// DigitalOcean Spaces, MinIO, etc. (R2 is just a preset of this.)
	S3Endpoint       string // e.g. https://s3.amazonaws.com or https://minio.local:9000; "" = AWS default
	S3Region         string
	S3Bucket         string
	S3AccessKey      string
	S3SecretKey      string
	S3PublicURL      string // public/CDN base URL objects are served from
	S3ForcePathStyle bool   // true for MinIO and other path-style providers
}

func Load() *Config {
	cfg := &Config{
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		JWTSecret:       getEnv("JWT_SECRET", "change-me-in-production-min-32-chars"),
		JWTExpiry:       time.Duration(getEnvInt("JWT_EXPIRY_HOURS", 24)) * time.Hour,
		LogLevel:        getEnv("LOG_LEVEL", "debug"),
		AppEnv:          getEnv("APP_ENV", "development"),
		BackupDir:       getEnv("BACKUP_DIR", "./backups"),
		TelegramEnabled: getEnv("TELEGRAM_ENABLED", "false") == "true",
		TelegramToken:   getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramChatID:  getEnv("TELEGRAM_CHAT_ID", ""),
		Storage: StorageConfig{
			Driver:      getEnv("STORAGE_DRIVER", "local"),
			UploadDir:   getEnv("UPLOAD_DIR", "/app/uploads"),
			PublicURL:   getEnv("STORAGE_PUBLIC_URL", "/uploads"),
			R2AccountID: getEnv("R2_ACCOUNT_ID", ""),
			R2Bucket:    getEnv("R2_BUCKET", ""),
			R2AccessKey: getEnv("R2_ACCESS_KEY_ID", ""),
			R2SecretKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
			R2PublicURL: getEnv("R2_PUBLIC_URL", ""),

			S3Endpoint:       getEnv("S3_ENDPOINT", ""),
			S3Region:         getEnv("S3_REGION", "auto"),
			S3Bucket:         getEnv("S3_BUCKET", ""),
			S3AccessKey:      getEnv("S3_ACCESS_KEY_ID", ""),
			S3SecretKey:      getEnv("S3_SECRET_ACCESS_KEY", ""),
			S3PublicURL:      getEnv("S3_PUBLIC_URL", ""),
			S3ForcePathStyle: getEnv("S3_FORCE_PATH_STYLE", "false") == "true",
		},
	}

	if cfg.AppEnv == "production" && cfg.JWTSecret == "change-me-in-production-min-32-chars" {
		panic("JWT_SECRET must be set to a non-default value in production")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
