package bot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"expenses-bot/backend/internal/user"

	"github.com/mymmrac/telego"
)

type Runner struct {
	telegram *telego.Bot
	users    *user.Service
	logger   *slog.Logger
}

func New(token string, users *user.Service, logger *slog.Logger) (*Runner, error) {
	telegram, err := telego.NewBot(token,
		telego.WithHTTPClient(&http.Client{Timeout: 35 * time.Second}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to configure Telegram bot: %w", err)
	}
	return &Runner{telegram: telegram, users: users, logger: logger}, nil
}

func (b *Runner) Run(ctx context.Context) error {
	me, err := b.telegram.GetMe(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to Telegram bot API: %w", redactedError{cause: err, token: b.telegram.Token()})
	}
	updates, err := b.telegram.UpdatesViaLongPolling(ctx, &telego.GetUpdatesParams{
		Timeout:        30,
		AllowedUpdates: []string{telego.MessageUpdates},
	})
	if err != nil {
		return fmt.Errorf("failed to start Telegram polling: %w", redactedError{cause: err, token: b.telegram.Token()})
	}
	b.logger.Info("starting Telegram bot", "username", me.Username)
	for update := range updates {
		message := update.Message
		if message == nil || message.Chat.Type != telego.ChatTypePrivate || message.From == nil || message.Text == "" {
			continue
		}
		member, err := b.users.GetByTelegramID(ctx, message.From.ID)
		switch {
		case errors.Is(err, user.ErrNotFound):
			b.logger.Info("ignored message from unknown Telegram user", "telegram_user_id", message.From.ID)
		case err != nil:
			b.logger.Error("failed to identify Telegram user", "telegram_user_id", message.From.ID, "error", err)
		default:
			b.logger.Info("received Telegram message", "user_id", member.ID, "telegram_user_id", member.TelegramUserID, "text", message.Text)
		}
	}
	b.logger.Info("stopped Telegram bot")
	return nil
}

type redactedError struct {
	cause error
	token string
}

func (e redactedError) Error() string {
	return redactToken(e.cause.Error(), e.token)
}

func (e redactedError) Unwrap() error {
	return e.cause
}

func redactToken(message, token string) string {
	if token == "" {
		return message
	}
	return strings.ReplaceAll(message, token, "[REDACTED]")
}
