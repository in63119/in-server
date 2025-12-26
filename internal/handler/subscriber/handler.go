package subscriber

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"in-server/internal/handler/httputil"
	"in-server/internal/service/subscriber"
	"in-server/pkg/apperr"
)

type Handler struct{ svc *subscriber.Service }

func New(svc *subscriber.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(r *gin.RouterGroup) {
	r.GET("", h.count)
	r.POST("", h.create)
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
	var req struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("subscriber create bind error:", err)
		httputil.WriteError(c, apperr.Subscriber.ErrInvalidBody)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		httputil.WriteError(c, apperr.Subscriber.ErrInvalidBody)
		return
	}

	if err := h.svc.Create("", req.Email); err != nil {
		httputil.WriteError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"subscribed": true})
}
