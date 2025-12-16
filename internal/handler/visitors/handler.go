package visitors

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"in-server/internal/service/visitor"
)

type Handler struct {
	svc *visitor.Service
}

func New(svc *visitor.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(r *gin.RouterGroup) {
	r.GET("/visitors", h.list)
	r.POST("/visitors", h.create)
	r.GET("/visitors/check", h.check)
}

func (h *Handler) list(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "list visitors not implemented",
	})
}

func (h *Handler) create(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "create visitor not implemented",
	})
}

func (h *Handler) check(c *gin.Context) {
	ip := clientIP(c)
	if ip == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_IP", "message": "invalid ip"})
		return
	}

	visited, err := h.svc.HasVisited(ip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAILED_TO_CHECK_VISIT", "message": "failed to check visit"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"visited": visited})
}

func clientIP(c *gin.Context) string {
	if v := c.GetHeader("X-Forwarded-For"); v != "" {
		parts := strings.Split(v, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if v := c.GetHeader("X-Real-Ip"); v != "" {
		return strings.TrimSpace(v)
	}
	return strings.TrimSpace(c.ClientIP())
}
