package router

import (
	"github.com/gin-gonic/gin"

	mediahandler "in-server/internal/handler/media"
)

func RegisterMediaRoutes(r gin.IRouter, h *mediahandler.Handler) {
	h.Register(r.Group("/media"))
}
