package server

import (
	"in-server/internal/config"
	"in-server/internal/handler/health"

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

func (s *Server) registerRoutes() {
	healthHandler := health.New(s.cfg)

	api := s.engine.Group("/api")
	{
		api.GET("/health", healthHandler.Health)
		api.GET("/ready", healthHandler.Ready)
	}
}

func setGinMode(env string) {
	if env == "production" || env == "prod" {
		gin.SetMode(gin.ReleaseMode)
		return
	}
	if env == "test" {
		gin.SetMode(gin.TestMode)
		return
	}
	gin.SetMode(gin.DebugMode)
}
