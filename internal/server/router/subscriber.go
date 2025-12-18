package router

import (
	"github.com/gin-gonic/gin"

	subscriberhandler "in-server/internal/handler/subscriber"
)

func RegisterSubscriberRoutes(r gin.IRouter, h *subscriberhandler.Handler) {
	h.Register(r.Group("/subscriber"))
}
