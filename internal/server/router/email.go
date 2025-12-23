package router

import (
	"github.com/gin-gonic/gin"

	emailhandler "in-server/internal/handler/email"
)

func RegisterEmailRoutes(r gin.IRouter, h *emailhandler.Handler) {
	h.Register(r.Group("/email"))
}
