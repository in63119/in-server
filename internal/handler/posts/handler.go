package post

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"in-server/internal/handler/httputil"
	"in-server/internal/service/post"
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
		httputil.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *Handler) create(c *gin.Context) {
	var req post.Post
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if err := h.svc.Create(req); err != nil {
		httputil.WriteError(c, err)
		return
	}
	c.Status(http.StatusCreated)
}
