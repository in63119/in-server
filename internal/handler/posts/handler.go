package post

import (
	"errors"
	"in-server/internal/service/post"
	"net/http"

	"github.com/gin-gonic/gin"
	"in-server/pkg/apperr"
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
		writeError(c, err)
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
		writeError(c, err)
		return
	}
	c.Status(http.StatusCreated)
}

func writeError(c *gin.Context, err error) {
	var appErr *apperr.Error
	if errors.As(err, &appErr) {
		c.JSON(appErr.Status, gin.H{"code": appErr.Code, "message": appErr.Message})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_SERVER_ERROR", "message": "internal server error"})
}
