package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/vinovest/sqlx"
)

func ConnectPostgres(ctx context.Context, databaseURL string) (*sqlx.DB, error) {
	config, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database connection string: %w", err)
	}
	config.RuntimeParams["application_name"] = "expenses-api"
	db := sqlx.NewDb(stdlib.OpenDB(*config), "pgx")
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(15 * time.Minute)
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}

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
			return Expense{}, ErrUserNotFound
		}
		return Expense{}, fmt.Errorf("failed to insert expense: %w", err)
	}
	return created, nil
}

func (r *Repository) List(ctx context.Context) ([]Expense, error) {
	expenses := make([]Expense, 0)
	if err := r.db.SelectContext(ctx, &expenses, `
		SELECT id, user_id, amount, category, spent_at,
			comment, created_at, updated_at
		FROM expenses ORDER BY spent_at DESC, id DESC`); err != nil {
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
		return Expense{}, ErrExpenseNotFound
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
		return Expense{}, ErrExpenseNotFound
	}
	if err != nil {
		if hasConstraint(err, "23503", "expenses_user_id_fkey") {
			return Expense{}, ErrUserNotFound
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
		return ErrExpenseNotFound
	}
	return nil
}

func (r *Repository) CreateUser(ctx context.Context, user User) (User, error) {
	var created User
	err := r.db.GetContext(ctx, &created, `
		INSERT INTO users (telegram_user_id, alias)
		VALUES ($1, $2)
		RETURNING id, telegram_user_id, alias`, user.TelegramUserID, user.Alias)
	if hasConstraint(err, "23505", "users_telegram_user_id_key") {
		return User{}, ErrTelegramUserIDTaken
	}
	if err != nil {
		return User{}, fmt.Errorf("failed to insert user: %w", err)
	}
	return created, nil
}

func (r *Repository) ListUsers(ctx context.Context) ([]User, error) {
	users := make([]User, 0)
	if err := r.db.SelectContext(ctx, &users, `
		SELECT id, telegram_user_id, alias
		FROM users ORDER BY alias, id`); err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

func (r *Repository) GetUser(ctx context.Context, id int64) (User, error) {
	var user User
	err := r.db.GetContext(ctx, &user, `
		SELECT id, telegram_user_id, alias
		FROM users WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (r *Repository) UpdateUser(ctx context.Context, user User) (User, error) {
	var updated User
	err := r.db.GetContext(ctx, &updated, `
		UPDATE users SET telegram_user_id = $2, alias = $3
		WHERE id = $1
		RETURNING id, telegram_user_id, alias`, user.ID, user.TelegramUserID, user.Alias)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if hasConstraint(err, "23505", "users_telegram_user_id_key") {
		return User{}, ErrTelegramUserIDTaken
	}
	if err != nil {
		return User{}, fmt.Errorf("failed to update user: %w", err)
	}
	return updated, nil
}

func (r *Repository) DeleteUser(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if hasConstraint(err, "23001", "expenses_user_id_fkey") {
		return ErrUserHasExpenses
	}
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check deleted user: %w", err)
	}
	if count == 0 {
		return ErrUserNotFound
	}
	return nil
}

func hasConstraint(err error, code, name string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == code && pgErr.ConstraintName == name
}
