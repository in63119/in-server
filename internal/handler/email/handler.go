package email

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"in-server/internal/handler/httputil"
	emailsvc "in-server/internal/service/email"
	"in-server/pkg/apperr"
)

type Handler struct{ svc *emailsvc.Service }

func New(svc *emailsvc.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(r *gin.RouterGroup) {
	r.POST("/pin", h.claimPinCode)
	r.POST("/pin/verify", h.verifyPinCode)
}

func (h *Handler) claimPinCode(c *gin.Context) {
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

	if err := h.svc.ClaimPinCode(c.Request.Context(), pinCode, email); err != nil {
		log.Println(err)
		httputil.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) verifyPinCode(c *gin.Context) {
	var req struct {
		Address string `json:"address"`
		PinCode string `json:"pinCode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		httputil.WriteError(c, apperr.Email.ErrInvalidBody)
		return
	}

	address := strings.TrimSpace(req.Address)
	pinCode := strings.TrimSpace(req.PinCode)
	if address == "" || pinCode == "" {
		httputil.WriteError(c, apperr.Email.ErrInvalidBody)
		return
	}

	verified, err := h.svc.VerifyPinCode(c.Request.Context(), address, pinCode)
	if err != nil {
		_ = c.Error(err)
		httputil.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"verified": verified})
}
