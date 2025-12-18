package post

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	gethtypes "github.com/ethereum/go-ethereum/core/types"

	"in-server/pkg/apperr"
	awss3 "in-server/pkg/aws/s3"
	pkgtypes "in-server/pkg/types"
)

func (s *Service) Publish(ctx context.Context, adminCode string, payload NftMetadata, metadataURL string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if s.eth == nil {
		return "", fmt.Errorf("eth client is nil")
	}

	adminCode = strings.TrimSpace(adminCode)
	if adminCode == "" {
		return "", apperr.Post.ErrAdminCodeMissing
	}

	if strings.TrimSpace(payload.Name) == "" ||
		strings.TrimSpace(payload.Description) == "" ||
		strings.TrimSpace(payload.ExternalURL) == "" ||
		len(payload.Attributes) == 0 {
		return "", apperr.Post.ErrInvalidBody
	}

	pk, ownerAddr, err := s.eth.Wallet(adminCode)
	if err != nil {
		return "", apperr.Wrap(err, apperr.Post.ErrInvalidAdminCode.Code, apperr.Post.ErrInvalidAdminCode.Message, apperr.Post.ErrInvalidAdminCode.Status)
	}

	labSegment := toPathSegment(attrValue(payload.Attributes, "Lab"))
	if labSegment == "" {
		labSegment = "lab"
	}
	slugSegment := toPathSegment(attrValue(payload.Attributes, "Slug"))
	if slugSegment == "" {
		slugSegment = "post"
	}

	normalized := normalizeMetadata(payload)

	metadataURL = strings.TrimSpace(metadataURL)
	existingKey := ""
	if metadataURL != "" {
		existingKey = extractKeyFromMetadataURL(metadataURL)
	}

	posts, err := s.listByOwner(ctx, ownerAddr)
	if err != nil {
		return "", err
	}
	for _, p := range posts {
		if p.LabSegment != labSegment || p.Slug != slugSegment {
			continue
		}
		if metadataURL != "" {
			if existingKey != "" {
				postKey := extractKeyFromMetadataURL(p.MetadataURL)
				if postKey != "" && postKey == existingKey {
					continue
				}
			} else if p.MetadataURL == metadataURL {
				continue
			}
		}
		return "", apperr.Post.ErrDuplicatePost
	}

	resolvedURL, isUpdate, err := s.saveMetadata(ctx, ownerAddr.Hex(), labSegment, normalized, metadataURL, existingKey)
	if err != nil {
		return "", err
	}

	if !isUpdate {
		if s.fb == nil {
			return "", fmt.Errorf("firebase client is nil")
		}
		receipt, err := s.eth.Excute(ctx, s.fb, pkgtypes.POSTSTORAGE, pk, "post", ownerAddr, resolvedURL)
		if err != nil {
			return "", apperr.Wrap(err, apperr.Post.ErrPublishFailed.Code, apperr.Post.ErrPublishFailed.Message, apperr.Post.ErrPublishFailed.Status)
		}
		if receipt == nil || receipt.Status != gethtypes.ReceiptStatusSuccessful {
			return "", apperr.Wrap(fmt.Errorf("meta tx status %v", receiptStatus(receipt)), apperr.Post.ErrPublishFailed.Code, apperr.Post.ErrPublishFailed.Message, apperr.Post.ErrPublishFailed.Status)
		}
	}

	return resolvedURL, nil
}

func normalizeMetadata(payload NftMetadata) NftMetadata {
	out := payload
	if len(payload.Attributes) == 0 {
		return out
	}

	outAttrs := make([]NftAttribute, 0, len(payload.Attributes))
	for _, attr := range payload.Attributes {
		if strings.TrimSpace(attr.TraitType) != "RelatedLinks" {
			outAttrs = append(outAttrs, attr)
			continue
		}
		var raw string
		if err := json.Unmarshal(attr.Value, &raw); err != nil {
			continue
		}
		cleaned := strings.Join(strings.Fields(raw), " ")
		attr.Value = json.RawMessage(mustJSONQuote(cleaned))
		outAttrs = append(outAttrs, attr)
	}
	out.Attributes = outAttrs
	return out
}

func (s *Service) saveMetadata(ctx context.Context, addressHex, labSegment string, payload NftMetadata, metadataURL, existingKey string) (resolvedURL string, isUpdate bool, _ error) {
	addressPrefix := fmt.Sprintf("users/%s/", addressHex)

	var key string
	if strings.TrimSpace(metadataURL) != "" {
		if existingKey == "" || !strings.HasPrefix(existingKey, addressPrefix) {
			return "", false, apperr.Post.ErrInvalidRequest
		}
		key = existingKey
		isUpdate = true
	} else {
		env := strings.TrimSpace(s.cfg.Env)
		if env == "" {
			env = "development"
		}
		key = fmt.Sprintf("%sposts/%s/%s/metadata-%s.json", addressPrefix, env, labSegment, timestampForKey())
		isUpdate = false
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", false, fmt.Errorf("marshal metadata: %w", err)
	}

	uploadedURL, err := awss3.PutObject(ctx, s.cfg, key, bytes.NewReader(data), "application/json", "")
	if err != nil {
		return "", false, apperr.Wrap(err, apperr.Post.ErrUploadMetadata.Code, apperr.Post.ErrUploadMetadata.Message, apperr.Post.ErrUploadMetadata.Status)
	}

	return uploadedURL, isUpdate, nil
}

func timestampForKey() string {
	ts := time.Now().UTC().Format(time.RFC3339Nano)
	ts = strings.ReplaceAll(ts, ":", "-")
	ts = strings.ReplaceAll(ts, ".", "-")
	ts = strings.TrimSuffix(ts, "Z")
	return ts
}

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

func toPathSegment(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return ""
	}
	s = nonAlnum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

func extractKeyFromMetadataURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err == nil && u.Path != "" {
		return strings.TrimPrefix(u.Path, "/")
	}
	return strings.TrimPrefix(raw, "/")
}

func attrValue(attrs []NftAttribute, trait string) string {
	for _, a := range attrs {
		if strings.TrimSpace(a.TraitType) != trait {
			continue
		}
		var s string
		if err := json.Unmarshal(a.Value, &s); err == nil {
			return strings.TrimSpace(s)
		}
		return strings.TrimSpace(string(a.Value))
	}
	return ""
}

func receiptStatus(r *gethtypes.Receipt) any {
	if r == nil {
		return nil
	}
	return r.Status
}

func mustJSONQuote(s string) []byte {
	b, err := json.Marshal(s)
	if err != nil {
		return []byte(`""`)
	}
	return b
}
