package bot

import (
	"context"
	"errors"
	"fmt"
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
	if message == nil || message.From == nil || message.Text == "" {
		return
	}
	member, err := h.users.GetByTelegramID(ctx, message.From.ID)
	switch {
	case errors.Is(err, user.ErrNotFound):
		slog.Info("ignored message from unknown Telegram user", "telegram_user_id", message.From.ID)
	case err != nil:
		slog.Error("failed to identify Telegram user", "telegram_user_id", message.From.ID, "error", err)
	default:
		if err := h.processMessage(ctx, member.ID, message); err != nil {
			slog.Error("failed to process Telegram message", "user_id", member.ID, "chat_id", message.Chat.ID, "error", err)
		}
	}
}

func (h *UpdateHandler) processMessage(ctx context.Context, userID int64, message *telego.Message) error {
	replies, err := h.assistant.ProcessMessage(ctx, userID, message.Text, time.Unix(message.Date, 0))
	if err != nil {
		if sendErr := h.telegramClient.SendError(ctx, message.Chat.ID, err.Error()); sendErr != nil {
			return fmt.Errorf("failed to send processing error (%v): %w", err, sendErr)
		}
		return err
	}
	for index, reply := range replies {
		if index > 0 {
			// Telegram ограничивает частоту отправки сообщений в одном чате.
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Second):
			}
		}
		if reply.RichHTML != "" {
			err = h.telegramClient.SendRichMessage(ctx, message.Chat.ID, reply.RichHTML)
		} else {
			err = h.telegramClient.SendMessage(ctx, message.Chat.ID, reply.Text)
		}
		if err != nil {
			return fmt.Errorf("failed to send Telegram reply: %w", err)
		}
	}
	return nil
}
