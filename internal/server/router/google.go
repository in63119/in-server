package router

import (
	"github.com/gin-gonic/gin"

	googlehandler "in-server/internal/handler/auth/google"
)

func RegisterGoogleAuthRoutes(r gin.IRouter, h *googlehandler.Handler) {
	h.Register(r.Group("/auth/google"))
}
