package main

import (
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
