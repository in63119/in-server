package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"

	"in-server/internal/config"
	"in-server/internal/server"
	"in-server/pkg/logger"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: .env not loaded: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if err := config.LoadSSM(context.Background(), &cfg); err != nil {
		log.Fatalf("load ssm config: %v", err)
	}
	// log.Printf("config loaded (raw): env=%s port=%s aws.region=%s aws.param=%s aws.access_key=%s aws.secret_access_key=%s aws.s3.bucket=%s aws.s3.access_key=%s aws.s3.secret_key=%s auth.hash=%s auth.jwt_access=%s firebase.project_id=%s firebase.client_email=%s firebase.private_key=%s firebase.database_url=%s blockchain.owner=%s blockchain.relayer=%s blockchain.relayer2=%s blockchain.relayer3=%s",
	// 	cfg.Env,
	// 	cfg.Port,
	// 	cfg.AWS.Region,
	// 	cfg.AWS.Param,
	// 	cfg.AWS.AccessKey,
	// 	cfg.AWS.SecretAccessKey,
	// 	cfg.AWS.S3.Bucket,
	// 	cfg.AWS.S3.AccessKey,
	// 	cfg.AWS.S3.SecretKey,
	// 	cfg.Auth.Hash,
	// 	cfg.Auth.JWT.AccessSecret,
	// 	cfg.Firebase.ProjectID,
	// 	cfg.Firebase.ClientEmail,
	// 	cfg.Firebase.PrivateKey,
	// 	cfg.Firebase.DatabaseURL,
	// 	cfg.Blockchain.PrivateKey.Owner,
	// 	cfg.Blockchain.PrivateKey.Relayer,
	// 	cfg.Blockchain.PrivateKey.Relayer2,
	// 	cfg.Blockchain.PrivateKey.Relayer3,
	// )

	logg, err := logger.New(cfg.Env)
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer func() {
		_ = logg.Sync()
	}()

	srv := server.New(cfg, logg)

	if err := srv.Run(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
