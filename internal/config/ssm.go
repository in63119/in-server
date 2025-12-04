package config

import (
    "context"
    "errors"
    "fmt"
    "strings"

    "github.com/aws/aws-sdk-go-v2/aws"
    awscfg "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/goccy/go-json"
)

// LoadSSM loads a JSON blob from AWS SSM (Parameter Store) stored under cfg.AWS.Param
// and merges it into cfg. This mirrors the archive-nest behavior of loading one
// JSON parameter and spreading it onto the config.
func LoadSSM(ctx context.Context, cfg *Config) error {
	if cfg.AWS.Param == "" {
		return nil
	}

	region := cfg.AWS.Region
	if region == "" {
		region = "ap-northeast-2"
	}

	opts := []func(*awscfg.LoadOptions) error{
		awscfg.WithRegion(region),
	}
	if cfg.AWS.AccessKey != "" && cfg.AWS.SecretAccessKey != "" {
		opts = append(opts, awscfg.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AWS.AccessKey, cfg.AWS.SecretAccessKey, ""),
		))
	}

	awsCfg, err := awscfg.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return fmt.Errorf("load aws config: %w", err)
	}

	client := ssm.NewFromConfig(awsCfg)
	// Support env-specific parameter names like "/in-server/production" or "/in-server/development".
	base := strings.TrimSuffix(cfg.AWS.Param, "/")
	candidates := []string{base}
	if cfg.Env != "" {
		envPath := base + "/" + cfg.Env
		candidates = append([]string{envPath}, candidates...)
	}

	var lastErr error
	for _, name := range candidates {
		out, err := client.GetParameter(ctx, &ssm.GetParameterInput{
			Name:           aws.String(name),
			WithDecryption: aws.Bool(true),
		})
		if err != nil {
			var pnf *types.ParameterNotFound
			if errors.As(err, &pnf) {
				lastErr = fmt.Errorf("ssm parameter %s not found", name)
				continue
			}
			return fmt.Errorf("get ssm parameter %s: %w", name, err)
		}
		if out.Parameter == nil || out.Parameter.Value == nil {
			lastErr = fmt.Errorf("ssm parameter %s empty", name)
			continue
		}
		return applySSMJSON(cfg, aws.ToString(out.Parameter.Value))
	}

	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("ssm parameter %v not found", candidates)
}

func applySSMJSON(cfg *Config, raw string) error {
	// Expected JSON keys mirror archive-nest:
	// {
	//   "AUTH": {"HASH": "...", "JWT": {"ACCESS_SECRET": "..."}},
	//   "AWS": {"S3": {"BUCKET": "...", "ACCESS_KEY_ID": "...", "SECRET_ACCESS_KEY": "..."}},
	//   "BLOCKCHAIN": {"PRIVATE_KEY": {"OWNER": "...", "RELAYER": "...", ...}},
	//   "FIREBASE": {"PROJECT_ID": "...", "CLIENT_EMAIL": "...", "PRIVATE_KEY": "...", "DATABASE_URL": "..."}
	// }
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return fmt.Errorf("parse ssm json: %w", err)
	}

	apply := func(path string) string {
		parts := strings.Split(path, ".")
		var cur interface{} = payload
		for _, p := range parts {
			m, ok := cur.(map[string]interface{})
			if !ok {
				return ""
			}
			cur, ok = m[p]
			if !ok {
				return ""
			}
		}
		if v, ok := cur.(string); ok {
			return v
		}
		return ""
	}

	cfg.Auth.Hash = firstNonEmpty(apply("AUTH.HASH"), cfg.Auth.Hash)
	cfg.Auth.JWT.AccessSecret = firstNonEmpty(apply("AUTH.JWT.ACCESS_SECRET"), cfg.Auth.JWT.AccessSecret)

	cfg.AWS.S3.Bucket = firstNonEmpty(apply("AWS.S3.BUCKET"), cfg.AWS.S3.Bucket)
	cfg.AWS.S3.AccessKey = firstNonEmpty(apply("AWS.S3.ACCESS_KEY_ID"), cfg.AWS.S3.AccessKey)
	cfg.AWS.S3.SecretKey = firstNonEmpty(apply("AWS.S3.SECRET_ACCESS_KEY"), cfg.AWS.S3.SecretKey)

	cfg.Blockchain.PrivateKey.Owner = firstNonEmpty(apply("BLOCKCHAIN.PRIVATE_KEY.OWNER"), cfg.Blockchain.PrivateKey.Owner)
	cfg.Blockchain.PrivateKey.Relayer = firstNonEmpty(apply("BLOCKCHAIN.PRIVATE_KEY.RELAYER"), cfg.Blockchain.PrivateKey.Relayer)
	cfg.Blockchain.PrivateKey.Relayer2 = firstNonEmpty(apply("BLOCKCHAIN.PRIVATE_KEY.RELAYER2"), cfg.Blockchain.PrivateKey.Relayer2)
	cfg.Blockchain.PrivateKey.Relayer3 = firstNonEmpty(apply("BLOCKCHAIN.PRIVATE_KEY.RELAYER3"), cfg.Blockchain.PrivateKey.Relayer3)

	cfg.Firebase.ProjectID = firstNonEmpty(apply("FIREBASE.PROJECT_ID"), cfg.Firebase.ProjectID)
	cfg.Firebase.ClientEmail = firstNonEmpty(apply("FIREBASE.CLIENT_EMAIL"), cfg.Firebase.ClientEmail)
	cfg.Firebase.PrivateKey = firstNonEmpty(apply("FIREBASE.PRIVATE_KEY"), cfg.Firebase.PrivateKey)
	cfg.Firebase.DatabaseURL = firstNonEmpty(apply("FIREBASE.DATABASE_URL"), cfg.Firebase.DatabaseURL)

	return nil
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}
