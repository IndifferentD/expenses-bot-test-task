package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"expenses-bot/backend/internal"
	"expenses-bot/backend/internal/bot"
	"expenses-bot/backend/internal/db"
	"expenses-bot/backend/internal/user"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	TelegramBotToken string `env:"TELEGRAM_BOT_TOKEN,required"`
	TelegramProxyURL string `env:"TELEGRAM_PROXY_URL"`
	DatabaseURL      string `env:"DATABASE_URL,required"`
	LogLevel         string `env:"LOG_LEVEL" envDefault:"INFO"`
}

func main() {
	settings, err := env.ParseAs[Config]()
	if err != nil {
		log.Fatal(err)
	}
	logger, err := internal.NewLogger(settings.LogLevel)
	if err != nil {
		log.Fatal(err)
	}
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	database, err := db.ConnectPostgres(ctx, settings.DatabaseURL)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	userService := user.NewService(user.NewRepository(database))
	telegramClient, err := bot.NewClient(settings.TelegramBotToken, settings.TelegramProxyURL)
	if err != nil {
		slog.Error("failed to initialize Telegram bot", "error", err)
		os.Exit(1)
	}
	updates, err := telegramClient.Updates(ctx)
	if err != nil {
		slog.Error("failed to receive Telegram updates", "error", err)
		os.Exit(1)
	}
	updateHandler := bot.NewUpdateHandler(userService)
	updateHandler.Run(ctx, updates)
	slog.Info("stopped Telegram bot")
}
