package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type Handlers struct {
	Expenses *ExpenseHandler
	Users    *UserHandler
}

func NewRouter(logger *slog.Logger, handlers Handlers) http.Handler {
	router := chi.NewRouter()
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.ClientIPFromRemoteAddr)
	router.Use(requestLogger(logger))
	router.Use(chimiddleware.Recoverer)
	router.Use(chimiddleware.Timeout(30 * time.Second))

	router.Route("/api", func(r chi.Router) {
		r.Get("/expenses", handlers.Expenses.ListExpenses)
		r.Get("/expenses/{id}", handlers.Expenses.GetExpense)
		r.Post("/expenses", handlers.Expenses.CreateExpense)
		r.Put("/expenses/{id}", handlers.Expenses.UpdateExpense)
		r.Delete("/expenses/{id}", handlers.Expenses.DeleteExpense)
		r.Get("/users", handlers.Users.ListUsers)
		r.Get("/users/{id}", handlers.Users.GetUser)
		r.Post("/users", handlers.Users.CreateUser)
		r.Put("/users/{id}", handlers.Users.UpdateUser)
		r.Delete("/users/{id}", handlers.Users.DeleteUser)
	})
	return router
}

func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			wrapped := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(wrapped, r)
			status := wrapped.Status()
			if status == 0 {
				status = http.StatusOK
			}
			logger.InfoContext(r.Context(), "http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", status,
				"response_bytes", wrapped.BytesWritten(),
				"duration_ms", time.Since(started).Milliseconds(),
				"client_ip", chimiddleware.GetClientIP(r.Context()),
				"request_id", chimiddleware.GetReqID(r.Context()),
			)
		})
	}
}
