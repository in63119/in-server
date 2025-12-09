package firebase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	firebaseAdmin "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/db"
	"google.golang.org/api/option"

	"in-server/pkg/config"
)

type Client struct {
	app *firebaseAdmin.App
	db  *db.Client
}

func New(ctx context.Context, cfg config.Config) (*Client, error) {
	credsJSON, err := cfg.FirebaseCredentialsJSON()
	if err != nil {
		return nil, fmt.Errorf("firebase credentials: %w", err)
	}

	app, err := firebaseAdmin.NewApp(ctx, &firebaseAdmin.Config{
		DatabaseURL: cfg.Firebase.DatabaseURL,
	}, option.WithCredentialsJSON(credsJSON))
	if err != nil {
		return nil, fmt.Errorf("init firebase app: %w", err)
	}

	dbClient, err := app.Database(ctx)
	if err != nil {
		return nil, fmt.Errorf("init firebase db: %w", err)
	}

	return &Client{app: app, db: dbClient}, nil
}

func normalizePath(path string) (string, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "", fmt.Errorf("firebase path must be a non-empty string")
	}

	if trimmed == "/" {
		return "", nil
	}

	return strings.TrimLeft(trimmed, "/"), nil
}

func Read[T any](ctx context.Context, c *Client, path string) (T, bool, error) {
	var zero T
	if c == nil || c.db == nil {
		return zero, false, fmt.Errorf("firebase client is nil")
	}

	norm, err := normalizePath(path)
	if err != nil {
		return zero, false, err
	}

	ref := c.db.NewRef(norm)

	var raw any
	if err := ref.Get(ctx, &raw); err != nil {
		return zero, false, fmt.Errorf("get %s: %w", norm, err)
	}
	if raw == nil {
		return zero, false, nil
	}

	buf, err := json.Marshal(raw)
	if err != nil {
		return zero, false, fmt.Errorf("marshal firebase data: %w", err)
	}
	if err := json.Unmarshal(buf, &zero); err != nil {
		return zero, false, fmt.Errorf("decode firebase data: %w", err)
	}

	return zero, true, nil
}

func Write(ctx context.Context, c *Client, path string, value any) error {
	if c == nil || c.db == nil {
		return fmt.Errorf("firebase client is nil")
	}

	norm, err := normalizePath(path)
	if err != nil {
		return err
	}

	ref := c.db.NewRef(norm)
	return ref.Set(ctx, value)
}
