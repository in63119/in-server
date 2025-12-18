package router

import (
	"github.com/gin-gonic/gin"

	"in-server/internal/handler/health"
)

func RegisterHealthRoutes(r gin.IRouter, h *health.Handler) {
	r.GET("/health", h.Health)
	r.GET("/ready", h.Ready)
}
