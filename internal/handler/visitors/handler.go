package visitors

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"in-server/internal/handler/httputil"
	"in-server/internal/service/visitor"
	"in-server/pkg/apperr"
)

type Handler struct {
	svc *visitor.Service
}

func New(svc *visitor.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(r *gin.RouterGroup) {
	r.GET("/visitors", h.count)
	r.POST("/visitors", h.visit)
	r.GET("/visitors/check", h.check)
}

func (h *Handler) count(c *gin.Context) {
	total, err := h.svc.Count()
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": total})
}

func (h *Handler) visit(c *gin.Context) {
	var req struct {
		URL string `json:"url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.URL) == "" {
		httputil.WriteError(c, apperr.Visitors.ErrInvalidBody)
		return
	}

	ip := clientIP(c)
	if ip == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_IP", "message": "invalid ip"})
		return
	}

	if err := h.svc.Visit(ip, req.URL); err != nil {
		httputil.WriteError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *Handler) check(c *gin.Context) {
	ip := clientIP(c)
	if ip == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_IP", "message": "invalid ip"})
		return
	}

	visited, err := h.svc.HasVisited(ip)
	if err != nil {
		httputil.WriteError(c, err)
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
