package server

import (
	"context"

	"go.uber.org/zap"

	"in-server/internal/handler/health"
	posthandler "in-server/internal/handler/posts"
	subscriberhandler "in-server/internal/handler/subscriber"
	visitorshandler "in-server/internal/handler/visitors"

	postsvc "in-server/internal/service/post"
	subscribersvc "in-server/internal/service/subscriber"
	visitorsvc "in-server/internal/service/visitor"
)

var routes = struct {
	Health string
	Ready  string
	// Auth   struct {
	// 	Root           string
	// 	Authentication struct {
	// 		Option string
	// 		Verify string
	// 	}
	// }
	Posts struct {
		Root    string
		publish string
		delete  string
	}
	Visitors struct {
		Root  string
		check string
	}
	Subscriber struct {
		Root string
	}
}{
	Health: "health",
	Ready:  "ready",
	Posts: struct {
		Root    string
		publish string
		delete  string
	}{
		Root:    "posts",
		publish: "publish",
		delete:  "delete",
	},
	Visitors: struct {
		Root  string
		check string
	}{
		Root:  "visitors",
		check: "check",
	},
	Subscriber: struct {
		Root string
	}{
		Root: "subscriber",
	},
	// Auth: struct {
	// 	Root           string
	// 	Authentication struct {
	// 		Option string
	// 		Verify string
	// 	}
	// }{
	// 	Root: "auth",
	// 	Authentication: struct {
	// 		Option string
	// 		Verify string
	// 	}{
	// 		Option: "authentication/option",
	// 		Verify: "authentication/verify",
	// 	},
	// },
}

func (s *Server) registerRoutes() {
	healthHandler := health.New(s.cfg)
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
	postHandler := posthandler.New(postSvc)
	visitorsHandler := visitorshandler.New(visitorSvc)
	subscriberHandler := subscriberhandler.New(subscriberSvc)

	r := s.engine
	{
		r.GET("/"+routes.Health, healthHandler.Health)
		r.GET("/"+routes.Ready, healthHandler.Ready)

		postHandler.Register(r.Group("/"))
		visitorsHandler.Register(r.Group("/"))
		subscriberHandler.Register(r.Group("/"))

		// auth := r.Group("/" + routes.Auth.Root)
		// {
		// 	auth.Any("/"+routes.Auth.Authentication.Option, notImplemented("auth option"))
		// 	auth.Any("/"+routes.Auth.Authentication.Verify, notImplemented("auth verify"))
		// }
	}
}

// func notImplemented(name string) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		c.JSON(http.StatusNotImplemented, gin.H{
// 			"endpoint": name,
// 			"message":  "not implemented yet",
// 		})
// 	}
// }
