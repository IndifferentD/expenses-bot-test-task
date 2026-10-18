package main

import (
	"context"
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
	settings, err := env.ParseAs[Config]()
	if err != nil {
		log.Fatalf("failed to load bot configuration: %v", err)
	}
	logger, err := internal.NewLogger(settings.LogLevel)
	if err != nil {
		log.Fatalf("failed to configure logging: %v", err)
	}
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	database, err := db.ConnectPostgres(ctx, settings.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	location, err := time.LoadLocation(settings.Timezone)
	if err != nil {
		log.Fatalf("failed to load timezone %q: %v", settings.Timezone, err)
	}
	expenseService := expense.NewService(expense.NewRepository(database))
	openAIClient, err := assistant.NewOpenAIClient(settings.OpenAIAPIKey, settings.OpenAIBaseURL, settings.OpenAIProxyURL)
	if err != nil {
		log.Fatalf("failed to initialize OpenAI client: %v", err)
	}
	assistantService, err := assistant.NewService(openAIClient, settings.OpenAIModel, expenseService, location)
	if err != nil {
		log.Fatalf("failed to initialize expense assistant: %v", err)
	}

	telegramClient, err := bot.NewClient(settings.TelegramBotToken, settings.TelegramProxyURL)
	if err != nil {
		log.Fatalf("failed to initialize Telegram bot: %v", err)
	}
	updates, err := telegramClient.StartPolling(ctx)
	if err != nil {
		log.Fatalf("failed to receive Telegram updates: %v", err)
	}

	userService := user.NewService(user.NewRepository(database))
	updateHandler := bot.NewUpdateHandler(userService, assistantService, telegramClient)
	updateHandler.Run(ctx, updates)

	slog.Info("stopped Telegram bot")
}
