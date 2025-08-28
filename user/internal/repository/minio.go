package repository

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/s3utils"
)

type Minio struct {
	Client  *minio.Client
	Bucket  string        // например: "users"
	BaseURL string        // например: "http://user_minio:9000" (опц., для формирования прямого URL)
	Expiry  time.Duration // срок presigned URL, если BaseURL пуст
}

func NewMinio(client *minio.Client, bucket, baseURL string) *Minio {
	return &Minio{
		Client:  client,
		Bucket:  bucket,
		BaseURL: strings.TrimRight(baseURL, "/"),
		Expiry:  24 * time.Hour,
	}
}

// (опционально) создать бакет, если отсутствует
func (m *Minio) EnsureBucket(ctx context.Context) error {
	exists, err := m.Client.BucketExists(ctx, m.Bucket)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return m.Client.MakeBucket(ctx, m.Bucket, minio.MakeBucketOptions{})
}

// Upload загружает фото и возвращает URL.
// Если указан BaseURL — вернёт "BaseURL/bucket/key" (path-style).
// Если BaseURL пуст — вернёт presigned GET URL с m.Expiry.
func (m *Minio) Upload(ctx context.Context, userID int64, r io.Reader) (string, error) {
	if m.Client == nil || m.Bucket == "" {
		return "", fmt.Errorf("minio: not configured (client or bucket is empty)")
	}

	// читаем в память, чтобы знать размер и content-type
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", fmt.Errorf("empty file")
	}

	key := fmt.Sprintf("users/%d/%d.jpg", userID, time.Now().UnixNano())

	// валидируем путь (рекомендуется minio-go)
	if err := s3utils.CheckValidObjectName(key); err != nil {
		return "", fmt.Errorf("invalid object name: %w", err)
	}

	ct := http.DetectContentType(data)
	if ct == "application/octet-stream" {
		ct = "image/jpeg" // дефолт
	}

	// загрузка
	_, err = m.Client.PutObject(ctx, m.Bucket, key, bytes.NewReader(data), int64(len(data)),
		minio.PutObjectOptions{
			ContentType:  ct,
			StorageClass: "", // можно оставить пустым
		})
	if err != nil {
		return "", err
	}

	// 1) если задан BaseURL — возвращаем прямой path-style URL
	if m.BaseURL != "" {
		return fmt.Sprintf("%s/%s/%s", m.BaseURL, m.Bucket, key), nil
	}

	// 2) иначе — presigned GET URL (удобно, если MinIO не публичен)
	u, err := m.Client.PresignedGetObject(ctx, m.Bucket, key, m.Expiry, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
