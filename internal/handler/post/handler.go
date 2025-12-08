package post

import (
	"in-server/internal/service/post"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct{ svc *post.Service }

func New(svc *post.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(r *gin.RouterGroup) {
	r.GET("/posts", h.list)
	r.POST("/posts", h.create)
}

func (h *Handler) list(c *gin.Context) {
	items, err := h.svc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler) create(c *gin.Context) {
	var req post.Post
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if err := h.svc.Create(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create"})
		return
	}
	c.Status(http.StatusCreated)
}
