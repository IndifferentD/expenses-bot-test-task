package assistant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"expenses-bot/backend/internal/expense"

	"github.com/openai/openai-go/v3"
)

type Service struct {
	client   openai.Client
	model    string
	expenses *expense.Service
	location *time.Location
}

func NewService(client openai.Client, model string, expenses *expense.Service, location *time.Location) (*Service, error) {
	if strings.TrimSpace(model) == "" {
		return nil, errors.New("OPENAI_MODEL is required")
	}
	return &Service{client: client, model: model, expenses: expenses, location: location}, nil
}

func (s *Service) ProcessMessage(ctx context.Context, userID int64, message string, sentAt time.Time) (string, error) {
	sentAt = sentAt.In(s.location)
	call, err := s.selectTool(ctx, message, sentAt)
	if err != nil {
		return "", fmt.Errorf("failed to interpret message: %w", err)
	}
	if call == nil {
		return "Пока я могу только добавлять траты.", nil
	}
	switch call.name {
	case "add_expense":
		created, err := s.addExpense(ctx, userID, call.arguments, sentAt)
		if err != nil {
			return "", fmt.Errorf("failed to execute add_expense: %w", err)
		}
		return expenseReply(created), nil
	default:
		return "", fmt.Errorf("failed to execute unknown tool %q", call.name)
	}
}

func expenseReply(created expense.Expense) string {
	comment := "—"
	if created.Comment != nil {
		comment = *created.Comment
	}
	return fmt.Sprintf(
		"✅ Трата добавлена\n\n💰 Сумма: %s ₽\n🏷 Категория: %s\n📝 Комментарий: %s\n📅 Дата: %s",
		formatAmount(created.Amount), categoryName(created.Category), comment, created.SpentAt.Format("02.01.2006"),
	)
}

func formatAmount(amount int64) string {
	digits := strconv.FormatInt(amount, 10)
	for i := len(digits) - 3; i > 0; i -= 3 {
		digits = digits[:i] + " " + digits[i:]
	}
	return digits
}

func categoryName(category expense.Category) string {
	switch category {
	case expense.CategoryFood:
		return "Еда"
	case expense.CategoryTransport:
		return "Транспорт"
	case expense.CategoryHousing:
		return "Жильё"
	case expense.CategoryEntertainment:
		return "Развлечения"
	case expense.CategoryHealth:
		return "Здоровье"
	case expense.CategoryShopping:
		return "Покупки"
	default:
		return "Другое"
	}
}

func decodeArguments(arguments string, target any) error {
	decoder := json.NewDecoder(strings.NewReader(arguments))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		if err == nil {
			return errors.New("unexpected data after JSON object")
		}
		return err
	}
	return nil
}
