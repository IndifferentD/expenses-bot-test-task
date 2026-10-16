package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"expenses-bot/backend/internal"
)

type ErrorCode string

const (
	CodeInvalidID       ErrorCode = "invalid_id"
	CodeInvalidBody     ErrorCode = "invalid_body"
	CodeInvalidJSON     ErrorCode = "invalid_json"
	CodeInvalidExpense  ErrorCode = "invalid_expense"
	CodeExpenseNotFound ErrorCode = "expense_not_found"
	CodeInternalError   ErrorCode = "internal_error"
)

type InputError struct {
	Code   ErrorCode
	Field  string
	Reason string
	Cause  error
}

func (e *InputError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s %s", e.Field, e.Reason)
	}
	return e.Reason
}

func (e *InputError) Unwrap() error {
	return e.Cause
}

func invalidExpenseID(cause error) *InputError {
	return &InputError{
		Code: CodeInvalidID, Field: "id", Reason: "must be a positive integer", Cause: cause,
	}
}

type problemDetails struct {
	Type          string         `json:"type"`
	Title         string         `json:"title"`
	Status        int            `json:"status"`
	Detail        string         `json:"detail"`
	Instance      string         `json:"instance"`
	Code          ErrorCode      `json:"code"`
	InvalidParams []invalidParam `json:"invalid-params,omitempty"`
}

type invalidParam struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	problem := problemDetails{
		Type:     "about:blank",
		Status:   http.StatusInternalServerError,
		Detail:   "internal server error",
		Instance: r.URL.Path,
		Code:     CodeInternalError,
	}
	var input *InputError
	switch {
	case errors.As(err, &input):
		problem.Status = http.StatusBadRequest
		problem.Detail = input.Error()
		problem.Code = input.Code
		if input.Field != "" {
			problem.InvalidParams = []invalidParam{{Name: input.Field, Reason: input.Reason}}
		}
	case errors.Is(err, internal.ErrExpenseNotFound):
		problem.Status = http.StatusNotFound
		problem.Detail = internal.ErrExpenseNotFound.Error()
		problem.Code = CodeExpenseNotFound
	default:
		slog.Error("failed to handle API request", "method", r.Method, "path", r.URL.Path, "error", err)
	}
	problem.Title = http.StatusText(problem.Status)
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(problem.Status)
	if err := json.NewEncoder(w).Encode(problem); err != nil {
		slog.Error("failed to write error response", "error", err)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Error("failed to write JSON response", "error", err)
	}
}
