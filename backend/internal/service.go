package internal

import (
	"context"
	"fmt"
	"strings"
)

type ExpenseRepository interface {
	Create(context.Context, Expense) (Expense, error)
	List(context.Context) ([]Expense, error)
	Get(context.Context, int64) (Expense, error)
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
	if expense.SpentAt.IsZero() {
		return &ValidationError{Code: CodeInvalidExpense, Field: "spent_at", Reason: "must be YYYY-MM-DD"}
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

func (s *ExpenseService) Get(ctx context.Context, id int64) (Expense, error) {
	if err := validateExpenseID(id); err != nil {
		return Expense{}, fmt.Errorf("failed to validate expense ID: %w", err)
	}
	expense, err := s.repository.Get(ctx, id)
	if err != nil {
		return Expense{}, fmt.Errorf("failed to get expense: %w", err)
	}
	return expense, nil
}

func (s *ExpenseService) Update(ctx context.Context, expense Expense) (Expense, error) {
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

func (s *ExpenseService) Delete(ctx context.Context, id int64) error {
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
		return &ValidationError{Code: CodeInvalidID, Field: "id", Reason: "must be a positive integer"}
	}
	return nil
}

type UserRepository interface {
	CreateUser(context.Context, User) (User, error)
	ListUsers(context.Context) ([]User, error)
	GetUser(context.Context, int64) (User, error)
	UpdateUser(context.Context, User) (User, error)
	DeleteUser(context.Context, int64) error
}

type UserService struct {
	repository UserRepository
}

func NewUserService(repository UserRepository) *UserService {
	return &UserService{repository: repository}
}

func (s *UserService) Create(ctx context.Context, user User) (User, error) {
	user.Alias = strings.TrimSpace(user.Alias)
	if err := validateUser(user); err != nil {
		return User{}, fmt.Errorf("failed to validate user: %w", err)
	}
	created, err := s.repository.CreateUser(ctx, user)
	if err != nil {
		return User{}, fmt.Errorf("failed to create user: %w", err)
	}
	return created, nil
}

func (s *UserService) List(ctx context.Context) ([]User, error) {
	users, err := s.repository.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

func (s *UserService) Get(ctx context.Context, id int64) (User, error) {
	if err := validateUserID(id); err != nil {
		return User{}, fmt.Errorf("failed to validate user ID: %w", err)
	}
	user, err := s.repository.GetUser(ctx, id)
	if err != nil {
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (s *UserService) Update(ctx context.Context, user User) (User, error) {
	if err := validateUserID(user.ID); err != nil {
		return User{}, fmt.Errorf("failed to validate user ID: %w", err)
	}
	user.Alias = strings.TrimSpace(user.Alias)
	if err := validateUser(user); err != nil {
		return User{}, fmt.Errorf("failed to validate user: %w", err)
	}
	updated, err := s.repository.UpdateUser(ctx, user)
	if err != nil {
		return User{}, fmt.Errorf("failed to update user: %w", err)
	}
	return updated, nil
}

func (s *UserService) Delete(ctx context.Context, id int64) error {
	if err := validateUserID(id); err != nil {
		return fmt.Errorf("failed to validate user ID: %w", err)
	}
	if err := s.repository.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func validateUser(user User) error {
	if user.TelegramUserID <= 0 {
		return &ValidationError{Code: CodeInvalidUser, Field: "telegram_user_id", Reason: "must be positive"}
	}
	if user.Alias == "" {
		return &ValidationError{Code: CodeInvalidUser, Field: "alias", Reason: "must not be empty"}
	}
	return nil
}

func validateUserID(id int64) error {
	if id <= 0 {
		return &ValidationError{Code: CodeInvalidID, Field: "id", Reason: "must be a positive integer"}
	}
	return nil
}
