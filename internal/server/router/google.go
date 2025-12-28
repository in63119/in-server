package router

import (
	"github.com/gin-gonic/gin"

	googlehandler "in-server/internal/handler/google"
)

func RegisterGoogleRoutes(r gin.IRouter, h *googlehandler.Handler) {
	h.Register(r.Group("/google"))
}
