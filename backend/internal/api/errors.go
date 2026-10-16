package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"expenses-bot/backend/internal"
)

const (
	codeInvalidBody     = "invalid_body"
	codeInvalidJSON     = "invalid_json"
	codeExpenseNotFound = "expense_not_found"
	codeInternalError   = "internal_error"
)

type requestError struct {
	Code   string
	Field  string
	Reason string
	Cause  error
}

func (e *requestError) Error() string {
	message := e.Reason
	if e.Field != "" {
		message = e.Field + " " + message
	}
	if e.Cause != nil {
		return message + ": " + e.Cause.Error()
	}
	return message
}

func (e *requestError) Unwrap() error {
	return e.Cause
}

type problemDetails struct {
	Type          string         `json:"type"`
	Title         string         `json:"title"`
	Status        int            `json:"status"`
	Detail        string         `json:"detail"`
	Instance      string         `json:"instance"`
	Code          string         `json:"code"`
	InvalidParams []invalidParam `json:"invalid-params,omitempty"`
}

type invalidParam struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	problem := problemFromError(err)
	problem.Instance = r.URL.Path
	if problem.Status == http.StatusInternalServerError {
		slog.Error("failed to handle API request", "method", r.Method, "path", r.URL.Path, "error", err)
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(problem.Status)
	if err := json.NewEncoder(w).Encode(problem); err != nil {
		slog.Error("failed to write error response", "error", err)
	}
}

func problemFromError(err error) problemDetails {
	var requestErr *requestError
	var validationErr *internal.ValidationError
	switch {
	case errors.As(err, &requestErr):
		return invalidProblem(requestErr.Code, requestErr.Field, requestErr.Reason)
	case errors.As(err, &validationErr):
		return invalidProblem(string(validationErr.Code), validationErr.Field, validationErr.Reason)
	case errors.Is(err, internal.ErrExpenseNotFound):
		return newProblem(http.StatusNotFound, codeExpenseNotFound, internal.ErrExpenseNotFound.Error())
	default:
		return newProblem(http.StatusInternalServerError, codeInternalError, "internal server error")
	}
}

func newProblem(status int, code, detail string) problemDetails {
	return problemDetails{
		Type:   "about:blank",
		Title:  http.StatusText(status),
		Status: status,
		Detail: detail,
		Code:   code,
	}
}

func invalidProblem(code, field, reason string) problemDetails {
	problem := newProblem(http.StatusBadRequest, code, reason)
	if field != "" {
		problem.Detail = field + " " + reason
		problem.InvalidParams = []invalidParam{{Name: field, Reason: reason}}
	}
	return problem
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Error("failed to write JSON response", "error", err)
	}
}
