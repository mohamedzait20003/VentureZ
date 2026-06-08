package mailer

import (
	"context"
	"log/slog"
)

type LogMailer struct{}

func (LogMailer) Send(_ context.Context, to, subject, _ string, textBody string) error {
	slog.Info("DEV email (not sent)", "to", to, "subject", subject, "body", textBody)
	return nil
}
