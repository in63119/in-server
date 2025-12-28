package email

import (
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"

	"in-server/internal/handler/httputil"
	emailsvc "in-server/internal/service/email"
	"in-server/pkg/apperr"
)

type Handler struct {
	mu  sync.RWMutex
	svc *emailsvc.Service
}

func New(svc *emailsvc.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) SetService(svc *emailsvc.Service) {
	h.mu.Lock()
	h.svc = svc
	h.mu.Unlock()
}

func (h *Handler) getSvc() *emailsvc.Service {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.svc
}

func (h *Handler) Register(r *gin.RouterGroup) {
	r.POST("/pin", h.claimPinCode)
	r.POST("/pin/verify", h.verifyPinCode)
}

func (h *Handler) claimPinCode(c *gin.Context) {
	svc := h.getSvc()
	if svc == nil {
		httputil.WriteError(c, apperr.Email.ErrInvalidBody)
		return
	}
	var req struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.WriteError(c, apperr.Email.ErrInvalidBody)
		return
	}

	email := strings.TrimSpace(req.Email)
	if email == "" {
		httputil.WriteError(c, apperr.Email.ErrInvalidBody)
		return
	}

	pinCode, err := emailsvc.GenerateFourDigitCode()
	if err != nil {
		httputil.WriteError(c, apperr.Wrap(err, apperr.Email.ErrClaimPinCode.Code, apperr.Email.ErrClaimPinCode.Message, apperr.Email.ErrClaimPinCode.Status))
		return
	}

	if err := svc.ClaimPinCode(c.Request.Context(), pinCode, email); err != nil {
		httputil.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) verifyPinCode(c *gin.Context) {
	svc := h.getSvc()
	if svc == nil {
		httputil.WriteError(c, apperr.Email.ErrInvalidBody)
		return
	}
	var req struct {
		PinCode string `json:"pinCode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		httputil.WriteError(c, apperr.Email.ErrInvalidBody)
		return
	}

	pinCode := strings.TrimSpace(req.PinCode)
	if pinCode == "" {
		httputil.WriteError(c, apperr.Email.ErrInvalidBody)
		return
	}

	verified, err := svc.VerifyPinCode(c.Request.Context(), pinCode)
	if err != nil {
		_ = c.Error(err)
		httputil.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"verified": verified})
}
