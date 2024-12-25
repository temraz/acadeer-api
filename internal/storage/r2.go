package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Client struct {
	client     *s3.Client
	bucketName string
	publicURL  string
}

func NewR2Client(accountID, accessKeyID, accessKeySecret, bucketName string) (*R2Client, error) {
	if bucketName == "" {
		return nil, fmt.Errorf("bucket name is required")
	}

	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID),
			SigningRegion: "auto",
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, accessKeySecret, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %v", err)
	}

	client := s3.NewFromConfig(cfg)

	return &R2Client{
		client:     client,
		bucketName: bucketName,
		publicURL:  fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID),
	}, nil
}

func (r *R2Client) UploadFile(ctx context.Context, file *multipart.FileHeader, folder string) (string, error) {
	if file == nil {
		return "", fmt.Errorf("file is required")
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("unable to open file: %v", err)
	}
	defer src.Close()

	// Generate unique filename
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
	key := fmt.Sprintf("%s/%s", folder, filename)

	// Upload to R2
	_, err = r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(key),
		Body:        src,
		ContentType: aws.String(file.Header.Get("Content-Type")),
	})
	if err != nil {
		return "", fmt.Errorf("unable to upload file: %v", err)
	}

	// Return only the file key
	return key, nil
}

// GetFileURL returns a presigned URL for the file
func (r *R2Client) GetFileURL(key string) string {
	if key == "" {
		return ""
	}

	presignClient := s3.NewPresignClient(r.client)

	// Create the presigned URL with a 7-day expiration
	presignResult, err := presignClient.PresignGetObject(context.TODO(),
		&s3.GetObjectInput{
			Bucket: aws.String(r.bucketName),
			Key:    aws.String(key),
		},
		func(opts *s3.PresignOptions) {
			opts.Expires = time.Hour * 24 * 7 // 7 days
		},
	)
	if err != nil {
		return ""
	}

	return presignResult.URL
}

func (r *R2Client) DeleteFile(ctx context.Context, key string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("unable to delete file: %v", err)
	}
	return nil
}

func (r *R2Client) GetFile(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := r.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

func (r *R2Client) GetKeyFromURL(url string) string {
	// Remove the base URL part to get the key
	return strings.TrimPrefix(url, r.publicURL+"/")
}
