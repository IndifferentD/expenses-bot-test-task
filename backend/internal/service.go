package internal

import (
	"context"
	"fmt"
	"time"
)

type ExpenseRepository interface {
	Create(context.Context, Expense) (Expense, error)
	List(context.Context) ([]Expense, error)
	Update(context.Context, Expense) (Expense, error)
	Delete(context.Context, int64) error
}

type ExpenseService struct {
	repository ExpenseRepository
}

func NewExpenseService(repository ExpenseRepository) *ExpenseService {
	return &ExpenseService{repository: repository}
}

func (s *ExpenseService) Create(ctx context.Context, expense Expense) (Expense, error) {
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
		return &ValidationError{Code: CodeInvalidExpense, Field: "user_id", Reason: "must be positive"}
	}
	if expense.Amount <= 0 {
		return &ValidationError{Code: CodeInvalidExpense, Field: "amount", Reason: "must be positive"}
	}
	if !expense.Category.Valid() {
		return &ValidationError{Code: CodeInvalidExpense, Field: "category", Reason: "must be a valid category"}
	}
	if _, err := time.Parse(time.DateOnly, expense.SpentAt); err != nil {
		return &ValidationError{Code: CodeInvalidExpense, Field: "spent_at", Reason: "must be YYYY-MM-DD", Cause: err}
	}
	return nil
}

func (s *ExpenseService) List(ctx context.Context) ([]Expense, error) {
	expenses, err := s.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list expenses: %w", err)
	}
	return expenses, nil
}

func (s *ExpenseService) Update(ctx context.Context, expense Expense) (Expense, error) {
	if expense.ID <= 0 {
		return Expense{}, fmt.Errorf("failed to validate expense: %w", &ValidationError{Code: CodeInvalidID, Field: "id", Reason: "must be a positive integer"})
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

func (s *ExpenseService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("failed to validate expense ID: %w", &ValidationError{Code: CodeInvalidID, Field: "id", Reason: "must be a positive integer"})
	}
	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete expense: %w", err)
	}
	return nil
}
