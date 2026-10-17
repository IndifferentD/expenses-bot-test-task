package assistant

import (
	"context"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"strings"
	"time"

	"expenses-bot/backend/internal/expense"
)

type editExpenseInput struct {
	Target struct {
		ID       *int64            `json:"id"`
		Latest   bool              `json:"latest"`
		Amount   *int64            `json:"amount"`
		Category *expense.Category `json:"category"`
		SpentAt  *string           `json:"spent_at"`
		Comment  *string           `json:"comment"`
	} `json:"target"`
	Changes struct {
		Amount   *int64            `json:"amount"`
		Category *expense.Category `json:"category"`
		SpentAt  *string           `json:"spent_at"`
		Comment  *string           `json:"comment"`
	} `json:"changes"`
}

func (s *Service) editExpense(ctx context.Context, userID int64, arguments string) (Reply, error) {
	var input editExpenseInput
	if err := decodeArguments(arguments, &input); err != nil {
		return Reply{}, fmt.Errorf("failed to decode edit details: %w", err)
	}
	if input.Target.ID != nil && input.Target.Latest {
		return Reply{Text: "Уточните, какую трату изменить: по ID или последнюю добавленную."}, nil
	}
	filter, err := editTargetFilter(input, userID)
	if err != nil {
		return Reply{}, err
	}
	changes, err := editChanges(input)
	if err != nil {
		return Reply{}, err
	}
	if changes.Amount == nil && changes.Category == nil && changes.SpentAt == nil && changes.Comment == nil {
		return Reply{Text: "Уточните, что изменить в трате."}, nil
	}

	var selected expense.Expense
	switch {
	case input.Target.ID != nil:
		if *input.Target.ID <= 0 {
			return Reply{}, errors.New("failed to validate edit target: ID must be positive")
		}
		selected, err = s.expenses.GetForUser(ctx, *input.Target.ID, userID)
	case input.Target.Latest:
		selected, err = s.expenses.GetLatestForUser(ctx, userID)
	default:
		if filter.Amount == 0 && len(filter.Categories) == 0 && filter.DateFrom.IsZero() && filter.Comment == "" {
			return Reply{Text: "Уточните, какую трату изменить: укажите описание, старую сумму, дату или ID."}, nil
		}
		page, listErr := s.expenses.List(ctx, expense.ListFilter{Filter: filter, Limit: 5})
		if listErr != nil {
			return Reply{}, fmt.Errorf("failed to find expenses for editing: %w", listErr)
		}
		if len(page.Items) == 0 {
			return Reply{Text: "Не нашёл подходящих трат среди ваших записей."}, nil
		}
		if len(page.Items) != 1 || page.NextCursor != nil {
			return Reply{RichHTML: editExpenseChoices(page, changes)}, nil
		}
		selected = page.Items[0]
	}
	if errors.Is(err, expense.ErrNotFound) {
		return Reply{Text: "Не нашёл подходящих трат среди ваших записей."}, nil
	}
	if err != nil {
		return Reply{}, fmt.Errorf("failed to get expense for editing: %w", err)
	}
	if !matchesEditTarget(selected, filter) {
		return Reply{Text: "Указанные признаки не совпадают с найденной тратой. Уточните её ID или старые значения."}, nil
	}

	updated, err := s.expenses.UpdateForUser(ctx, selected.ID, userID, changes)
	if errors.Is(err, expense.ErrNotFound) {
		return Reply{Text: "Не нашёл подходящих трат среди ваших записей."}, nil
	}
	if err != nil {
		return Reply{}, fmt.Errorf("failed to update expense %d: %w", selected.ID, err)
	}
	slog.Info("updated expense from assistant message", "user_id", userID, "expense_id", updated.ID, "amount", updated.Amount, "category", updated.Category, "spent_at", updated.SpentAt.String())
	return Reply{Text: editedExpenseReply(selected, updated)}, nil
}

func editedExpenseReply(before, after expense.Expense) string {
	return fmt.Sprintf(
		"✏️ Трата #%d изменена\n\n💰 Сумма: %s\n🏷 Категория: %s\n📝 Комментарий: %s\n📅 Дата: %s",
		after.ID,
		changedValue(fmt.Sprintf("%s ₽", formatAmount(before.Amount)), fmt.Sprintf("%s ₽", formatAmount(after.Amount))),
		changedValue(categoryName(before.Category), categoryName(after.Category)),
		changedValue(expenseComment(before), expenseComment(after)),
		changedValue(before.SpentAt.Format("02.01.2006"), after.SpentAt.Format("02.01.2006")),
	)
}

