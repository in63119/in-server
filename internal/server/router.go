package server

import (
	"context"

	"go.uber.org/zap"

	emailhandler "in-server/internal/handler/email"
	googhandler "in-server/internal/handler/google"
	"in-server/internal/handler/health"
	mediahandler "in-server/internal/handler/media"
	posthandler "in-server/internal/handler/posts"
	subscriberhandler "in-server/internal/handler/subscriber"
	visitorshandler "in-server/internal/handler/visitors"
	"in-server/internal/server/router"

	emailsvc "in-server/internal/service/email"
	googlesvc "in-server/internal/service/google"
	mediasvc "in-server/internal/service/media"
	postsvc "in-server/internal/service/post"
	subscribersvc "in-server/internal/service/subscriber"
	visitorsvc "in-server/internal/service/visitor"
)

func (s *Server) registerRoutes() {
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
	googleSvc, err := googlesvc.New(context.Background(), s.cfg)
	if err != nil {
		s.log.Fatal("init google service", zap.Error(err))
	}
	mediaSvc, err := mediasvc.New(context.Background(), s.cfg)
	if err != nil {
		s.log.Fatal("init media service", zap.Error(err))
	}

	healthHandler := health.New(s.cfg)
	googleHandler := googhandler.New(googleSvc, s.reloadAll)
	emailHandler := emailhandler.New(emailSvc)
	mediaHandler := mediahandler.New(mediaSvc)
	postHandler := posthandler.New(postSvc)
	visitorsHandler := visitorshandler.New(visitorSvc)
	subscriberHandler := subscriberhandler.New(subscriberSvc)

	s.googleHandler = googleHandler
	s.emailHandler = emailHandler
	s.mediaHandler = mediaHandler
	s.postHandler = postHandler
	s.visitorsHandler = visitorsHandler
	s.subscriberHandler = subscriberHandler

	r := s.engine
	{
		router.RegisterHealthRoutes(r, healthHandler)
		router.RegisterGoogleRoutes(r, googleHandler)
		router.RegisterEmailRoutes(r, emailHandler)
		router.RegisterPostRoutes(r, postHandler)
		router.RegisterVisitorRoutes(r, visitorsHandler)
		router.RegisterSubscriberRoutes(r, subscriberHandler)
		router.RegisterMediaRoutes(r, mediaHandler)
	}
}
