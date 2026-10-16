package api

import (
	"context"
	"fmt"
	"time"

	"expenses-bot/backend/internal"
)

type ExpenseRepository interface {
	Create(context.Context, internal.Expense) (internal.Expense, error)
	List(context.Context) ([]internal.Expense, error)
	Update(context.Context, internal.Expense) (internal.Expense, error)
	Delete(context.Context, int64) error
}

type ExpenseService struct {
	repository ExpenseRepository
}

func NewExpenseService(repository ExpenseRepository) *ExpenseService {
	return &ExpenseService{repository: repository}
}

func (s *ExpenseService) Create(ctx context.Context, expense internal.Expense) (internal.Expense, error) {
	if err := validateExpense(expense); err != nil {
		return internal.Expense{}, fmt.Errorf("failed to validate expense: %w", err)
	}

	expense, err := s.repository.Create(ctx, expense)
	if err != nil {
		return internal.Expense{}, fmt.Errorf("failed to create expense: %w", err)
	}
	return expense, nil
}

func validateExpense(expense internal.Expense) error {
	if expense.UserID <= 0 {
		return &InputError{Code: CodeInvalidExpense, Field: "user_id", Reason: "must be positive"}
	}
	if expense.Amount <= 0 {
		return &InputError{Code: CodeInvalidExpense, Field: "amount", Reason: "must be positive"}
	}
	if !expense.Category.Valid() {
		return &InputError{Code: CodeInvalidExpense, Field: "category", Reason: "must be a valid category"}
	}
	if _, err := time.Parse(time.DateOnly, expense.SpentAt); err != nil {
		return &InputError{Code: CodeInvalidExpense, Field: "spent_at", Reason: "must be YYYY-MM-DD", Cause: err}
	}
	return nil
}

func (s *ExpenseService) List(ctx context.Context) ([]internal.Expense, error) {
	expenses, err := s.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list expenses: %w", err)
	}
	return expenses, nil
}

func (s *ExpenseService) Update(ctx context.Context, expense internal.Expense) (internal.Expense, error) {
	if expense.ID <= 0 {
		return internal.Expense{}, fmt.Errorf("failed to validate expense: %w", invalidExpenseID(nil))
	}
	if err := validateExpense(expense); err != nil {
		return internal.Expense{}, fmt.Errorf("failed to validate expense: %w", err)
	}
	updated, err := s.repository.Update(ctx, expense)
	if err != nil {
		return internal.Expense{}, fmt.Errorf("failed to update expense: %w", err)
	}
	return updated, nil
}

func (s *ExpenseService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("failed to validate expense ID: %w", invalidExpenseID(nil))
	}
	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete expense: %w", err)
	}
	return nil
}
