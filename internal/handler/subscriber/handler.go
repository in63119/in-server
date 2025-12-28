package subscriber

import (
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"

	"in-server/internal/handler/httputil"
	"in-server/internal/service/subscriber"
	"in-server/pkg/apperr"
)

type Handler struct {
	mu  sync.RWMutex
	svc *subscriber.Service
}

func New(svc *subscriber.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) SetService(svc *subscriber.Service) {
	h.mu.Lock()
	h.svc = svc
	h.mu.Unlock()
}

func (h *Handler) getSvc() *subscriber.Service {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.svc
}

func (h *Handler) Register(r *gin.RouterGroup) {
	r.GET("", h.count)
	r.POST("", h.create)
	r.POST("/list", h.list)
}

func (h *Handler) count(c *gin.Context) {
	svc := h.getSvc()
	if svc == nil {
		httputil.WriteError(c, apperr.Subscriber.ErrGetSubscribers)
		return
	}
	total, err := svc.Count()
	if err != nil {
		httputil.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": total})
}

func (h *Handler) create(c *gin.Context) {
	svc := h.getSvc()
	if svc == nil {
		httputil.WriteError(c, apperr.Subscriber.ErrInvalidBody)
		return
	}
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

	if err := svc.Create("", req.Email); err != nil {
		httputil.WriteError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"subscribed": true})
}

func (h *Handler) list(c *gin.Context) {
	svc := h.getSvc()
	if svc == nil {
		httputil.WriteError(c, apperr.Subscriber.ErrGetSubscribers)
		return
	}
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

	subscribers, err := svc.List(0, 0)
	if err != nil {
		httputil.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"subscribers": subscribers})
}
