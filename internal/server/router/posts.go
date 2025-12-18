package router

import (
	"github.com/gin-gonic/gin"

	posthandler "in-server/internal/handler/posts"
)

func RegisterPostRoutes(r gin.IRouter, h *posthandler.Handler) {
	h.Register(r.Group("/posts"))
}
