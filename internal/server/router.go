package server

import (
	"context"

	"go.uber.org/zap"

	googhandler "in-server/internal/handler/auth/google"
	emailhandler "in-server/internal/handler/email"
	"in-server/internal/handler/health"
	posthandler "in-server/internal/handler/posts"
	subscriberhandler "in-server/internal/handler/subscriber"
	visitorshandler "in-server/internal/handler/visitors"
	"in-server/internal/server/router"

	emailsvc "in-server/internal/service/email"
	postsvc "in-server/internal/service/post"
	subscribersvc "in-server/internal/service/subscriber"
	visitorsvc "in-server/internal/service/visitor"
)

func (s *Server) registerRoutes() {
	healthHandler := health.New(s.cfg)
	emailSvc, err := emailsvc.New(context.Background(), s.cfg)
	if err != nil {
		s.log.Fatal("init email service", zap.Error(err))
	}
	postSvc, err := postsvc.New(context.Background(), s.cfg)
	if err != nil {
		s.log.Fatal("init post service", zap.Error(err))
	}
	visitorSvc, err := visitorsvc.New(context.Background(), s.cfg)
	if err != nil {
		s.log.Fatal("init visitor service", zap.Error(err))
	}
	subscriberSvc, err := subscribersvc.New(context.Background(), s.cfg)
	if err != nil {
		s.log.Fatal("init subscriber service", zap.Error(err))
	}

	googleHandler := googhandler.New(s.cfg)
	emailHandler := emailhandler.New(emailSvc)
	postHandler := posthandler.New(postSvc)
	visitorsHandler := visitorshandler.New(visitorSvc)
	subscriberHandler := subscriberhandler.New(subscriberSvc)

	r := s.engine
	{
		router.RegisterHealthRoutes(r, healthHandler)
		router.RegisterGoogleAuthRoutes(r, googleHandler)
		router.RegisterEmailRoutes(r, emailHandler)
		router.RegisterPostRoutes(r, postHandler)
		router.RegisterVisitorRoutes(r, visitorsHandler)
		router.RegisterSubscriberRoutes(r, subscriberHandler)
	}
}
