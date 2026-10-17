package assistant

import (
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

func addExpenseTool() responses.ToolUnionParam {
	parameters := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"amount": map[string]any{"type": "integer", "description": "Целая положительная сумма в рублях из сообщения пользователя."},
			"category": map[string]any{
				"type":        "string",
				"enum":        []string{"food", "transport", "housing", "entertainment", "health", "shopping", "other"},
				"description": "Категория: food — еда, transport — транспорт, housing — жильё, entertainment — развлечения, health — здоровье, shopping — покупки, other — другое.",
			},
			"spent_at": map[string]any{"type": []string{"string", "null"}, "description": "Дата траты YYYY-MM-DD, если указана, иначе null."},
			"comment":  map[string]any{"type": []string{"string", "null"}, "description": "Предмет покупки или услуга, если названы, иначе null."},
		},
		"required":             []string{"amount", "category", "spent_at", "comment"},
		"additionalProperties": false,
	}
	tool := responses.ToolParamOfFunction("add_expense", parameters, true)
	tool.OfFunction.Description = openai.String("Записать новую трату пользователя.")
	return tool
}
