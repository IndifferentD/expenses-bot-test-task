package internal

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func OpenSQLite(ctx context.Context, path string) (*sqlx.DB, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve database path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	dsn := url.URL{Scheme: "file", Path: absPath}
	dsn.RawQuery = url.Values{
		"_busy_timeout": {"5000"},
		"_journal_mode": {"WAL"},
	}.Encode()
	db, err := sqlx.Open("sqlite3", dsn.String())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	db.SetMaxOpenConns(1)
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}

type Repository struct {
	db *sqlx.DB
}

var ErrExpenseNotFound = errors.New("expense not found")

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, expense Expense) (Expense, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	expense.CreatedAt = now
	expense.UpdatedAt = now
	result, err := r.db.NamedExecContext(ctx, `
		INSERT INTO expenses (user_id, amount, category, spent_at, comment, created_at, updated_at)
		VALUES (:user_id, :amount, :category, :spent_at, :comment, :created_at, :updated_at)`, expense)
	if err != nil {
		return Expense{}, fmt.Errorf("failed to insert expense: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Expense{}, fmt.Errorf("failed to get ID of inserted expense: %w", err)
	}
	expense.ID = id
	return expense, nil
}

func (r *Repository) List(ctx context.Context) ([]Expense, error) {
	expenses := make([]Expense, 0)
	if err := r.db.SelectContext(ctx, &expenses, `
		SELECT id, user_id, amount, category, spent_at, comment, created_at, updated_at
		FROM expenses ORDER BY spent_at DESC, id DESC`); err != nil {
		return nil, fmt.Errorf("failed to list expenses: %w", err)
	}
	return expenses, nil
}

func (r *Repository) Update(ctx context.Context, expense Expense) (Expense, error) {
	expense.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	result, err := r.db.NamedExecContext(ctx, `
		UPDATE expenses
		SET user_id = :user_id, amount = :amount, category = :category,
			spent_at = :spent_at, comment = :comment, updated_at = :updated_at
		WHERE id = :id`, expense)
	if err != nil {
		return Expense{}, fmt.Errorf("failed to update expense: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return Expense{}, fmt.Errorf("failed to check updated expense: %w", err)
	}
	if count == 0 {
		return Expense{}, ErrExpenseNotFound
	}
	var updated Expense
	if err := r.db.GetContext(ctx, &updated, `
		SELECT id, user_id, amount, category, spent_at, comment, created_at, updated_at
		FROM expenses WHERE id = ?`, expense.ID); err != nil {
		return Expense{}, fmt.Errorf("failed to get updated expense: %w", err)
	}
	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM expenses WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete expense: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check deleted expense: %w", err)
	}
	if count == 0 {
		return ErrExpenseNotFound
	}
	return nil
}
