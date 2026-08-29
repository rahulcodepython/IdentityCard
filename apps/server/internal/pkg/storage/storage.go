// Package storage wraps the S3-compatible object store (MinIO locally,
// any S3-compatible service in production) — currently used only for
// organization logos; user photos stay external URLs and generated PDFs
// are never persisted, per the project's data-retention decisions.
package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"identitycard-server/internal/config"
)

type Storage struct {
	client *minio.Client
	bucket string
}

func Connect(ctx context.Context, cfg *config.Config) (*Storage, error) {
	client, err := minio.New(cfg.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Secure: cfg.S3UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("storage: create client: %w", err)
	}

	exists, err := client.BucketExists(ctx, cfg.S3Bucket)
	if err != nil {
		return nil, fmt.Errorf("storage: check bucket: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.S3Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("storage: create bucket: %w", err)
		}
	}

	return &Storage{client: client, bucket: cfg.S3Bucket}, nil
}

func (s *Storage) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, r, size, minio.PutObjectOptions{ContentType: contentType})
	return err
}

// Get returns the object's bytes directly rather than a stream — logos
// are small (see the size cap in services.OrganizationsService.UploadLogo) and
// every caller needs the full bytes anyway (embed in a PDF, or serve
// with a known Content-Length).
func (s *Storage) Get(ctx context.Context, key string) ([]byte, string, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", err
	}
	defer obj.Close()

	info, err := obj.Stat()
	if err != nil {
		return nil, "", err
	}

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, "", err
	}
	return data, info.ContentType, nil
}

func (s *Storage) Delete(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}
