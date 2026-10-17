package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"expenses-bot/backend/internal"
	"expenses-bot/backend/internal/assistant"
	"expenses-bot/backend/internal/bot"
	"expenses-bot/backend/internal/db"
	"expenses-bot/backend/internal/expense"
	"expenses-bot/backend/internal/user"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	TelegramBotToken string `env:"TELEGRAM_BOT_TOKEN,required"`
	TelegramProxyURL string `env:"TELEGRAM_PROXY_URL"`
	DatabaseURL      string `env:"DATABASE_URL,required"`
	LogLevel         string `env:"LOG_LEVEL" envDefault:"INFO"`
	OpenAIAPIKey     string `env:"OPENAI_API_KEY"`
	OpenAIBaseURL    string `env:"OPENAI_BASE_URL" envDefault:"https://api.openai.com/v1"`
	OpenAIModel      string `env:"OPENAI_MODEL" envDefault:"gpt-6-luna"`
	OpenAIProxyURL   string `env:"OPENAI_PROXY_URL"`
	Timezone         string `env:"TIMEZONE" envDefault:"Europe/Moscow"`
}

func main() {
	slog.SetLogLoggerLevel(slog.LevelError)
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	settings, err := env.ParseAs[Config]()
	if err != nil {
		return fmt.Errorf("failed to load bot configuration: %w", err)
	}
	logger, err := internal.NewLogger(settings.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to configure logging: %w", err)
	}
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	database, err := db.ConnectPostgres(ctx, settings.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	location, err := time.LoadLocation(settings.Timezone)
	if err != nil {
		return fmt.Errorf("failed to load timezone %q: %w", settings.Timezone, err)
	}
	expenseService := expense.NewService(expense.NewRepository(database))
	openAIClient, err := assistant.NewOpenAIClient(settings.OpenAIAPIKey, settings.OpenAIBaseURL, settings.OpenAIProxyURL)
	if err != nil {
		return fmt.Errorf("failed to initialize OpenAI client: %w", err)
	}
	assistantService, err := assistant.NewService(openAIClient, settings.OpenAIModel, expenseService, location)
	if err != nil {
		return fmt.Errorf("failed to initialize expense assistant: %w", err)
	}

	telegramClient, err := bot.NewClient(settings.TelegramBotToken, settings.TelegramProxyURL)
	if err != nil {
		return fmt.Errorf("failed to initialize Telegram bot: %w", err)
	}
	updates, err := telegramClient.StartPolling(ctx)
	if err != nil {
		return fmt.Errorf("failed to receive Telegram updates: %w", err)
	}

	userService := user.NewService(user.NewRepository(database))
	updateHandler := bot.NewUpdateHandler(userService, assistantService, telegramClient)
	updateHandler.Run(ctx, updates)

	slog.Info("stopped Telegram bot")
	return nil
}
