package config

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/goccy/go-json"
)

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
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return fmt.Errorf("parse ssm json: %w", err)
	}

	apply := func(path string) string {
		parts := strings.Split(path, ".")
		var cur any = payload
		for _, p := range parts {
			m, ok := cur.(map[string]any)
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
	cfg.Auth.AdminCode = firstNonEmpty(apply("AUTH.ADMIN_CODE"), cfg.Auth.AdminCode)

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

	cfg.Google.ClientKey = firstNonEmpty(apply("GOOGLE.CLIENT_KEY"), cfg.Google.ClientKey)
	cfg.Google.SecretKey = firstNonEmpty(apply("GOOGLE.SECRET_KEY"), cfg.Google.SecretKey)
	cfg.Google.RefreshToken = firstNonEmpty(apply("GOOGLE.REFRESH_TOKEN"), cfg.Google.RefreshToken)
	cfg.Google.GmailSender = firstNonEmpty(apply("GOOGLE.GMAIL_SENDER"), cfg.Google.GmailSender)
	cfg.Google.RedirectURIEndpoint = firstNonEmpty(apply("GOOGLE.REDIRECT_URI_ENDPOINT"), cfg.Google.RedirectURIEndpoint)
	cfg.Google.GeminiAPIKey = firstNonEmpty(apply("GOOGLE.GEMINI_API_KEY"), cfg.Google.GeminiAPIKey)

	return nil
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

type SaveSSMInput struct {
	AccessKey       string
	SecretAccessKey string
	Region          string
	Param           string
	Value           *string
	Patch           map[string]any
	Overwrite       *bool
}

func SaveSSM(ctx context.Context, in SaveSSMInput) error {
	if strings.TrimSpace(in.AccessKey) == "" || strings.TrimSpace(in.SecretAccessKey) == "" {
		return fmt.Errorf("aws accessKey or secretAccessKey not found")
	}
	if strings.TrimSpace(in.Param) == "" {
		return fmt.Errorf("ssm parameter name is required")
	}
	if in.Value == nil && in.Patch == nil {
		return fmt.Errorf("either value or patch must be provided")
	}

	region := strings.TrimSpace(in.Region)
	if region == "" {
		region = "ap-northeast-2"
	}

	awsCfg, err := awscfg.LoadDefaultConfig(ctx,
		awscfg.WithRegion(region),
		awscfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(in.AccessKey, in.SecretAccessKey, "")),
	)
	if err != nil {
		return fmt.Errorf("load aws config: %w", err)
	}

	client := ssm.NewFromConfig(awsCfg)

	finalValue := ""
	if in.Value != nil {
		finalValue = *in.Value
	}

	if in.Patch != nil {
		getOut, err := client.GetParameter(ctx, &ssm.GetParameterInput{
			Name:           aws.String(in.Param),
			WithDecryption: aws.Bool(false),
		})
		existing := map[string]any{}
		if err == nil && getOut != nil && getOut.Parameter != nil && getOut.Parameter.Value != nil {
			var parsed map[string]any
			if e := json.Unmarshal([]byte(aws.ToString(getOut.Parameter.Value)), &parsed); e == nil && parsed != nil {
				existing = parsed
			}
		} else if err != nil {
			log.Printf("saveSSM: failed to read existing parameter %s, proceeding with patch only: %v", in.Param, err)
		}

		applyPatch(existing, in.Patch)

		merged, err := json.Marshal(existing)
		if err != nil {
			return fmt.Errorf("marshal patched value: %w", err)
		}
		finalValue = string(merged)
	}

	overwrite := true
	if in.Overwrite != nil {
		overwrite = *in.Overwrite
	}

	_, err = client.PutParameter(ctx, &ssm.PutParameterInput{
		Name:      aws.String(in.Param),
		Value:     aws.String(finalValue),
		Type:      types.ParameterTypeString,
		Overwrite: aws.Bool(overwrite),
	})
	if err != nil {
		return fmt.Errorf("put parameter: %w", err)
	}

	return nil
}

func applyPatch(dst map[string]any, patch map[string]any) {
	if dst == nil {
		return
	}
	for k, v := range patch {
		if strings.Contains(k, ".") {
			setNested(dst, k, v)
			delete(dst, k)
			continue
		}

		if vMap, ok := v.(map[string]any); ok {
			if existMap, ok2 := dst[k].(map[string]any); ok2 && existMap != nil {
				applyPatch(existMap, vMap)
				dst[k] = existMap
				continue
			}
		}
		dst[k] = v
	}
}

func setNested(dst map[string]any, dotted string, val any) {
	parts := strings.Split(dotted, ".")
	cur := dst
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if i == len(parts)-1 {
			cur[part] = val
			return
		}
		next, ok := cur[part].(map[string]any)
		if !ok || next == nil {
			next = map[string]any{}
			cur[part] = next
		}
		cur = next
	}
}
