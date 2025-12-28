package google

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"

	googlesvc "in-server/internal/service/google"
	"in-server/pkg/config"
)

type Handler struct {
	mu  sync.RWMutex
	svc *googlesvc.Service
}

func New(svc *googlesvc.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) getSvc() *googlesvc.Service {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.svc
}

func (h *Handler) setSvc(svc *googlesvc.Service) {
	h.mu.Lock()
	h.svc = svc
	h.mu.Unlock()
}

func (h *Handler) Register(r *gin.RouterGroup) {
	r.GET("/authorize", h.redirectToConsent)
	r.GET("/callback", h.handleCallback)
	r.GET("/status", h.status)
}

func (h *Handler) redirectToConsent(c *gin.Context) {
	baseURL := inferBaseURL(c)
	state := strings.TrimSpace(c.Query("state"))
	svc := h.getSvc()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "google service unavailable"})
		return
	}
	url, err := svc.BuildGmailOAuthConsentURL(baseURL, state)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Redirect(http.StatusFound, url)
}

func (h *Handler) handleCallback(c *gin.Context) {
	code := strings.TrimSpace(c.Query("code"))
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "message": "code 파라미터가 없습니다."})
		return
	}

	svc := h.getSvc()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": "google service unavailable"})
		return
	}

	baseURL := inferBaseURL(c)
	refreshToken, err := svc.ExchangeGoogleAuthCodeToRefreshToken(c.Request.Context(), baseURL, code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	env := svc.Env()
	accessKey := config.GetEnv("AWS_SSM_ACCESS_KEY")
	secretAccessKey := config.GetEnv("AWS_SSM_SECRET_KEY")
	region := config.GetEnv("AWS_REGION")
	ssmServer := config.GetEnv("AWS_SSM_SERVER")

	if accessKey == "" || secretAccessKey == "" || region == "" || ssmServer == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": "AWS SSM 환경변수가 설정되지 않았습니다."})
		return
	}

	param := fmt.Sprintf("%s/%s", strings.TrimSuffix(ssmServer, "/"), env)
	if err := config.SaveSSM(c.Request.Context(), config.SaveSSMInput{
		AccessKey:       accessKey,
		SecretAccessKey: secretAccessKey,
		Region:          region,
		Param:           param,
		Patch: map[string]any{
			"GOOGLE.REFRESH_TOKEN": refreshToken,
		},
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
		return
	}

	newCfg, err := config.Reload(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
		return
	}

	// Reinitialize Google service with refreshed config so subsequent requests use latest token.
	newSvc, err := googlesvc.New(c.Request.Context(), newCfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
		return
	}
	h.setSvc(newSvc)

	c.JSON(http.StatusOK, gin.H{"ok": true, "refreshToken": refreshToken})
}

func inferBaseURL(c *gin.Context) string {
	scheme := strings.TrimSpace(c.GetHeader("X-Forwarded-Proto"))
	if scheme == "" {
		if c.Request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	host := strings.TrimSpace(c.GetHeader("X-Forwarded-Host"))
	if host == "" {
		host = strings.TrimSpace(c.Request.Host)
	}
	return fmt.Sprintf("%s://%s", scheme, host)
}

func (h *Handler) status(c *gin.Context) {
	svc := h.getSvc()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "google service unavailable"})
		return
	}
	valid, err := svc.ValidateGmailRefreshToken(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "valid": valid})
}
