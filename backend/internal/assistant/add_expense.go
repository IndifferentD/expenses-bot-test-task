package assistant

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"expenses-bot/backend/internal/expense"
)

type addExpenseInput struct {
	Amount   int64            `json:"amount"`
	Category expense.Category `json:"category"`
	SpentAt  *string          `json:"spent_at"`
	Comment  *string          `json:"comment"`
}

func (s *Service) addExpense(ctx context.Context, userID int64, input addExpenseInput, sentAt time.Time) (Reply, error) {
	if input.Amount <= 0 {
		return Reply{}, errors.New("failed to validate expense details: amount must be positive")
	}
	if !input.Category.Valid() {
		return Reply{}, errors.New("failed to validate expense details: category is invalid")
	}
	date := sentAt
	if input.SpentAt != nil {
		parsed, err := time.Parse(time.DateOnly, *input.SpentAt)
		if err != nil {
			return Reply{}, fmt.Errorf("failed to parse expense date: %w", err)
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
		return Reply{}, fmt.Errorf("failed to save recognized expense: %w", err)
	}
	slog.Info("added expense from assistant message", "user_id", userID, "expense_id", created.ID, "amount", created.Amount, "category", created.Category, "spent_at", created.SpentAt.String())
	return Reply{Text: expenseReply("✅ Трата добавлена", created)}, nil
}
