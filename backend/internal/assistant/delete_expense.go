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

func (s *Service) deleteExpense(ctx context.Context, userID int64, arguments string) (string, error) {
	var input deleteExpenseInput
	if err := decodeArguments(arguments, &input); err != nil {
		return "", fmt.Errorf("failed to decode deletion details: %w", err)
	}
	if input.ID != nil {
		if *input.ID <= 0 {
			return "", errors.New("failed to validate deletion details: ID must be positive")
		}
		return s.deleteOwnedExpense(ctx, userID, *input.ID)
	}

	filter := expense.ListFilter{Filter: expense.Filter{UserIDs: []int64{userID}}, Limit: 5}
	if input.Amount != nil {
		if *input.Amount <= 0 {
			return "", errors.New("failed to validate deletion details: amount must be positive")
		}
		filter.Amount = *input.Amount
	}
	if input.Category != nil {
		if !input.Category.Valid() {
			return "", errors.New("failed to validate deletion details: category is invalid")
		}
		filter.Categories = []expense.Category{*input.Category}
	}
	if input.SpentAt != nil {
		date, err := time.Parse(time.DateOnly, *input.SpentAt)
		if err != nil {
			return "", fmt.Errorf("failed to parse deletion date: %w", err)
		}
		filter.DateFrom = expense.NewDate(date)
		filter.DateTo = filter.DateFrom
	}
	if input.Comment != nil {
		filter.Comment = strings.TrimSpace(*input.Comment)
	}
	if filter.Amount == 0 && len(filter.Categories) == 0 && filter.DateFrom.IsZero() && filter.Comment == "" {
		return "Уточните, какую трату удалить: укажите описание, сумму, дату или ID.", nil
	}

	page, err := s.expenses.List(ctx, filter)
	if err != nil {
		return "", fmt.Errorf("failed to find expenses for deletion: %w", err)
	}
	if len(page.Items) == 0 {
		return "Не нашёл подходящих трат среди ваших записей.", nil
	}
	if len(page.Items) == 1 && page.NextCursor == nil {
		return s.deleteOwnedExpense(ctx, userID, page.Items[0].ID)
	}
	return expenseChoices(page, "Для удаления отправьте новое сообщение с номером нужной траты: «удали #номер»."), nil
}

func (s *Service) deleteOwnedExpense(ctx context.Context, userID, id int64) (string, error) {
	deleted, err := s.expenses.DeleteForUser(ctx, id, userID)
	if errors.Is(err, expense.ErrNotFound) {
		return "Не нашёл такую трату среди ваших записей.", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to delete expense %d: %w", id, err)
	}
	slog.Info("deleted expense from assistant message", "user_id", userID, "expense_id", deleted.ID, "amount", deleted.Amount, "category", deleted.Category, "spent_at", deleted.SpentAt.String())
	return expenseReply("🗑 Трата удалена", deleted), nil
}

func expenseChoices(page expense.ListPage, instruction string) string {
	var reply strings.Builder
	reply.WriteString("Нашёл несколько трат:\n\n")
	for _, item := range page.Items {
		reply.WriteString(expenseChoiceLine(item))
		reply.WriteByte('\n')
	}
	if page.NextCursor != nil {
		reply.WriteString("\nПоказаны первые пять совпадений.\n")
	}
	reply.WriteString("\n")
	reply.WriteString(instruction)
	return reply.String()
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
