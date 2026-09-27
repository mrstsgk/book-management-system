package book

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
)

// urlTTL bounds how long a handed-out image URL stays valid; clients re-fetch the book for a fresh one.
const urlTTL = 15 * time.Minute

type Config struct {
	// Endpoint is set only for S3-compatible services such as LocalStack; empty means AWS S3.
	Endpoint        string
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
}

type imageStorage struct {
	client  *s3.Client
	presign *s3.PresignClient
	bucket  string
}

func NewImageStorage(ctx context.Context, cfg Config) (domainbook.ImageStorage, error) {
	opts := []func(*awsconfig.LoadOptions) error{awsconfig.WithRegion(cfg.Region)}
	if cfg.AccessKeyID != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			// LocalStack serves buckets by path, not by <bucket>.<host> subdomains.
			o.UsePathStyle = true
		}
	})
	return &imageStorage{client: client, presign: s3.NewPresignClient(client), bucket: cfg.Bucket}, nil
}

// Put は画像を key に保存する。署名にボディの再読込が必要なため、シークできないボディは一度メモリに読み込む（上限 5MiB）。
func (s *imageStorage) Put(ctx context.Context, key domainbook.ImageKey, image domainbook.Image, body io.Reader) error {
	if _, ok := body.(io.ReadSeeker); !ok {
		b, err := io.ReadAll(io.LimitReader(body, image.Size()+1))
		if err != nil {
			return fmt.Errorf("read image body: %w", err)
		}
		body = bytes.NewReader(b)
	}
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key.String()),
		Body:          body,
		ContentLength: aws.Int64(image.Size()),
		ContentType:   aws.String(image.ContentType()),
	})
	if err != nil {
		return fmt.Errorf("put image %s: %w", key, err)
	}
	return nil
}

func (s *imageStorage) Delete(ctx context.Context, key domainbook.ImageKey) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key.String())})
	if err != nil {
		return fmt.Errorf("delete image %s: %w", key, err)
	}
	return nil
}

func (s *imageStorage) URL(ctx context.Context, key domainbook.ImageKey) (string, error) {
	req, err := s.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key.String()),
	}, s3.WithPresignExpires(urlTTL))
	if err != nil {
		return "", fmt.Errorf("presign image %s: %w", key, err)
	}
	return req.URL, nil
}
