package config

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (cfg Config) FirebaseCredentialsJSON() ([]byte, error) {
	privateKey := strings.ReplaceAll(cfg.Firebase.PrivateKey, `\n`, "\n")

	if strings.TrimSpace(cfg.Firebase.ProjectID) == "" ||
		strings.TrimSpace(cfg.Firebase.ClientEmail) == "" ||
		strings.TrimSpace(privateKey) == "" {
		return nil, fmt.Errorf("firebase config incomplete")
	}

	payload := map[string]string{
		"type":                        "service_account",
		"project_id":                  cfg.Firebase.ProjectID,
		"client_email":                cfg.Firebase.ClientEmail,
		"private_key":                 privateKey,
		"token_uri":                   "https://oauth2.googleapis.com/token",
		"auth_uri":                    "https://accounts.google.com/o/oauth2/auth",
		"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
	}

	return json.Marshal(payload)
}
