package expense

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"expenses-bot/backend/internal/user"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/vinovest/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, expense Expense) (Expense, error) {
	var created Expense
	err := r.db.GetContext(ctx, &created, `
		INSERT INTO expenses (user_id, amount, category, spent_at, comment)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, amount, category, spent_at,
			comment, created_at, updated_at`,
		expense.UserID, expense.Amount, expense.Category, expense.SpentAt, expense.Comment,
	)
	if err != nil {
		if hasConstraint(err, "23503", "expenses_user_id_fkey") {
			return Expense{}, user.ErrNotFound
		}
		return Expense{}, fmt.Errorf("failed to insert expense: %w", err)
	}
	return created, nil
}

func (r *Repository) List(ctx context.Context, filter ListFilter) ([]Expense, error) {
	expenses := make([]Expense, 0)
	query := `
		SELECT id, user_id, amount, category, spent_at,
			comment, created_at, updated_at
		FROM expenses WHERE 1 = 1`
	args := make([]any, 0, 9)
	if len(filter.UserIDs) > 0 {
		query += " AND user_id IN (?)"
		args = append(args, filter.UserIDs)
	}
	if len(filter.Categories) > 0 {
		query += " AND category IN (?)"
		args = append(args, filter.Categories)
	}
	if !filter.DateFrom.IsZero() {
		query += " AND spent_at >= ?"
		args = append(args, filter.DateFrom)
	}
	if !filter.DateTo.IsZero() {
		query += " AND spent_at <= ?"
		args = append(args, filter.DateTo)
	}
	if filter.Amount > 0 {
		query += " AND amount = ?"
		args = append(args, filter.Amount)
	}
	if filter.Comment != "" {
		query += " AND strpos(lower(coalesce(comment, '')), lower(?)) > 0"
		args = append(args, filter.Comment)
	}
	if filter.Cursor != nil {
		query += " AND (spent_at, id) < (?, ?)"
		args = append(args, filter.Cursor.SpentAt, filter.Cursor.ID)
	}
	query += " ORDER BY spent_at DESC, id DESC LIMIT ?"
	args = append(args, filter.Limit+1)
	query, args, err := sqlx.In(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to build expense filter query: %w", err)
	}
	if err := r.db.SelectContext(ctx, &expenses, r.db.Rebind(query), args...); err != nil {
		return nil, fmt.Errorf("failed to list expenses: %w", err)
	}
	return expenses, nil
}

func (r *Repository) Get(ctx context.Context, id int64) (Expense, error) {
	var expense Expense
	err := r.db.GetContext(ctx, &expense, `
		SELECT id, user_id, amount, category, spent_at,
			comment, created_at, updated_at
		FROM expenses WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Expense{}, ErrNotFound
	}
	if err != nil {
		return Expense{}, fmt.Errorf("failed to get expense: %w", err)
	}
	return expense, nil
}

func (r *Repository) Update(ctx context.Context, expense Expense) (Expense, error) {
	var updated Expense
	err := r.db.GetContext(ctx, &updated, `
		UPDATE expenses
		SET user_id = $2, amount = $3, category = $4,
			spent_at = $5, comment = $6, updated_at = now()
		WHERE id = $1
		RETURNING id, user_id, amount, category, spent_at,
			comment, created_at, updated_at`,
		expense.ID, expense.UserID, expense.Amount, expense.Category, expense.SpentAt, expense.Comment,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Expense{}, ErrNotFound
	}
	if err != nil {
		if hasConstraint(err, "23503", "expenses_user_id_fkey") {
			return Expense{}, user.ErrNotFound
		}
		return Expense{}, fmt.Errorf("failed to update expense: %w", err)
	}
	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM expenses WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete expense: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check deleted expense: %w", err)
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteForUser(ctx context.Context, id, userID int64) (Expense, error) {
	var deleted Expense
	err := r.db.GetContext(ctx, &deleted, `
		DELETE FROM expenses WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, amount, category, spent_at,
			comment, created_at, updated_at`, id, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return Expense{}, ErrNotFound
	}
	if err != nil {
		return Expense{}, fmt.Errorf("failed to delete user expense: %w", err)
	}
	return deleted, nil
}

func hasConstraint(err error, code, name string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == code && pgErr.ConstraintName == name
}
