package media

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"in-server/pkg/apperr"
	"in-server/pkg/aws/s3"
	"in-server/pkg/config"
	"in-server/pkg/eth"
)

type Service struct {
	cfg config.Config
	eth *eth.Client
}

func New(ctx context.Context, cfg config.Config) (*Service, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ethClient, err := eth.Dial(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("dial eth client: %w", err)
	}
	return &Service{cfg: cfg, eth: ethClient}, nil
}

func (s *Service) UploadMedia(ctx context.Context, form *multipart.Form) (string, string, error) {
	if form == nil {
		return "", "", apperr.Post.ErrInvalidBody
	}

	files := form.File["file"]
	if len(files) == 0 {
		return "", "", apperr.Post.ErrNoImageFile
	}
	fh := files[0]

	adminCode := ""
	if vals, ok := form.Value["adminCode"]; ok && len(vals) > 0 {
		adminCode = strings.TrimSpace(vals[0])
	}
	if adminCode == "" {
		return "", "", apperr.Post.ErrInvalidAdminCode
	}

	labName := ""
	if vals, ok := form.Value["labName"]; ok && len(vals) > 0 {
		labName = strings.TrimSpace(vals[0])
	}
	slug := ""
	if vals, ok := form.Value["slug"]; ok && len(vals) > 0 {
		slug = strings.TrimSpace(vals[0])
	}

	_, addr, err := s.eth.Wallet(adminCode)
	if err != nil {
		return "", "", apperr.Post.ErrInvalidAdminCode
	}

	file, err := fh.Open()
	if err != nil {
		return "", "", apperr.Post.ErrInvalidUpload
	}
	defer file.Close()

	key := buildMediaKey(addr.Hex(), labName, slug, fh.Filename)
	url, err := s3.PutObject(ctx, s.cfg, key, file, fh.Header.Get("Content-Type"), "")
	if err != nil {
		return "", "", apperr.Wrap(err, apperr.Post.ErrInvalidUpload.Code, "upload media", apperr.Post.ErrInvalidUpload.Status)
	}

	return url, key, nil
}

func buildMediaKey(address, labName, slug, filename string) string {
	toSegment := func(v, fallback string) string {
		v = strings.TrimSpace(v)
		if v == "" {
			v = fallback
		}
		v = strings.ToLower(v)
		var b strings.Builder
		for _, r := range v {
			switch {
			case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
				b.WriteRune(r)
			case r == ' ':
				b.WriteRune('-')
			}
		}
		out := b.String()
		if out == "" {
			return fallback
		}
		return out
	}

	labSegment := toSegment(labName, "lab")
	slugSegment := toSegment(slug, "media")

	ext := "png"
	if parts := strings.Split(strings.TrimSpace(filename), "."); len(parts) > 1 {
		last := strings.ToLower(strings.TrimSpace(parts[len(parts)-1]))
		if last != "" {
			ext = last
		}
	}
	timestamp := time.Now().UTC().Format("2006-01-02T15-04-05")

	return fmt.Sprintf("users/%s/media/%s/%s/%s.%s", strings.ToLower(strings.TrimSpace(address)), labSegment, slugSegment, timestamp, ext)
}
