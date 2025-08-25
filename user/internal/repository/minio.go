package repository

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Minio struct {
	Client     *s3.Client
	BucketName string
}

// Конструктор
func NewMinio(client *s3.Client, bucketName string) *Minio {
	return &Minio{
		Client:     client,
		BucketName: bucketName,
	}
}

// Реализация интерфейса PhotoUploader
func (u *Minio) Upload(userID int64, file io.Reader) (string, error) {
	key := fmt.Sprintf("users/%d/%d.jpg", userID, time.Now().UnixNano())

	_, err := u.Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(u.BucketName),
		Key:    aws.String(key),
		Body:   file,
		ACL:    "public-read",
	})
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", u.BucketName, key)
	return url, nil
}
