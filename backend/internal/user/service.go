package user

import (
	"context"
	"fmt"
	"strings"

	"expenses-bot/backend/internal"
)

type UserRepository interface {
	Create(context.Context, User) (User, error)
	List(context.Context) ([]User, error)
	Get(context.Context, int64) (User, error)
	GetByTelegramID(context.Context, int64) (User, error)
	Update(context.Context, User) (User, error)
	Delete(context.Context, int64) error
}

type Service struct {
	repository UserRepository
}

func NewService(repository UserRepository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, user User) (User, error) {
	user.Alias = strings.TrimSpace(user.Alias)
	if err := validateUser(user); err != nil {
		return User{}, fmt.Errorf("failed to validate user: %w", err)
	}
	created, err := s.repository.Create(ctx, user)
	if err != nil {
		return User{}, fmt.Errorf("failed to create user: %w", err)
	}
	return created, nil
}

func (s *Service) List(ctx context.Context) ([]User, error) {
	users, err := s.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

func (s *Service) Get(ctx context.Context, id int64) (User, error) {
	if err := validateUserID(id); err != nil {
		return User{}, fmt.Errorf("failed to validate user ID: %w", err)
	}
	user, err := s.repository.Get(ctx, id)
	if err != nil {
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (s *Service) GetByTelegramID(ctx context.Context, telegramID int64) (User, error) {
	user, err := s.repository.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return User{}, fmt.Errorf("failed to get user by Telegram ID: %w", err)
	}
	return user, nil
}

func (s *Service) Update(ctx context.Context, user User) (User, error) {
	if err := validateUserID(user.ID); err != nil {
		return User{}, fmt.Errorf("failed to validate user ID: %w", err)
	}
	user.Alias = strings.TrimSpace(user.Alias)
	if err := validateUser(user); err != nil {
		return User{}, fmt.Errorf("failed to validate user: %w", err)
	}
	updated, err := s.repository.Update(ctx, user)
	if err != nil {
		return User{}, fmt.Errorf("failed to update user: %w", err)
	}
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := validateUserID(id); err != nil {
		return fmt.Errorf("failed to validate user ID: %w", err)
	}
	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func validateUser(user User) error {
	if user.TelegramUserID <= 0 {
		return &internal.ValidationError{Code: internal.CodeInvalidUser, Field: "telegram_user_id", Reason: "must be positive"}
	}
	if user.Alias == "" {
		return &internal.ValidationError{Code: internal.CodeInvalidUser, Field: "alias", Reason: "must not be empty"}
	}
	return nil
}

func validateUserID(id int64) error {
	if id <= 0 {
		return &internal.ValidationError{Code: internal.CodeInvalidID, Field: "id", Reason: "must be a positive integer"}
	}
	return nil
}
