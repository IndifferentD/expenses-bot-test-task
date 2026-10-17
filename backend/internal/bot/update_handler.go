package bot

import (
	"context"
	"errors"
	"log/slog"

	"expenses-bot/backend/internal/user"

	"github.com/mymmrac/telego"
)

type UpdateHandler struct {
	users *user.Service
}

func NewUpdateHandler(users *user.Service) *UpdateHandler {
	return &UpdateHandler{users: users}
}

func (h *UpdateHandler) Run(ctx context.Context, updates <-chan telego.Update) {
	for update := range updates {
		h.Handle(ctx, update)
	}
}

func (h *UpdateHandler) Handle(ctx context.Context, update telego.Update) {
	message := update.Message
	if message == nil || message.Chat.Type != telego.ChatTypePrivate || message.From == nil || message.Text == "" {
		return
	}
	member, err := h.users.GetByTelegramID(ctx, message.From.ID)
	switch {
	case errors.Is(err, user.ErrNotFound):
		slog.Info("ignored message from unknown Telegram user", "telegram_user_id", message.From.ID)
	case err != nil:
		slog.Error("failed to identify Telegram user", "telegram_user_id", message.From.ID, "error", err)
	default:
		slog.Info("received Telegram message", "user_id", member.ID, "telegram_user_id", member.TelegramUserID, "text", message.Text)
	}
}
