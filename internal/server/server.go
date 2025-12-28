package server

import (
	"context"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	emailhandler "in-server/internal/handler/email"
	googhandler "in-server/internal/handler/google"
	posthandler "in-server/internal/handler/posts"
	subscriberhandler "in-server/internal/handler/subscriber"
	visitorshandler "in-server/internal/handler/visitors"
	emailsvc "in-server/internal/service/email"
	googlesvc "in-server/internal/service/google"
	postsvc "in-server/internal/service/post"
	subscribersvc "in-server/internal/service/subscriber"
	visitorsvc "in-server/internal/service/visitor"
	"in-server/pkg/config"
)

type Server struct {
	cfg    config.Config
	engine *gin.Engine
	log    *zap.Logger

	emailHandler      *emailhandler.Handler
	postHandler       *posthandler.Handler
	visitorsHandler   *visitorshandler.Handler
	subscriberHandler *subscriberhandler.Handler
	googleHandler     *googhandler.Handler
	mu                sync.RWMutex
}

func New(cfg config.Config, log *zap.Logger) *Server {
	setGinMode(cfg.Env)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(gin.Logger())
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-Requested-With", "Accept", "Origin"},
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

func (s *Server) reloadAll(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	newCfg, err := config.Reload(ctx)
	if err != nil {
		return err
	}

	emailSvc, err := emailsvc.New(ctx, newCfg)
	if err != nil {
		return err
	}
	postSvc, err := postsvc.New(ctx, newCfg)
	if err != nil {
		return err
	}
	visitorSvc, err := visitorsvc.New(ctx, newCfg)
	if err != nil {
		return err
	}
	subscriberSvc, err := subscribersvc.New(ctx, newCfg)
	if err != nil {
		return err
	}
	googleSvc, err := googlesvc.New(ctx, newCfg)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.cfg = newCfg
	if s.emailHandler != nil {
		s.emailHandler.SetService(emailSvc)
	}
	if s.postHandler != nil {
		s.postHandler.SetService(postSvc)
	}
	if s.visitorsHandler != nil {
		s.visitorsHandler.SetService(visitorSvc)
	}
	if s.subscriberHandler != nil {
		s.subscriberHandler.SetService(subscriberSvc)
	}
	if s.googleHandler != nil {
		s.googleHandler.SetService(googleSvc)
	}
	s.mu.Unlock()

	return nil
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
