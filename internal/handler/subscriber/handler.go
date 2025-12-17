package subscriber

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"in-server/internal/handler/httputil"
	"in-server/internal/service/subscriber"
	"in-server/pkg/apperr"
)

type Handler struct{ svc *subscriber.Service }

func New(svc *subscriber.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(r *gin.RouterGroup) {
	r.GET("/subscriber", h.count)
	r.POST("/subscriber", h.create)
}

func (h *Handler) count(c *gin.Context) {
	total, err := h.svc.Count()
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": total})
}

func (h *Handler) create(c *gin.Context) {
	httputil.WriteError(c, apperr.System.ErrNotImplemented)
}
