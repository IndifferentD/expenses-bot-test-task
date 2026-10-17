package expense

import (
	"context"
	"fmt"

	"expenses-bot/backend/internal"
)

type ExpenseRepository interface {
	Create(context.Context, Expense) (Expense, error)
	List(context.Context) ([]Expense, error)
	Get(context.Context, int64) (Expense, error)
	Update(context.Context, Expense) (Expense, error)
	Delete(context.Context, int64) error
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

func (s *Service) List(ctx context.Context) ([]Expense, error) {
	expenses, err := s.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list expenses: %w", err)
	}
	return expenses, nil
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
