package config

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type GmailCredentials struct {
	ClientID     string
	ClientSecret string
	RefreshToken string
	Sender       string
	RedirectURL  string
}

type GoogleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
	IDToken      string `json:"id_token,omitempty"`
}

const gmailSendScope = "https://www.googleapis.com/auth/gmail.send"

func (cfg Config) GmailCredentials(apiBaseURL string) (GmailCredentials, error) {
	clientID := strings.TrimSpace(cfg.Google.ClientKey)
	clientSecret := strings.TrimSpace(cfg.Google.SecretKey)
	refreshToken := strings.TrimSpace(cfg.Google.RefreshToken)
	sender := strings.TrimSpace(cfg.Google.GmailSender)

	redirectURL, err := joinURL(apiBaseURL, cfg.Google.RedirectURIEndpoint)
	if err != nil {
		return GmailCredentials{}, err
	}

	if clientID == "" || clientSecret == "" || sender == "" || redirectURL == "" {
		return GmailCredentials{}, fmt.Errorf("google gmail config incomplete")
	}

	return GmailCredentials{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RefreshToken: refreshToken,
		Sender:       sender,
		RedirectURL:  redirectURL,
	}, nil
}

func (cfg Config) BuildGmailOAuthConsentURL(apiBaseURL string, state string) (string, error) {
	creds, err := cfg.GmailCredentials(apiBaseURL)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(state) == "" {
		state = "state"
	}

	oauthCfg := oauth2.Config{
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		RedirectURL:  creds.RedirectURL,
		Scopes:       []string{gmailSendScope},
		Endpoint:     google.Endpoint,
	}

	return oauthCfg.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
		oauth2.SetAuthURLParam("include_granted_scopes", "true"),
	), nil
}

func (cfg Config) ExchangeGoogleAuthCodeToRefreshToken(ctx context.Context, apiBaseURL string, code string) (string, error) {
	creds, err := cfg.GmailCredentials(apiBaseURL)
	if err != nil {
		return "", err
	}

	code = strings.TrimSpace(code)
	if code == "" {
		return "", fmt.Errorf("google auth code is empty")
	}

	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", creds.ClientID)
	form.Set("client_secret", creds.ClientSecret)
	form.Set("redirect_uri", creds.RedirectURL)
	form.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("build google token exchange request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("google token exchange: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("google token exchange failed: status %d %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var tokens GoogleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return "", fmt.Errorf("decode google token response: %w", err)
	}

	if strings.TrimSpace(tokens.RefreshToken) == "" {
		return "", fmt.Errorf("missing refresh_token (check offline access + consent prompt)")
	}

	return tokens.RefreshToken, nil
}

func (cfg Config) NewGmailClient(ctx context.Context) (*gmail.Service, string, error) {
	clientID := strings.TrimSpace(cfg.Google.ClientKey)
	clientSecret := strings.TrimSpace(cfg.Google.SecretKey)
	refreshToken := strings.TrimSpace(cfg.Google.RefreshToken)
	sender := strings.TrimSpace(cfg.Google.GmailSender)

	if clientID == "" || clientSecret == "" || sender == "" {
		return nil, "", fmt.Errorf("google gmail config incomplete")
	}
	if refreshToken == "" {
		return nil, "", fmt.Errorf("google refresh token missing")
	}

	oauthCfg := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{gmailSendScope},
	}

	ts := oauthCfg.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken})
	svc, err := gmail.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, "", fmt.Errorf("init gmail client: %w", err)
	}

	return svc, sender, nil
}

func (cfg Config) ValidateGmailRefreshToken(ctx context.Context) (bool, error) {
	clientID := strings.TrimSpace(cfg.Google.ClientKey)
	clientSecret := strings.TrimSpace(cfg.Google.SecretKey)
	refreshToken := strings.TrimSpace(cfg.Google.RefreshToken)

	if clientID == "" || clientSecret == "" || refreshToken == "" {
		return false, fmt.Errorf("google gmail config incomplete")
	}

	oauthCfg := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{gmailSendScope},
	}

	ts := oauthCfg.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken})
	_, err := ts.Token()
	if err == nil {
		return true, nil
	}

	var rErr *oauth2.RetrieveError
	if errors.As(err, &rErr) {
		if rErr.Response != nil && rErr.Response.StatusCode == http.StatusBadRequest {
			if bytes.Contains(bytes.ToLower(rErr.Body), []byte("invalid_grant")) {
				return false, nil
			}
		}
	}

	if strings.Contains(strings.ToLower(err.Error()), "invalid_grant") {
		return false, nil
	}

	return false, err
}

func joinURL(baseURL, endpoint string) (string, error) {
	baseURL = strings.TrimSpace(baseURL)
	endpoint = strings.TrimSpace(endpoint)
	if baseURL == "" {
		return "", fmt.Errorf("api base url is empty")
	}
	if endpoint == "" {
		return "", fmt.Errorf("redirect uri endpoint is empty")
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse api base url: %w", err)
	}
	ref, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("parse redirect uri endpoint: %w", err)
	}
	return base.ResolveReference(ref).String(), nil
}
