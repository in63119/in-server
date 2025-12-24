package google

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"in-server/pkg/config"
)

type Handler struct {
	cfg config.Config
}

func New(cfg config.Config) *Handler { return &Handler{cfg: cfg} }

func (h *Handler) Register(r *gin.RouterGroup) {
	r.GET("/authorize", h.redirectToConsent)
	r.GET("/callback", h.handleCallback)
}

func (h *Handler) redirectToConsent(c *gin.Context) {
	baseURL := inferBaseURL(c)
	state := strings.TrimSpace(c.Query("state"))
	url, err := h.cfg.BuildGmailOAuthConsentURL(baseURL, state)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Redirect(http.StatusFound, url)
}

func (h *Handler) handleCallback(c *gin.Context) {
	code := strings.TrimSpace(c.Query("code"))
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	baseURL := inferBaseURL(c)
	refreshToken, err := h.cfg.ExchangeGoogleAuthCodeToRefreshToken(c.Request.Context(), baseURL, code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"refreshToken": refreshToken})
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
