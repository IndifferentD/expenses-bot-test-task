package assistant

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	"expenses-bot/backend/internal/expense"
)

func expenseReply(title string, item expense.Expense) string {
	return fmt.Sprintf(
		"%s\n\n💰 Сумма: %s ₽\n🏷 Категория: %s\n📝 Комментарий: %s\n📅 Дата: %s",
		title, formatAmount(item.Amount), categoryName(item.Category), expenseComment(item), item.SpentAt.Format("02.01.2006"),
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

func expenseComment(item expense.Expense) string {
	if item.Comment == nil || *item.Comment == "" {
		return "—"
	}
	return *item.Comment
}

func expenseChoices(page expense.ListPage, command func(int64) string) string {
	var reply strings.Builder
	reply.WriteString("<p>Нашёл несколько трат:</p>")
	for _, item := range page.Items {
		reply.WriteString("<p>")
		reply.WriteString(html.EscapeString(expenseChoiceLine(item)))
		fmt.Fprintf(&reply, "<br><code>%s</code></p>", html.EscapeString(command(item.ID)))
	}
	if page.NextCursor != nil {
		reply.WriteString("<p>Показаны первые пять совпадений.</p>")
	}
	reply.WriteString("<p>Скопируйте команду под нужной тратой и отправьте боту.</p>")
	return reply.String()
}
