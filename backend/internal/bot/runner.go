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

type Runner struct {
	telegram *telego.Bot
	handler  *UpdateHandler
	logger   *slog.Logger
}

func New(token, proxyAddress string, handler *UpdateHandler, logger *slog.Logger) (*Runner, error) {
	client := &http.Client{Timeout: 35 * time.Second}
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
		client.Transport = transport
	}
	telegram, err := telego.NewBot(token,
		telego.WithHTTPClient(client),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to configure Telegram bot: %w", err)
	}
	return &Runner{telegram: telegram, handler: handler, logger: logger}, nil
}

func (r *Runner) Run(ctx context.Context) error {
	me, err := r.telegram.GetMe(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to Telegram bot API: %w", redactedError{cause: err, token: r.telegram.Token()})
	}
	updates, err := r.telegram.UpdatesViaLongPolling(ctx, &telego.GetUpdatesParams{
		Timeout:        30,
		AllowedUpdates: []string{telego.MessageUpdates},
	})
	if err != nil {
		return fmt.Errorf("failed to start Telegram polling: %w", redactedError{cause: err, token: r.telegram.Token()})
	}
	r.logger.Info("starting Telegram bot", "username", me.Username)
	for update := range updates {
		r.handler.Handle(ctx, update)
	}
	r.logger.Info("stopped Telegram bot")
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
