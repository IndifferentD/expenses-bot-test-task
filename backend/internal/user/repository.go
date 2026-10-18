package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/vinovest/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, user User) (User, error) {
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

func (r *Repository) List(ctx context.Context) ([]User, error) {
	users := make([]User, 0)
	if err := r.db.SelectContext(ctx, &users, `
		SELECT id, telegram_user_id, alias
		FROM users ORDER BY alias, id`); err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

func (r *Repository) Get(ctx context.Context, id int64) (User, error) {
	var user User
	err := r.db.GetContext(ctx, &user, `
		SELECT id, telegram_user_id, alias
		FROM users WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (r *Repository) GetByTelegramID(ctx context.Context, telegramID int64) (User, error) {
	var user User
	err := r.db.GetContext(ctx, &user, `
		SELECT id, telegram_user_id, alias
		FROM users WHERE telegram_user_id = $1`, telegramID)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("failed to get user by Telegram ID: %w", err)
	}
	return user, nil
}

func (r *Repository) Update(ctx context.Context, user User) (User, error) {
	var updated User
	err := r.db.GetContext(ctx, &updated, `
		UPDATE users SET telegram_user_id = $2, alias = $3
		WHERE id = $1
		RETURNING id, telegram_user_id, alias`, user.ID, user.TelegramUserID, user.Alias)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if hasConstraint(err, "23505", "users_telegram_user_id_key") {
		return User{}, ErrTelegramUserIDTaken
	}
	if err != nil {
		return User{}, fmt.Errorf("failed to update user: %w", err)
	}
	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if hasConstraint(err, "23503", "expenses_user_id_fkey") {
		return ErrHasExpenses
	}
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check deleted user: %w", err)
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func hasConstraint(err error, code, name string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == code && pgErr.ConstraintName == name
}