func changedValue(before, after string) string {
	if before == after {
		return after
	}
	return before + " → " + after
}

func expenseComment(item expense.Expense) string {
	if item.Comment == nil || *item.Comment == "" {
		return "—"
	}
	return *item.Comment
}

func editExpenseChoices(page expense.ListPage, changes expense.ExpenseChanges) string {
	var reply strings.Builder
	reply.WriteString("Нашёл несколько трат:\n\n")
	for _, item := range page.Items {
		reply.WriteString(html.EscapeString(expenseChoiceLine(item)))
		reply.WriteByte('\n')
		fmt.Fprintf(&reply, "<code>%s</code>\n\n", html.EscapeString(editExpenseCommand(item.ID, changes)))
	}
	if page.NextCursor != nil {
		reply.WriteString("Показаны первые пять совпадений.\n\n")
	}
	reply.WriteString("Нажмите на нужную строку, скопируйте её и отправьте боту.")
	return reply.String()
}

func editExpenseCommand(id int64, changes expense.ExpenseChanges) string {
	parts := make([]string, 0, 4)
	if changes.Amount != nil {
		parts = append(parts, fmt.Sprintf("сумма %d", *changes.Amount))
	}
	if changes.Category != nil {
		parts = append(parts, "категория "+categoryName(*changes.Category))
	}
	if changes.SpentAt != nil {
		parts = append(parts, "дата "+changes.SpentAt.String())
	}
	if changes.Comment != nil {
		if *changes.Comment == "" {
			parts = append(parts, "убери название")
		} else {
			parts = append(parts, "название «"+strings.Join(strings.Fields(*changes.Comment), " ")+"»")
		}
	}
	return fmt.Sprintf("измени #%d: %s", id, strings.Join(parts, ", "))
}

func editTargetFilter(input editExpenseInput, userID int64) (expense.Filter, error) {
	filter := expense.Filter{UserIDs: []int64{userID}}
	if input.Target.Amount != nil {
		if *input.Target.Amount <= 0 {
			return filter, errors.New("failed to validate old amount: must be positive")
		}
		filter.Amount = *input.Target.Amount
	}
	if input.Target.Category != nil {
		if !input.Target.Category.Valid() {
			return filter, errors.New("failed to validate old category: invalid category")
		}
		filter.Categories = []expense.Category{*input.Target.Category}
	}
	if input.Target.SpentAt != nil {
		date, err := time.Parse(time.DateOnly, *input.Target.SpentAt)
		if err != nil {
			return filter, fmt.Errorf("failed to parse old expense date: %w", err)
		}
		filter.DateFrom = expense.NewDate(date)
		filter.DateTo = filter.DateFrom
	}
	if input.Target.Comment != nil {
		filter.Comment = strings.TrimSpace(*input.Target.Comment)
	}
	return filter, nil
}

func editChanges(input editExpenseInput) (expense.ExpenseChanges, error) {
	changes := expense.ExpenseChanges{Amount: input.Changes.Amount, Category: input.Changes.Category}
	if changes.Amount != nil && *changes.Amount <= 0 {
		return changes, errors.New("failed to validate new amount: must be positive")
	}
	if changes.Category != nil && !changes.Category.Valid() {
		return changes, errors.New("failed to validate new category: invalid category")
	}
	if input.Changes.SpentAt != nil {
		date, err := time.Parse(time.DateOnly, *input.Changes.SpentAt)
		if err != nil {
			return changes, fmt.Errorf("failed to parse new expense date: %w", err)
		}
		spentAt := expense.NewDate(date)
		changes.SpentAt = &spentAt
	}
	if input.Changes.Comment != nil {
		comment := strings.TrimSpace(*input.Changes.Comment)
		changes.Comment = &comment
	}
	return changes, nil
}

func matchesEditTarget(item expense.Expense, filter expense.Filter) bool {
	if filter.Amount != 0 && item.Amount != filter.Amount {
		return false
	}
	if len(filter.Categories) != 0 && item.Category != filter.Categories[0] {
		return false
	}
	if !filter.DateFrom.IsZero() && item.SpentAt.String() != filter.DateFrom.String() {
		return false
	}
	if filter.Comment != "" && (item.Comment == nil || !strings.Contains(strings.ToLower(*item.Comment), strings.ToLower(filter.Comment))) {
		return false
	}
	return true
}
