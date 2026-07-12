package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/cakeru/autostock/internal/config"
)

// s3Store keeps files in any S3-compatible bucket (AWS S3, Cloudflare R2,
// Backblaze B2, DigitalOcean Spaces, MinIO, …). Objects are served from the
// bucket's public/CDN domain, so Save returns a fully-qualified URL the frontend
// loads directly. Both the "s3" and "r2" drivers share this implementation.
type s3Store struct {
	client    *s3.Client
	uploader  *manager.Uploader
	bucket    string
	publicURL string
}

// s3Params is the provider-agnostic set of knobs needed to reach a bucket.
type s3Params struct {
	endpoint       string // "" uses the SDK's default AWS endpoints
	region         string
	bucket         string
	accessKey      string
	secretKey      string
	publicURL      string
	forcePathStyle bool // path-style addressing (MinIO and some others need it)
}

func newS3Compatible(p s3Params) (Storage, error) {
	if p.bucket == "" || p.accessKey == "" || p.secretKey == "" || p.publicURL == "" {
		return nil, fmt.Errorf("s3 storage requires a bucket, access key, secret key and public URL")
	}
	region := p.region
	if region == "" {
		region = "auto"
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(p.accessKey, p.secretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if p.endpoint != "" {
			o.BaseEndpoint = aws.String(p.endpoint)
		}
		if p.forcePathStyle {
			o.UsePathStyle = true
		}
	})

	return &s3Store{
		client:    client,
		uploader:  manager.NewUploader(client),
		bucket:    p.bucket,
		publicURL: strings.TrimRight(p.publicURL, "/"),
	}, nil
}

// newS3Store builds the generic S3 driver from the S3_* environment.
func newS3Store(cfg config.StorageConfig) (Storage, error) {
	return newS3Compatible(s3Params{
		endpoint:       cfg.S3Endpoint,
		region:         cfg.S3Region,
		bucket:         cfg.S3Bucket,
		accessKey:      cfg.S3AccessKey,
		secretKey:      cfg.S3SecretKey,
		publicURL:      cfg.S3PublicURL,
		forcePathStyle: cfg.S3ForcePathStyle,
	})
}

func (s *s3Store) Save(ctx context.Context, key string, r io.Reader, contentType string) (string, error) {
	_, err := s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        r,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("s3 upload: %w", err)
	}
	return s.publicURL + "/" + key, nil
}

func (s *s3Store) Delete(ctx context.Context, url string) error {
	prefix := s.publicURL + "/"
	if !strings.HasPrefix(url, prefix) {
		return nil // not one of ours
	}
	key := strings.TrimPrefix(url, prefix)
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}
