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

type deleteExpenseInput struct {
	ID       *int64            `json:"id"`
	Amount   *int64            `json:"amount"`
	Category *expense.Category `json:"category"`
	SpentAt  *string           `json:"spent_at"`
	Comment  *string           `json:"comment"`
}

func (s *Service) deleteExpense(ctx context.Context, userID int64, input deleteExpenseInput) (Reply, error) {
	if input.ID != nil {
		if *input.ID <= 0 {
			return Reply{}, errors.New("failed to validate deletion details: ID must be positive")
		}
		return s.deleteOwnedExpense(ctx, userID, *input.ID)
	}

	filter := expense.ListFilter{Filter: expense.Filter{UserIDs: []int64{userID}}, Limit: 5}
	if input.Amount != nil {
		if *input.Amount <= 0 {
			return Reply{}, errors.New("failed to validate deletion details: amount must be positive")
		}
		filter.Amount = *input.Amount
	}
	if input.Category != nil {
		if !input.Category.Valid() {
			return Reply{}, errors.New("failed to validate deletion details: category is invalid")
		}
		filter.Categories = []expense.Category{*input.Category}
	}
	if input.SpentAt != nil {
		date, err := time.Parse(time.DateOnly, *input.SpentAt)
		if err != nil {
			return Reply{}, fmt.Errorf("failed to parse deletion date: %w", err)
		}
		filter.DateFrom = expense.NewDate(date)
		filter.DateTo = filter.DateFrom
	}
	if input.Comment != nil {
		filter.Comment = strings.TrimSpace(*input.Comment)
	}
	if filter.Amount == 0 && len(filter.Categories) == 0 && filter.DateFrom.IsZero() && filter.Comment == "" {
		return Reply{Text: "Уточните, какую трату удалить: укажите описание, сумму, дату или ID."}, nil
	}

	page, err := s.expenses.List(ctx, filter)
	if err != nil {
		return Reply{}, fmt.Errorf("failed to find expenses for deletion: %w", err)
	}
	if len(page.Items) == 0 {
		return Reply{Text: "Не нашёл подходящих трат среди ваших записей."}, nil
	}
	if len(page.Items) == 1 && page.NextCursor == nil {
		return s.deleteOwnedExpense(ctx, userID, page.Items[0].ID)
	}
	return Reply{RichHTML: expenseChoices(page, func(id int64) string { return fmt.Sprintf("удали #%d", id) })}, nil
}

func (s *Service) deleteOwnedExpense(ctx context.Context, userID, id int64) (Reply, error) {
	deleted, err := s.expenses.DeleteForUser(ctx, id, userID)
	if errors.Is(err, expense.ErrNotFound) {
		return Reply{Text: "Не нашёл такую трату среди ваших записей."}, nil
	}
	if err != nil {
		return Reply{}, fmt.Errorf("failed to delete expense %d: %w", id, err)
	}
	slog.Info("deleted expense from assistant message", "user_id", userID, "expense_id", deleted.ID, "amount", deleted.Amount, "category", deleted.Category, "spent_at", deleted.SpentAt.String())
	return Reply{Text: expenseReply("🗑 Трата удалена", deleted)}, nil
}

func expenseChoiceLine(item expense.Expense) string {
	comment := "—"
	if item.Comment != nil {
		comment = strings.Join(strings.Fields(*item.Comment), " ")
		if comment == "" {
			comment = "—"
		}
	}
	runes := []rune(comment)
	if len(runes) > 60 {
		comment = string(runes[:60]) + "…"
	}
	return fmt.Sprintf("#%d · %s ₽ · %s · %s · %s", item.ID, formatAmount(item.Amount), categoryName(item.Category), comment, item.SpentAt.Format("02.01.2006"))
}
