package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"expenses-bot/backend/internal"
	"expenses-bot/backend/internal/expense"

	"github.com/go-chi/chi/v5"
)

type ExpenseHandler struct {
	service *expense.Service
}

func NewExpenseHandler(service *expense.Service) *ExpenseHandler {
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
	filter, err := readExpenseFilter(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	page, err := h.service.List(r.Context(), filter)
	if err != nil {
		writeError(w, r, err)
		return
	}
	var nextCursor *string
	if page.NextCursor != nil {
		value := fmt.Sprintf("%s,%d", page.NextCursor.SpentAt, page.NextCursor.ID)
		nextCursor = &value
	}
	writeJSON(w, http.StatusOK, struct {
		Items      []expense.Expense `json:"items"`
		NextCursor *string           `json:"next_cursor"`
	}{Items: page.Items, NextCursor: nextCursor})
}

func readExpenseFilter(r *http.Request) (expense.ListFilter, error) {
	values := r.URL.Query()
	var filter expense.ListFilter
	userIDs, err := queryList(values, "user_id")
	if err != nil {
		return expense.ListFilter{}, err
	}
	for _, raw := range userIDs {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return expense.ListFilter{}, fmt.Errorf("failed to parse user_id filter: %w", &requestError{
				Code: string(internal.CodeInvalidFilter), Field: "user_id", Reason: "must be a positive integer", Cause: err,
			})
		}
		filter.UserIDs = append(filter.UserIDs, id)
	}
	categories, err := queryList(values, "category")
	if err != nil {
		return expense.ListFilter{}, err
	}
	for _, raw := range categories {
		filter.Categories = append(filter.Categories, expense.Category(raw))
	}
	filter.DateFrom, err = queryDate(values, "date_from")
	if err != nil {
		return expense.ListFilter{}, err
	}
	filter.DateTo, err = queryDate(values, "date_to")
	if err != nil {
		return expense.ListFilter{}, err
	}
	filter.Limit, err = queryLimit(values)
	if err != nil {
		return expense.ListFilter{}, err
	}
	filter.Cursor, err = queryCursor(values)
	if err != nil {
		return expense.ListFilter{}, err
	}
	return filter, nil
}

func queryList(values url.Values, key string) ([]string, error) {
	raw, ok := values[key]
	if !ok {
		return nil, nil
	}
	if len(raw) != 1 {
		return nil, fmt.Errorf("failed to parse %s filter: %w", key, &requestError{
			Code: string(internal.CodeInvalidFilter), Field: key, Reason: "must be specified once as a comma-separated list",
		})
	}
	parts := strings.Split(raw[0], ",")
	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
		if parts[i] == "" {
			return nil, fmt.Errorf("failed to parse %s filter: %w", key, &requestError{
				Code: string(internal.CodeInvalidFilter), Field: key, Reason: "must contain non-empty comma-separated values",
			})
		}
	}
	return parts, nil
}

func queryLimit(values url.Values) (int, error) {
	raw, ok := values["limit"]
	if !ok {
		return 20, nil
	}
	if len(raw) != 1 {
		return 0, fmt.Errorf("failed to parse limit filter: %w", &requestError{
			Code: string(internal.CodeInvalidFilter), Field: "limit", Reason: "must be specified once",
		})
	}
	limit, err := strconv.Atoi(raw[0])
	if err != nil {
		return 0, fmt.Errorf("failed to parse limit filter: %w", &requestError{
			Code: string(internal.CodeInvalidFilter), Field: "limit", Reason: "must be an integer between 1 and 100", Cause: err,
		})
	}
	if limit < 1 || limit > 100 {
		return 0, fmt.Errorf("failed to parse limit filter: %w", &requestError{
			Code: string(internal.CodeInvalidFilter), Field: "limit", Reason: "must be an integer between 1 and 100",
		})
	}
	return limit, nil
}

func queryCursor(values url.Values) (*expense.ListCursor, error) {
	raw, ok := values["cursor"]
	if !ok {
		return nil, nil
	}
	if len(raw) != 1 {
		return nil, fmt.Errorf("failed to parse cursor filter: %w", &requestError{
			Code: string(internal.CodeInvalidFilter), Field: "cursor", Reason: "must be specified once",
		})
	}
	date, idText, ok := strings.Cut(raw[0], ",")
	if !ok {
		return nil, fmt.Errorf("failed to parse cursor filter: %w", &requestError{
			Code: string(internal.CodeInvalidFilter), Field: "cursor", Reason: "must be YYYY-MM-DD,positive ID",
		})
	}
	parsed, dateErr := time.Parse(time.DateOnly, date)
	id, idErr := strconv.ParseInt(idText, 10, 64)
	if dateErr != nil || idErr != nil || id <= 0 {
		return nil, fmt.Errorf("failed to parse cursor filter: %w", &requestError{
			Code: string(internal.CodeInvalidFilter), Field: "cursor", Reason: "must be YYYY-MM-DD,positive ID", Cause: errors.Join(dateErr, idErr),
		})
	}
	return &expense.ListCursor{SpentAt: expense.NewDate(parsed), ID: id}, nil
}

func queryDate(values url.Values, key string) (expense.Date, error) {
	raw, ok := values[key]
	if !ok {
		return expense.Date{}, nil
	}
	if len(raw) != 1 {
		return expense.Date{}, fmt.Errorf("failed to parse %s filter: %w", key, &requestError{
			Code: string(internal.CodeInvalidFilter), Field: key, Reason: "must be specified once",
		})
	}
	parsed, err := time.Parse(time.DateOnly, raw[0])
	if err != nil {
		return expense.Date{}, fmt.Errorf("failed to parse %s filter: %w", key, &requestError{
			Code: string(internal.CodeInvalidFilter), Field: key, Reason: "must be YYYY-MM-DD", Cause: err,
		})
	}
	return expense.NewDate(parsed), nil
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

func readExpense(r *http.Request) (expense.Expense, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return expense.Expense{}, fmt.Errorf("failed to read expense request: %w", &requestError{
			Code: codeInvalidBody, Reason: "failed to read request body", Cause: err,
		})
	}
	var value expense.Expense
	if err := json.Unmarshal(body, &value); err != nil {
		if errors.Is(err, expense.ErrInvalidDate) {
			return expense.Expense{}, fmt.Errorf("failed to decode expense date: %w", &requestError{
				Code: string(internal.CodeInvalidExpense), Field: "spent_at", Reason: "must be YYYY-MM-DD", Cause: err,
			})
		}
		return expense.Expense{}, fmt.Errorf("failed to decode expense request: %w", &requestError{
			Code: codeInvalidJSON, Reason: "invalid JSON request", Cause: err,
		})
	}
	return value, nil
}
