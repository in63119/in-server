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
	r.POST("/list", h.list)
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

func (h *Handler) list(c *gin.Context) {
	var req struct {
		AdminCode string `json:"adminCode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("subscriber list bind error:", err)
		httputil.WriteError(c, apperr.Subscriber.ErrInvalidBody)
		return
	}

	req.AdminCode = strings.TrimSpace(req.AdminCode)
	if req.AdminCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "관리자 코드가 필요합니다."})
		return
	}

	subscribers, err := h.svc.List(0, 0)
	if err != nil {
		httputil.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"subscribers": subscribers})
}
