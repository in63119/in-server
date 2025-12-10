package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"in-server/internal/handler/health"
	posthandler "in-server/internal/handler/posts"
	postsvc "in-server/internal/service/post"
)

var routes = struct {
	Health string
	Ready  string
	Auth   struct {
		Root           string
		Authentication struct {
			Option string
			Verify string
		}
	}
	Posts struct {
		Root    string
		publish string
		delete  string
	}
}{
	Health: "health",
	Ready:  "ready",
	Auth: struct {
		Root           string
		Authentication struct {
			Option string
			Verify string
		}
	}{
		Root: "auth",
		Authentication: struct {
			Option string
			Verify string
		}{
			Option: "authentication/option",
			Verify: "authentication/verify",
		},
	},
}

func (s *Server) registerRoutes() {
	healthHandler := health.New(s.cfg)
	postHandler := posthandler.New(postsvc.New())

	r := s.engine
	{
		r.GET("/"+routes.Health, healthHandler.Health)
		r.GET("/"+routes.Ready, healthHandler.Ready)

		postHandler.Register(r.Group("/"))

		auth := r.Group("/" + routes.Auth.Root)
		{
			auth.Any("/"+routes.Auth.Authentication.Option, notImplemented("auth option"))
			auth.Any("/"+routes.Auth.Authentication.Verify, notImplemented("auth verify"))
		}
	}
}

func notImplemented(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{
			"endpoint": name,
			"message":  "not implemented yet",
		})
	}
}
