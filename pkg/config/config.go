package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Env  string `envconfig:"ENV" default:"development"`
	Port string `envconfig:"PORT" default:":4000"`

	Auth struct {
		Hash string `envconfig:"AUTH_HASH"`
		JWT  struct {
			AccessSecret string `envconfig:"AUTH_JWT_ACCESS_SECRET"`
		}
		AdminCode string `envconfig:"AUTH_ADMIN_CODE"`
	}

	AWS struct {
		AccessKey       string `envconfig:"AWS_SSM_ACCESS_KEY"`
		SecretAccessKey string `envconfig:"AWS_SSM_SECRET_KEY"`
		Region          string `envconfig:"AWS_REGION" default:"ap-northeast-2"`
		Param           string `envconfig:"AWS_SSM_SERVER"`
		S3              struct {
			Bucket    string `envconfig:"AWS_S3_BUCKET"`
			MomBucket string `envconfig:"AWS_S3_MOM_BUCKET"`
			AccessKey string `envconfig:"AWS_S3_ACCESS_KEY_ID"`
			SecretKey string `envconfig:"AWS_S3_SECRET_ACCESS_KEY"`
		}
	}

	Firebase struct {
		ProjectID   string `envconfig:"FIREBASE_PROJECT_ID"`
		ClientEmail string `envconfig:"FIREBASE_CLIENT_EMAIL"`
		PrivateKey  string `envconfig:"FIREBASE_PRIVATE_KEY"`
		DatabaseURL string `envconfig:"FIREBASE_DATABASE_URL"`
	}

	Blockchain struct {
		PrivateKey struct {
			Owner    string `envconfig:"BLOCKCHAIN_PRIVATE_KEY_OWNER"`
			Relayer  string `envconfig:"BLOCKCHAIN_PRIVATE_KEY_RELAYER"`
			Relayer2 string `envconfig:"BLOCKCHAIN_PRIVATE_KEY_RELAYER2"`
			Relayer3 string `envconfig:"BLOCKCHAIN_PRIVATE_KEY_RELAYER3"`
		}
	}

	Google struct {
		ClientKey           string `envconfig:"GOOGLE_CLIENT_KEY"`
		SecretKey           string `envconfig:"GOOGLE_SECRET_KEY"`
		RefreshToken        string `envconfig:"GOOGLE_REFRESH_TOKEN"`
		GmailSender         string `envconfig:"GOOGLE_GMAIL_SENDER"`
		RedirectURIEndpoint string `envconfig:"GOOGLE_REDIRECT_URI_ENDPOINT"`
		GeminiAPIKey        string `envconfig:"GOOGLE_GEMINI_API_KEY"`
	}
}

func Load() (Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	return cfg, err
}

func Reload(ctx context.Context) (Config, error) {
	cfg, err := Load()
	if err != nil {
		return cfg, err
	}

	region := cfg.AWS.Region
	if region == "" {
		region = "ap-northeast-2"
	}

	if strings.TrimSpace(cfg.AWS.Param) != "" {
		ssmPath := fmt.Sprintf("%s/%s", strings.TrimSuffix(cfg.AWS.Param, "/"), strings.TrimSpace(cfg.Env))
		ssmCfg := cfg
		ssmCfg.AWS.Param = ssmPath
		if err := LoadSSM(ctx, &ssmCfg); err != nil {
			return cfg, err
		}
		cfg = ssmCfg
	}

	if ads := strings.TrimSpace(getEnv("AWS_SSM_ADS")); ads != "" {
		adsCfg := cfg
		adsCfg.AWS.Param = ads
		if err := LoadSSM(ctx, &adsCfg); err != nil {
			return cfg, err
		}
		cfg = adsCfg
	}

	return cfg, nil
}

func getEnv(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

// GetEnv trims and returns the environment variable value.
func GetEnv(key string) string {
	return getEnv(key)
}
