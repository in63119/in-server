package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"

	"in-server/pkg/config"
)

const (
	abisPrefix    = "abis/"
	localBasePath = "."
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("warning: .env not loaded: %v", err)
	}

	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if err := config.LoadSSM(ctx, &cfg); err != nil {
		log.Fatalf("load ssm: %v", err)
	}

	region := cfg.AWS.Region
	if region == "" {
		region = "ap-northeast-2"
	}
	if cfg.AWS.S3.Bucket == "" {
		log.Fatalf("aws s3 bucket missing (check SSM or env)")
	}

	// Prefer S3-specific access keys, fallback to general AWS keys.
	accessKey := firstNonEmpty(cfg.AWS.S3.AccessKey, cfg.AWS.AccessKey)
	secretKey := firstNonEmpty(cfg.AWS.S3.SecretKey, cfg.AWS.SecretAccessKey)

	opts := []func(*awscfg.LoadOptions) error{
		awscfg.WithRegion(region),
	}
	if accessKey != "" && secretKey != "" {
		opts = append(opts, awscfg.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		))
	}

	awsCfg, err := awscfg.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		log.Fatalf("load aws config: %v", err)
	}

	client := s3.NewFromConfig(awsCfg)
	if err := syncAbis(ctx, client, cfg.AWS.S3.Bucket); err != nil {
		log.Fatalf("sync abis: %v", err)
	}
}

func syncAbis(ctx context.Context, client *s3.Client, bucket string) error {
	keys, err := listObjects(ctx, client, bucket, abisPrefix)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		log.Printf("no ABI files found in s3://%s/%s", bucket, abisPrefix)
		return nil
	}

	for _, key := range keys {
		if key == "" || strings.HasSuffix(key, "/") {
			continue
		}

		body, err := downloadObject(ctx, client, bucket, key)
		if err != nil {
			log.Printf("skip %s: %v", key, err)
			continue
		}

		dest := filepath.Join(localBasePath, key)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			log.Printf("mkdir %s: %v", filepath.Dir(dest), err)
			continue
		}

		if err := os.WriteFile(dest, body, 0o644); err != nil {
			log.Printf("write %s: %v", dest, err)
			continue
		}

		log.Printf("saved ABI file: %s", dest)
	}

	return nil
}

func listObjects(ctx context.Context, client *s3.Client, bucket, prefix string) ([]string, error) {
	var keys []string
	p := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list objects: %w", err)
		}
		for _, obj := range page.Contents {
			if obj.Key != nil && *obj.Key != "" {
				keys = append(keys, *obj.Key)
			}
		}
	}
	return keys, nil
}

func downloadObject(ctx context.Context, client *s3.Client, bucket, key string) ([]byte, error) {
	resp, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("get object %s: %w", key, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read object %s: %w", key, err)
	}
	return data, nil
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}
