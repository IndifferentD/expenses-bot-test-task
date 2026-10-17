package expense

import (
	"context"
	"fmt"

	"expenses-bot/backend/internal"
)

type ExpenseRepository interface {
	Create(context.Context, Expense) (Expense, error)
	List(context.Context, ListFilter) ([]Expense, error)
	Get(context.Context, int64) (Expense, error)
	Update(context.Context, Expense) (Expense, error)
	Delete(context.Context, int64) error
}

type ListFilter struct {
	UserIDs    []int64
	Categories []Category
	DateFrom   Date
	DateTo     Date
	Limit      int
	Cursor     *ListCursor
}

type ListCursor struct {
	SpentAt Date
	ID      int64
}

type ListPage struct {
	Items      []Expense
	NextCursor *ListCursor
}

type Service struct {
	repository ExpenseRepository
}

func NewService(repository ExpenseRepository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, expense Expense) (Expense, error) {
	if err := validateExpense(expense); err != nil {
		return Expense{}, fmt.Errorf("failed to validate expense: %w", err)
	}

	expense, err := s.repository.Create(ctx, expense)
	if err != nil {
		return Expense{}, fmt.Errorf("failed to create expense: %w", err)
	}
	return expense, nil
}

func validateExpense(expense Expense) error {
	if expense.UserID <= 0 {
		return &internal.ValidationError{Code: internal.CodeInvalidExpense, Field: "user_id", Reason: "must be positive"}
	}
	if expense.Amount <= 0 {
		return &internal.ValidationError{Code: internal.CodeInvalidExpense, Field: "amount", Reason: "must be positive"}
	}
	if !expense.Category.Valid() {
		return &internal.ValidationError{Code: internal.CodeInvalidExpense, Field: "category", Reason: "must be a valid category"}
	}
	if expense.SpentAt.IsZero() {
		return &internal.ValidationError{Code: internal.CodeInvalidExpense, Field: "spent_at", Reason: "must be YYYY-MM-DD"}
	}
	return nil
}

func (s *Service) List(ctx context.Context, filter ListFilter) (ListPage, error) {
	if err := validateListFilter(filter); err != nil {
		return ListPage{}, fmt.Errorf("failed to validate expense filter: %w", err)
	}
	expenses, err := s.repository.List(ctx, filter)
	if err != nil {
		return ListPage{}, fmt.Errorf("failed to list expenses: %w", err)
	}
	page := ListPage{Items: expenses}
	if len(expenses) > filter.Limit {
		page.Items = expenses[:filter.Limit]
		last := page.Items[len(page.Items)-1]
		page.NextCursor = &ListCursor{SpentAt: last.SpentAt, ID: last.ID}
	}
	return page, nil
}

func validateListFilter(filter ListFilter) error {
	if filter.Limit < 1 || filter.Limit > 100 {
		return &internal.ValidationError{Code: internal.CodeInvalidFilter, Field: "limit", Reason: "must be between 1 and 100"}
	}
	if filter.Cursor != nil && (filter.Cursor.ID <= 0 || filter.Cursor.SpentAt.IsZero()) {
		return &internal.ValidationError{Code: internal.CodeInvalidFilter, Field: "cursor", Reason: "must contain a date and positive expense ID"}
	}
	for _, id := range filter.UserIDs {
		if id <= 0 {
			return &internal.ValidationError{Code: internal.CodeInvalidFilter, Field: "user_id", Reason: "must be a positive integer"}
		}
	}
	for _, category := range filter.Categories {
		if !category.Valid() {
			return &internal.ValidationError{Code: internal.CodeInvalidFilter, Field: "category", Reason: "must be a valid category"}
		}
	}
	if !filter.DateFrom.IsZero() && !filter.DateTo.IsZero() && filter.DateTo.Before(filter.DateFrom) {
		return &internal.ValidationError{Code: internal.CodeInvalidFilter, Field: "date_to", Reason: "must not be before date_from"}
	}
	return nil
}

func (s *Service) Get(ctx context.Context, id int64) (Expense, error) {
	if err := validateExpenseID(id); err != nil {
		return Expense{}, fmt.Errorf("failed to validate expense ID: %w", err)
	}
	expense, err := s.repository.Get(ctx, id)
	if err != nil {
		return Expense{}, fmt.Errorf("failed to get expense: %w", err)
	}
	return expense, nil
}

func (s *Service) Update(ctx context.Context, expense Expense) (Expense, error) {
	if err := validateExpenseID(expense.ID); err != nil {
		return Expense{}, fmt.Errorf("failed to validate expense: %w", err)
	}
	if err := validateExpense(expense); err != nil {
		return Expense{}, fmt.Errorf("failed to validate expense: %w", err)
	}
	updated, err := s.repository.Update(ctx, expense)
	if err != nil {
		return Expense{}, fmt.Errorf("failed to update expense: %w", err)
	}
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := validateExpenseID(id); err != nil {
		return fmt.Errorf("failed to validate expense ID: %w", err)
	}
	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete expense: %w", err)
	}
	return nil
}

func validateExpenseID(id int64) error {
	if id <= 0 {
		return &internal.ValidationError{Code: internal.CodeInvalidID, Field: "id", Reason: "must be a positive integer"}
	}
	return nil
}
