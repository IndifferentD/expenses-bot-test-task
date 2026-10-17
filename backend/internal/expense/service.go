package expense

import (
	"context"
	"fmt"

	"expenses-bot/backend/internal"
)

type ExpenseRepository interface {
	Create(context.Context, Expense) (Expense, error)
	List(context.Context, ListFilter) ([]Expense, error)
	ListFiltered(context.Context, Filter, int) ([]Expense, error)
	Summarize(context.Context, Filter) (Summary, error)
	Get(context.Context, int64) (Expense, error)
	GetForUser(context.Context, int64, int64) (Expense, error)
	GetLatestForUser(context.Context, int64) (Expense, error)
	Update(context.Context, Expense) (Expense, error)
	UpdateForUser(context.Context, int64, int64, ExpenseChanges) (Expense, error)
	Delete(context.Context, int64) error
	DeleteForUser(context.Context, int64, int64) (Expense, error)
}

type Filter struct {
	UserIDs    []int64
	Categories []Category
	DateFrom   Date
	DateTo     Date
	Amount     int64
	Comment    string
}

type ListFilter struct {
	Filter
	Limit  int
	Cursor *ListCursor
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

// ListFiltered возвращает все подходящие траты, если limit равен нулю.
func (s *Service) ListFiltered(ctx context.Context, filter Filter, limit int) ([]Expense, error) {
	if err := validateFilter(filter); err != nil {
		return nil, fmt.Errorf("failed to validate expense filter: %w", err)
	}
	if limit < 0 {
		return nil, fmt.Errorf("failed to validate expense limit: %w", &internal.ValidationError{
			Code: internal.CodeInvalidFilter, Field: "limit", Reason: "must not be negative",
		})
	}
	expenses, err := s.repository.ListFiltered(ctx, filter, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list filtered expenses: %w", err)
	}
	return expenses, nil
}

func validateListFilter(filter ListFilter) error {
	if filter.Limit < 1 {
		return &internal.ValidationError{Code: internal.CodeInvalidFilter, Field: "limit", Reason: "must be positive"}
	}
	if filter.Cursor != nil && (filter.Cursor.ID <= 0 || filter.Cursor.SpentAt.IsZero()) {
		return &internal.ValidationError{Code: internal.CodeInvalidFilter, Field: "cursor", Reason: "must contain a date and positive expense ID"}
	}
	return validateFilter(filter.Filter)
}

func (s *Service) Summarize(ctx context.Context, filter Filter) (Summary, error) {
	if err := validateFilter(filter); err != nil {
		return Summary{}, fmt.Errorf("failed to validate expense filter: %w", err)
	}
	result, err := s.repository.Summarize(ctx, filter)
	if err != nil {
		return Summary{}, fmt.Errorf("failed to summarize expenses: %w", err)
	}
	return result, nil
}

func validateFilter(filter Filter) error {
	if filter.Amount < 0 {
		return &internal.ValidationError{Code: internal.CodeInvalidFilter, Field: "amount", Reason: "must not be negative"}
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

func (s *Service) GetForUser(ctx context.Context, id, userID int64) (Expense, error) {
	if err := validateExpenseID(id); err != nil {
		return Expense{}, fmt.Errorf("failed to validate expense ID: %w", err)
	}
	if err := validateUserID(userID); err != nil {
		return Expense{}, fmt.Errorf("failed to validate expense owner: %w", err)
	}
	expense, err := s.repository.GetForUser(ctx, id, userID)
	if err != nil {
		return Expense{}, fmt.Errorf("failed to get user expense: %w", err)
	}
	return expense, nil
}

func (s *Service) GetLatestForUser(ctx context.Context, userID int64) (Expense, error) {
	if err := validateUserID(userID); err != nil {
		return Expense{}, fmt.Errorf("failed to validate expense owner: %w", err)
	}
	expense, err := s.repository.GetLatestForUser(ctx, userID)
	if err != nil {
		return Expense{}, fmt.Errorf("failed to get latest user expense: %w", err)
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

func (s *Service) UpdateForUser(ctx context.Context, id, userID int64, changes ExpenseChanges) (Expense, error) {
	if err := validateExpenseID(id); err != nil {
		return Expense{}, fmt.Errorf("failed to validate expense ID: %w", err)
	}
	if err := validateUserID(userID); err != nil {
		return Expense{}, fmt.Errorf("failed to validate expense owner: %w", err)
	}
	if err := validateExpenseChanges(changes); err != nil {
		return Expense{}, fmt.Errorf("failed to validate expense changes: %w", err)
	}
	updated, err := s.repository.UpdateForUser(ctx, id, userID, changes)
	if err != nil {
		return Expense{}, fmt.Errorf("failed to update user expense: %w", err)
	}
	return updated, nil
}

func validateExpenseChanges(changes ExpenseChanges) error {
	if changes.Amount == nil && changes.Category == nil && changes.SpentAt == nil && changes.Comment == nil {
		return &internal.ValidationError{Code: internal.CodeInvalidExpense, Reason: "must change at least one field"}
	}
	if changes.Amount != nil && *changes.Amount <= 0 {
		return &internal.ValidationError{Code: internal.CodeInvalidExpense, Field: "amount", Reason: "must be positive"}
	}
	if changes.Category != nil && !changes.Category.Valid() {
		return &internal.ValidationError{Code: internal.CodeInvalidExpense, Field: "category", Reason: "must be a valid category"}
	}
	if changes.SpentAt != nil && changes.SpentAt.IsZero() {
		return &internal.ValidationError{Code: internal.CodeInvalidExpense, Field: "spent_at", Reason: "must be YYYY-MM-DD"}
	}
	return nil
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

func (s *Service) DeleteForUser(ctx context.Context, id, userID int64) (Expense, error) {
	if err := validateExpenseID(id); err != nil {
		return Expense{}, fmt.Errorf("failed to validate expense ID: %w", err)
	}
	if userID <= 0 {
		return Expense{}, fmt.Errorf("failed to validate expense owner: %w", &internal.ValidationError{Code: internal.CodeInvalidID, Field: "user_id", Reason: "must be a positive integer"})
	}
	deleted, err := s.repository.DeleteForUser(ctx, id, userID)
	if err != nil {
		return Expense{}, fmt.Errorf("failed to delete user expense: %w", err)
	}
	return deleted, nil
}

func validateUserID(userID int64) error {
	if userID <= 0 {
		return &internal.ValidationError{Code: internal.CodeInvalidID, Field: "user_id", Reason: "must be a positive integer"}
	}
	return nil
}

func validateExpenseID(id int64) error {
	if id <= 0 {
		return &internal.ValidationError{Code: internal.CodeInvalidID, Field: "id", Reason: "must be a positive integer"}
	}
	return nil
}
