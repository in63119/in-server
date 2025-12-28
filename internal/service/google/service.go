package google

import (
	"context"
	"fmt"

	"in-server/pkg/config"
)

type Service struct {
	cfg config.Config
}

func New(ctx context.Context, cfg config.Config) (*Service, error) {
	if cfg.Env == "" {
		return nil, fmt.Errorf("config is empty")
	}
	return &Service{cfg: cfg}, nil
}

func (s *Service) BuildGmailOAuthConsentURL(apiBaseURL, state string) (string, error) {
	return s.cfg.BuildGmailOAuthConsentURL(apiBaseURL, state)
}

func (s *Service) ExchangeGoogleAuthCodeToRefreshToken(ctx context.Context, apiBaseURL, code string) (string, error) {
	return s.cfg.ExchangeGoogleAuthCodeToRefreshToken(ctx, apiBaseURL, code)
}

func (s *Service) ValidateGmailRefreshToken(ctx context.Context) (bool, error) {
	return s.cfg.ValidateGmailRefreshToken(ctx)
}

func (s *Service) Env() string {
	return s.cfg.Env
}
