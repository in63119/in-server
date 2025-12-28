package media

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"in-server/internal/handler/httputil"
	mediasvc "in-server/internal/service/media"
	"in-server/pkg/apperr"
)

type Handler struct {
	mu  sync.RWMutex
	svc *mediasvc.Service
}

func New(svc *mediasvc.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) SetService(svc *mediasvc.Service) {
	h.mu.Lock()
	h.svc = svc
	h.mu.Unlock()
}

func (h *Handler) getSvc() *mediasvc.Service {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.svc
}

func (h *Handler) Register(r *gin.RouterGroup) {
	r.POST("/upload", h.upload)
}

func (h *Handler) upload(c *gin.Context) {
	svc := h.getSvc()
	if svc == nil {
		httputil.WriteError(c, apperr.Post.ErrInvalidRequest)
		return
	}

	if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32MB
		httputil.WriteError(c, apperr.Post.ErrInvalidBody)
		return
	}

	url, key, err := svc.UploadMedia(c.Request.Context(), c.Request.MultipartForm)
	if err != nil {
		httputil.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":  true,
		"url": url,
		"key": key,
	})
}
