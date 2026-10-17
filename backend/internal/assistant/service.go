package assistant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
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
		slog.Info("added expense from assistant message", "user_id", userID, "expense_id", created.ID, "amount", created.Amount, "category", created.Category, "spent_at", created.SpentAt.String())
		return expenseReply(created), nil
	default:
		return "", fmt.Errorf("failed to execute unknown tool %q", call.name)
	}
}

func expenseReply(created expense.Expense) string {
	if created.Comment != nil {
		return fmt.Sprintf("Добавил: %s — %d ₽, %s, %s.", *created.Comment, created.Amount, categoryName(created.Category), created.SpentAt)
	}
	return fmt.Sprintf("Добавил трату: %d ₽, %s, %s.", created.Amount, categoryName(created.Category), created.SpentAt)
}

func categoryName(category expense.Category) string {
	switch category {
	case expense.CategoryFood:
		return "еда"
	case expense.CategoryTransport:
		return "транспорт"
	case expense.CategoryHousing:
		return "жильё"
	case expense.CategoryEntertainment:
		return "развлечения"
	case expense.CategoryHealth:
		return "здоровье"
	case expense.CategoryShopping:
		return "покупки"
	default:
		return "другое"
	}
}

func (s *Service) addExpense(ctx context.Context, userID int64, arguments string, sentAt time.Time) (expense.Expense, error) {
	var input struct {
		Amount   int64            `json:"amount"`
		Category expense.Category `json:"category"`
		SpentAt  *string          `json:"spent_at"`
		Comment  *string          `json:"comment"`
	}
	decoder := json.NewDecoder(strings.NewReader(arguments))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return expense.Expense{}, fmt.Errorf("failed to decode expense details: %w", err)
	}
	if err := decoder.Decode(new(any)); err == nil {
		return expense.Expense{}, errors.New("failed to decode expense details: unexpected data after JSON object")
	} else if err != io.EOF {
		return expense.Expense{}, fmt.Errorf("failed to decode expense details: %w", err)
	}
	if input.Amount <= 0 {
		return expense.Expense{}, errors.New("failed to validate expense details: amount must be positive")
	}
	if !input.Category.Valid() {
		return expense.Expense{}, errors.New("failed to validate expense details: category is invalid")
	}
	date := sentAt
	if input.SpentAt != nil {
		parsed, err := time.Parse(time.DateOnly, *input.SpentAt)
		if err != nil {
			return expense.Expense{}, fmt.Errorf("failed to parse expense date: %w", err)
		}
		date = parsed
	}
	newExpense := expense.Expense{
		UserID:   userID,
		Amount:   input.Amount,
		Category: input.Category,
		SpentAt:  expense.NewDate(date),
	}
	if input.Comment != nil {
		if comment := strings.TrimSpace(*input.Comment); comment != "" {
			newExpense.Comment = &comment
		}
	}
	created, err := s.expenses.Create(ctx, newExpense)
	if err != nil {
		return expense.Expense{}, fmt.Errorf("failed to save recognized expense: %w", err)
	}
	return created, nil
}
