package config

import "github.com/kelseyhightower/envconfig"

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
}

func Load() (Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	return cfg, err
}
