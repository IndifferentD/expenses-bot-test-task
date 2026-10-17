package bot

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"expenses-bot/backend/internal/assistant"
	"expenses-bot/backend/internal/user"

	"github.com/mymmrac/telego"
)

type UpdateHandler struct {
	users          *user.Service
	assistant      *assistant.Service
	telegramClient *Client
}

func NewUpdateHandler(users *user.Service, assistant *assistant.Service, telegramClient *Client) *UpdateHandler {
	return &UpdateHandler{users: users, assistant: assistant, telegramClient: telegramClient}
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
		reply, err := h.assistant.ProcessMessage(ctx, member.ID, message.Text, time.Unix(message.Date, 0))
		if err != nil {
			if errors.Is(err, assistant.ErrUnrecognized) {
				slog.Warn("failed to interpret Telegram message", "user_id", member.ID, "error", err)
			} else {
				slog.Error("failed to process Telegram message", "user_id", member.ID, "error", err)
			}
			if sendErr := h.telegramClient.SendError(ctx, message.Chat.ID, err.Error()); sendErr != nil {
				slog.Error("failed to respond to Telegram message", "chat_id", message.Chat.ID, "error", sendErr)
			}
			return
		}
		if err := h.telegramClient.SendMessage(ctx, message.Chat.ID, reply); err != nil {
			slog.Error("failed to respond to Telegram message", "chat_id", message.Chat.ID, "error", err)
		}
	}
}
