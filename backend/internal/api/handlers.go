package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"expenses-bot/backend/internal"

	"github.com/go-chi/chi/v5"
)

type ExpenseHandler struct {
	service *internal.ExpenseService
}

func NewExpenseHandler(service *internal.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{service: service}
}

func (h *ExpenseHandler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	expense, err := readExpense(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	expense, err = h.service.Create(r.Context(), expense)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, expense)
}

func (h *ExpenseHandler) ListExpenses(w http.ResponseWriter, r *http.Request) {
	expenses, err := h.service.List(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, struct {
		Items []internal.Expense `json:"items"`
	}{Items: expenses})
}

func (h *ExpenseHandler) GetExpense(w http.ResponseWriter, r *http.Request) {
	id, err := expenseID(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	expense, err := h.service.Get(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, expense)
}

func (h *ExpenseHandler) UpdateExpense(w http.ResponseWriter, r *http.Request) {
	id, err := expenseID(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	expense, err := readExpense(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	expense.ID = id

	updated, err := h.service.Update(r.Context(), expense)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *ExpenseHandler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	id, err := expenseID(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if err := h.service.Delete(r.Context(), id); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func expenseID(r *http.Request) (int64, error) {
	return resourceID(r, "expense")
}

func resourceID(r *http.Request, resource string) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("failed to get %s id: %w", resource, &requestError{
			Code: string(internal.CodeInvalidID), Field: "id", Reason: "must be a positive integer", Cause: err,
		})
	}
	return id, nil
}

func readExpense(r *http.Request) (internal.Expense, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return internal.Expense{}, fmt.Errorf("failed to read expense request: %w", &requestError{
			Code: codeInvalidBody, Reason: "failed to read request body", Cause: err,
		})
	}
	var expense internal.Expense
	if err := json.Unmarshal(body, &expense); err != nil {
		if errors.Is(err, internal.ErrInvalidDate) {
			return internal.Expense{}, fmt.Errorf("failed to decode expense date: %w", &requestError{
				Code: string(internal.CodeInvalidExpense), Field: "spent_at", Reason: "must be YYYY-MM-DD", Cause: err,
			})
		}
		return internal.Expense{}, fmt.Errorf("failed to decode expense request: %w", &requestError{
			Code: codeInvalidJSON, Reason: "invalid JSON request", Cause: err,
		})
	}
	return expense, nil
}
