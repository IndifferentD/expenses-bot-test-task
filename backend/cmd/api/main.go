package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"expenses-bot/backend/internal"
	"expenses-bot/backend/internal/api"
	"expenses-bot/backend/internal/db"
	"expenses-bot/backend/internal/expense"
	"expenses-bot/backend/internal/user"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	HTTPAddr    string `env:"HTTP_ADDR" envDefault:":8080"`
	DatabaseURL string `env:"DATABASE_URL,required"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"INFO"`
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("API stopped: %v", err)
	}
}

func run() error {
	settings, err := env.ParseAs[Config]()
	if err != nil {
		return fmt.Errorf("failed to load API configuration: %w", err)
	}
	logger, err := internal.NewLogger(settings.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to configure API logging: %w", err)
	}
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	database, err := db.ConnectPostgres(ctx, settings.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			slog.Error("failed to close database", "error", err)
		}
	}()

	expenseService := expense.NewService(expense.NewRepository(database))
	userService := user.NewService(user.NewRepository(database))
	handlers := api.Handlers{
		Expenses: api.NewExpenseHandler(expenseService),
		Users:    api.NewUserHandler(userService),
	}
	server := &http.Server{
		Addr:              settings.HTTPAddr,
		Handler:           api.NewRouter(logger, handlers),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("failed to shut down HTTP server", "error", err)
		}
	}()

	slog.Info("starting api", "addr", settings.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to serve API: %w", err)
	}
	return nil
}
