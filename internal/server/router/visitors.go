package router

import (
	"github.com/gin-gonic/gin"

	visitorshandler "in-server/internal/handler/visitors"
)

func RegisterVisitorRoutes(r gin.IRouter, h *visitorshandler.Handler) {
	h.Register(r.Group("/visitors"))
}
