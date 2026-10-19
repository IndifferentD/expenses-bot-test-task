package bot

import (
	"context"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/mymmrac/telego"
)

type Client struct {
	telegramAPI *telego.Bot
}

func NewClient(token, proxyAddress string) (*Client, error) {
	httpClient := &http.Client{Timeout: 35 * time.Second}
	if proxyAddress != "" {
		proxyURL, err := url.Parse(proxyAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Telegram proxy URL: %w", err)
		}
		if proxyURL.Host == "" {
			return nil, errors.New("telegram proxy URL must include a host")
		}
		switch proxyURL.Scheme {
		case "http", "https", "socks5", "socks5h":
		default:
			return nil, fmt.Errorf("unsupported Telegram proxy scheme %q", proxyURL.Scheme)
		}
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.Proxy = http.ProxyURL(proxyURL)
		httpClient.Transport = transport
	}
	telegramAPI, err := telego.NewBot(token,
		telego.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to configure Telegram bot: %w", err)
	}
	return &Client{telegramAPI: telegramAPI}, nil
}

func (c *Client) StartPolling(ctx context.Context) (<-chan telego.Update, error) {
	me, err := c.telegramAPI.GetMe(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Telegram bot API: %w", err)
	}
	updates, err := c.telegramAPI.UpdatesViaLongPolling(ctx, &telego.GetUpdatesParams{
		Timeout:        30,
		AllowedUpdates: []string{telego.MessageUpdates},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start Telegram polling: %w", err)
	}
	slog.Info("started Telegram bot", "username", me.Username)
	return updates, nil
}

func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) error {
	_, err := c.telegramAPI.SendMessage(ctx, &telego.SendMessageParams{
		ChatID: telego.ChatID{ID: chatID},
		Text:   text,
	})
	if err != nil {
		return fmt.Errorf("failed to send Telegram message: %w", err)
	}
	return nil
}

func (c *Client) SendRichMessage(ctx context.Context, chatID int64, richHTML string) error {
	_, err := c.telegramAPI.SendRichMessage(ctx, &telego.SendRichMessageParams{
		ChatID: telego.ChatID{ID: chatID},
		RichMessage: telego.InputRichMessage{
			HTML: richHTML,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send rich Telegram message: %w", err)
	}
	return nil
}

func (c *Client) SendError(ctx context.Context, chatID int64, message string) error {
	_, err := c.telegramAPI.SendMessage(ctx, &telego.SendMessageParams{
		ChatID:    telego.ChatID{ID: chatID},
		Text:      "Не удалось обработать запрос:\n<pre>" + html.EscapeString(message) + "</pre>",
		ParseMode: telego.ModeHTML,
	})
	if err != nil {
		return fmt.Errorf("failed to send Telegram error message: %w", err)
	}
	return nil
}
