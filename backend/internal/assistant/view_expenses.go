package assistant

import (
	"context"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"

	"expenses-bot/backend/internal/expense"
)

type viewExpensesInput struct {
	Category *expense.Category `json:"category"`
	Comment  *string           `json:"comment"`
	DateFrom *string           `json:"date_from"`
	DateTo   *string           `json:"date_to"`
	Limit    *int              `json:"limit"`
	All      bool              `json:"all"`
}

const expenseRowsPerMessage = 50

func (s *Service) listExpenses(ctx context.Context, userID int64, arguments string) ([]Reply, error) {
	var input viewExpensesInput
	if err := decodeArguments(arguments, &input); err != nil {
		return nil, fmt.Errorf("failed to decode expense list request: %w", err)
	}
	filter, err := input.filter(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare expense list: %w", err)
	}
	limit := 10
	if input.All {
		limit = 0
	}
	if !input.All && input.Limit != nil {
		if *input.Limit <= 0 {
			return nil, errors.New("failed to validate expense list: limit must be positive")
		}
		limit = *input.Limit
	}
	items, err := s.expenses.ListFiltered(ctx, filter, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list expenses: %w", err)
	}
	if len(items) == 0 {
		return []Reply{{Text: "Трат по запросу не нашёл."}}, nil
	}
	var messages []Reply
	for start := 0; start < len(items); start += expenseRowsPerMessage {
		end := min(start+expenseRowsPerMessage, len(items))
		messages = append(messages, Reply{RichHTML: expenseListMessage(items[start:end], start > 0)})
	}
	return messages, nil
}

func (s *Service) summarizeExpenses(ctx context.Context, userID int64, arguments string) (string, error) {
	var input viewExpensesInput
	if err := decodeArguments(arguments, &input); err != nil {
		return "", fmt.Errorf("failed to decode expense summary request: %w", err)
	}
	filter, err := input.filter(userID)
	if err != nil {
		return "", fmt.Errorf("failed to prepare expense summary: %w", err)
	}
	summary, err := s.expenses.Summarize(ctx, filter)
	if err != nil {
		return "", fmt.Errorf("failed to summarize expenses: %w", err)
	}
	var reply strings.Builder
	reply.WriteString("📊 Сводка по тратам\n")
	if input.Category != nil {
		fmt.Fprintf(&reply, "🏷 Категория: %s\n", categoryName(*input.Category))
	}
	if filter.Comment != "" {
		fmt.Fprintf(&reply, "📝 Поиск: %s\n", filter.Comment)
	}
	if !filter.DateFrom.IsZero() || !filter.DateTo.IsZero() {
		fmt.Fprintf(&reply, "📅 Период: %s — %s\n", formatFilterDate(filter.DateFrom), formatFilterDate(filter.DateTo))
	}
	fmt.Fprintf(&reply, "\n💰 Сумма: %s ₽\n🧾 Трат: %d", formatAmount(summary.Total), summary.Count)
	return reply.String(), nil
}

func (input viewExpensesInput) filter(userID int64) (expense.Filter, error) {
	filter := expense.Filter{UserIDs: []int64{userID}}
	if input.Category != nil {
		filter.Categories = []expense.Category{*input.Category}
	}
	if input.Comment != nil {
		filter.Comment = strings.TrimSpace(*input.Comment)
	}
	if input.DateFrom != nil {
		date, err := time.Parse(time.DateOnly, *input.DateFrom)
		if err != nil {
			return expense.Filter{}, fmt.Errorf("failed to parse start date: %w", err)
		}
		filter.DateFrom = expense.NewDate(date)
	}
	if input.DateTo != nil {
		date, err := time.Parse(time.DateOnly, *input.DateTo)
		if err != nil {
			return expense.Filter{}, fmt.Errorf("failed to parse end date: %w", err)
		}
		filter.DateTo = expense.NewDate(date)
	}
	return filter, nil
}

func expenseListComment(item expense.Expense) string {
	comment := "—"
	if item.Comment != nil {
		comment = strings.Join(strings.Fields(*item.Comment), " ")
		if comment == "" {
			comment = "—"
		}
	}
	runes := []rune(comment)
	if len(runes) > 32 {
		comment = string(runes[:32]) + "…"
	}
	return comment
}

func expenseListMessage(items []expense.Expense, continuation bool) string {
	title := "Траты:"
	if continuation {
		title = "Траты (продолжение):"
	}
	var message strings.Builder
	message.WriteString("<table bordered striped compact><caption>")
	message.WriteString(html.EscapeString(title))
	message.WriteString("</caption><tr><th>ID</th><th>Сумма ₽</th><th>Категория</th><th>Дата</th><th>Комментарий</th></tr>")
	for _, item := range items {
		fmt.Fprintf(&message, "<tr><td>#%d</td><td align=\"right\">%s</td><td>%s</td><td>%s</td><td>%s</td></tr>",
			item.ID, formatAmount(item.Amount), html.EscapeString(categoryName(item.Category)), item.SpentAt.Format("02.01.2006"), html.EscapeString(expenseListComment(item)))
	}
	message.WriteString("</table>")
	return message.String()
}

func formatFilterDate(date expense.Date) string {
	if date.IsZero() {
		return "…"
	}
	return date.Format("02.01.2006")
}
