package database

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"strings"
)

func ConnectMinio(ctx context.Context, endpoint, accessKey, secretKey string) (*minio.Client, error) {
	ep := strings.TrimSpace(endpoint)

	cli, err := minio.New(ep, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}

	if _, err := cli.ListBuckets(ctx); err != nil {
		return nil, err
	}

	return cli, nil
}

func EnsureBucket(ctx context.Context, cli *minio.Client, bucket string) error {
	exists, err := cli.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if !exists {
		return cli.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
	}
	return nil
}
