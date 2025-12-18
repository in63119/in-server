package google

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"google.golang.org/api/gmail/v1"

	"in-server/pkg/apperr"
	"in-server/pkg/config"
)

type EmailContent struct {
	Recipient string `json:"recipient"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
}

func encodeToBase64URL(value string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(value))
	encoded = strings.TrimRight(encoded, "=")
	encoded = strings.NewReplacer("+", "-", "/", "_").Replace(encoded)
	return encoded
}

func EncodeSubject(subject string) string {
	return fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(subject)))
}

func SendEmail(ctx context.Context, cfg config.Config, content EmailContent) error {
	if ctx == nil {
		ctx = context.Background()
	}

	gmailSvc, sender, err := cfg.NewGmailClient(ctx)
	if err != nil {
		return apperr.Wrap(err, apperr.Email.ErrFailedSendingEmail.Code, apperr.Email.ErrFailedSendingEmail.Message, apperr.Email.ErrFailedSendingEmail.Status)
	}

	rawMessage := strings.Join([]string{
		fmt.Sprintf("From: %s", sender),
		fmt.Sprintf("To: %s", strings.TrimSpace(content.Recipient)),
		fmt.Sprintf("Subject: %s", content.Subject),
		`Content-Type: text/html; charset="UTF-8"`,
		"",
		content.Body,
	}, "\r\n")

	msg := &gmail.Message{Raw: encodeToBase64URL(rawMessage)}
	if _, err := gmailSvc.Users.Messages.Send("me", msg).Do(); err != nil {
		return apperr.Wrap(err, apperr.Email.ErrFailedSendingEmail.Code, apperr.Email.ErrFailedSendingEmail.Message, apperr.Email.ErrFailedSendingEmail.Status)
	}
	return nil
}
