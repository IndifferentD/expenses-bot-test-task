package bot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
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
			return nil, errors.New("Telegram proxy URL must include a host")
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

func (c *Client) Updates(ctx context.Context) (<-chan telego.Update, error) {
	me, err := c.telegramAPI.GetMe(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Telegram bot API: %w", redactedError{cause: err, token: c.telegramAPI.Token()})
	}
	updates, err := c.telegramAPI.UpdatesViaLongPolling(ctx, &telego.GetUpdatesParams{
		Timeout:        30,
		AllowedUpdates: []string{telego.MessageUpdates},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start Telegram polling: %w", redactedError{cause: err, token: c.telegramAPI.Token()})
	}
	slog.Info("starting Telegram bot", "username", me.Username)
	return updates, nil
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
