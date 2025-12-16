package server

import (
	"time"

	"in-server/pkg/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	cfg    config.Config
	engine *gin.Engine
	log    *zap.Logger
}

func New(cfg config.Config, log *zap.Logger) *Server {
	setGinMode(cfg.Env)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(gin.Logger())
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowOriginFunc:  func(origin string) bool { return true }, // dev-friendly: allow any origin
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	s := &Server{
		cfg:    cfg,
		engine: engine,
		log:    log,
	}
	s.registerRoutes()
	return s
}

func (s *Server) Run() error {
	s.log.Info("starting http server", zap.String("addr", s.cfg.Port))
	return s.engine.Run(s.cfg.Port)
}

func setGinMode(env string) {
	switch env {
	case "production":
		gin.SetMode(gin.ReleaseMode)
	case "development":
		gin.SetMode(gin.DebugMode)
	default:
		gin.SetMode(gin.DebugMode)
	}
}
