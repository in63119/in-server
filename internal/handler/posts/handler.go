package post

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"in-server/internal/handler/httputil"
	"in-server/internal/service/post"
	"in-server/pkg/apperr"
)

type Handler struct{ svc *post.Service }

func New(svc *post.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(r *gin.RouterGroup) {
	r.GET("", h.list)
	r.POST("", h.create)
	r.POST("/publish", h.publish)
}

func (h *Handler) list(c *gin.Context) {
	items, err := h.svc.List()
	if err != nil {
		httputil.WriteError(c, err)
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
		httputil.WriteError(c, err)
		return
	}
	c.Status(http.StatusCreated)
}

func (h *Handler) publish(c *gin.Context) {
	var req struct {
		AdminCode   string          `json:"adminCode"`
		MetadataURL *string         `json:"metadataUrl"`
		Payload     json.RawMessage `json:"payload"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println(err)
		httputil.WriteError(c, apperr.Post.ErrInvalidBody)
		return
	}

	adminCode := strings.TrimSpace(req.AdminCode)
	metadataURL := ""
	if req.MetadataURL != nil {
		metadataURL = strings.TrimSpace(*req.MetadataURL)
	}

	payload, err := parseNFTMetadata(req.Payload)
	if err != nil {
		httputil.WriteError(c, err)
		return
	}

	savedMetadataURL, err := h.svc.Publish(c.Request.Context(), adminCode, payload, metadataURL)
	if err != nil {
		httputil.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "metadataUrl": savedMetadataURL})
}

func parseNFTMetadata(raw json.RawMessage) (post.NftMetadata, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return post.NftMetadata{}, apperr.Post.ErrInvalidBody
	}

	if raw[0] == '"' {
		var inner string
		if err := json.Unmarshal(raw, &inner); err != nil {
			return post.NftMetadata{}, apperr.Post.ErrInvalidBody
		}
		raw = bytes.TrimSpace([]byte(inner))
	}

	var meta post.NftMetadata
	if err := json.Unmarshal(raw, &meta); err != nil {
		return post.NftMetadata{}, apperr.Post.ErrInvalidBody
	}

	if strings.TrimSpace(meta.Name) == "" ||
		strings.TrimSpace(meta.Description) == "" ||
		strings.TrimSpace(meta.ExternalURL) == "" ||
		len(meta.Attributes) == 0 {
		return post.NftMetadata{}, apperr.Post.ErrInvalidBody
	}
	return meta, nil
}
