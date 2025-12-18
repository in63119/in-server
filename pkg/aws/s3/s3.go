package s3

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	"in-server/pkg/config"
)

type Resources struct {
	Client *s3.Client
	Bucket string
	Region string
}

func Resolve(ctx context.Context, cfg config.Config, bucketOverride string) (Resources, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	bucket := strings.TrimSpace(bucketOverride)
	if bucket == "" {
		bucket = strings.TrimSpace(cfg.AWS.S3.Bucket)
	}
	if bucket == "" {
		return Resources{}, fmt.Errorf("aws s3 bucket missing")
	}

	region := strings.TrimSpace(cfg.AWS.Region)
	if region == "" {
		region = "ap-northeast-2"
	}

	opts := []func(*awscfg.LoadOptions) error{
		awscfg.WithRegion(region),
	}

	ak := strings.TrimSpace(cfg.AWS.S3.AccessKey)
	sk := strings.TrimSpace(cfg.AWS.S3.SecretKey)
	if ak != "" && sk != "" {
		opts = append(opts, awscfg.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(ak, sk, ""),
		))
	}

	awsCfg, err := awscfg.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return Resources{}, fmt.Errorf("load aws config: %w", err)
	}

	return Resources{
		Client: s3.NewFromConfig(awsCfg),
		Bucket: bucket,
		Region: region,
	}, nil
}

func BuildObjectURL(bucket, region, key string) string {
	bucket = strings.TrimSpace(bucket)
	region = strings.TrimSpace(region)
	key = strings.TrimPrefix(strings.TrimSpace(key), "/")

	encodedKey := encodePath(key)

	if region == "" || region == "us-east-1" {
		return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, encodedKey)
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, encodedKey)
}

func PutObject(ctx context.Context, cfg config.Config, key string, body io.Reader, contentType string, bucketOverride string) (string, error) {
	res, err := Resolve(ctx, cfg, bucketOverride)
	if err != nil {
		return "", err
	}

	if _, err := res.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(res.Bucket),
		Key:         aws.String(strings.TrimPrefix(strings.TrimSpace(key), "/")),
		Body:        body,
		ContentType: aws.String(strings.TrimSpace(contentType)),
	}); err != nil {
		return "", err
	}

	return BuildObjectURL(res.Bucket, res.Region, key), nil
}

func GetObject(ctx context.Context, cfg config.Config, key string, bucketOverride string) (*s3.GetObjectOutput, error) {
	res, err := Resolve(ctx, cfg, bucketOverride)
	if err != nil {
		return nil, err
	}

	return res.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(res.Bucket),
		Key:    aws.String(strings.TrimPrefix(strings.TrimSpace(key), "/")),
	})
}

func DeleteObject(ctx context.Context, cfg config.Config, key string, bucketOverride string) error {
	res, err := Resolve(ctx, cfg, bucketOverride)
	if err != nil {
		return err
	}

	_, err = res.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(res.Bucket),
		Key:    aws.String(strings.TrimPrefix(strings.TrimSpace(key), "/")),
	})
	return err
}

func ListObjects(ctx context.Context, cfg config.Config, prefix string, bucketOverride string) ([]s3types.Object, error) {
	res, err := Resolve(ctx, cfg, bucketOverride)
	if err != nil {
		return nil, err
	}

	out, err := res.Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(res.Bucket),
		Prefix: aws.String(strings.TrimPrefix(strings.TrimSpace(prefix), "/")),
	})
	if err != nil {
		return nil, err
	}
	if out == nil || out.Contents == nil {
		return []s3types.Object{}, nil
	}
	return out.Contents, nil
}

func MomBucketName(cfg config.Config) (string, error) {
	name := strings.TrimSpace(cfg.AWS.S3.MomBucket)
	if name == "" {
		return "", fmt.Errorf("aws s3 mom bucket missing")
	}
	return name, nil
}

func PutObjectInMomBucket(ctx context.Context, cfg config.Config, key string, body io.Reader, contentType string) (string, error) {
	bucket, err := MomBucketName(cfg)
	if err != nil {
		return "", err
	}
	return PutObject(ctx, cfg, key, body, contentType, bucket)
}

func DeleteObjectInMomBucket(ctx context.Context, cfg config.Config, key string) error {
	bucket, err := MomBucketName(cfg)
	if err != nil {
		return err
	}
	return DeleteObject(ctx, cfg, key, bucket)
}

func ListObjectsInMomBucket(ctx context.Context, cfg config.Config, prefix string) ([]s3types.Object, error) {
	bucket, err := MomBucketName(cfg)
	if err != nil {
		return nil, err
	}
	return ListObjects(ctx, cfg, prefix, bucket)
}

func encodePath(key string) string {
	if key == "" {
		return ""
	}
	parts := strings.Split(key, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return strings.Join(parts, "/")
}
