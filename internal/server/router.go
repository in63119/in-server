package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"in-server/internal/handler/health"
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

	r := s.engine
	{
		r.GET("/"+routes.Health, healthHandler.Health)
		r.GET("/"+routes.Ready, healthHandler.Ready)

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
