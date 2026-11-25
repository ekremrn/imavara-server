package storage

import (
	"bytes"
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"os"
)

func client() (*minio.Client, error) {
	return minio.New(os.Getenv("S3_ENDPOINT"), &minio.Options{Creds: credentials.NewStaticV4(os.Getenv("S3_ACCESS_KEY"), os.Getenv("S3_SECRET_KEY"), ""), Region: "garage", BucketLookup: minio.BucketLookupPath})
}
func Put(ctx context.Context, bucket, key, mime string, data []byte) error {
	c, err := client()
	if err != nil {
		return err
	}
	_, err = c.PutObject(ctx, bucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{ContentType: mime})
	return err
}
func Get(ctx context.Context, bucket, key string) (*minio.Object, error) {
	c, err := client()
	if err != nil {
		return nil, err
	}
	return c.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
}

func Stream(ctx context.Context, bucket, key, mime string, src io.Reader, size int64) error {
	c, err := client()
	if err != nil {
		return err
	}
	_, err = c.PutObject(ctx, bucket, key, src, size, minio.PutObjectOptions{ContentType: mime})
	return err
}
