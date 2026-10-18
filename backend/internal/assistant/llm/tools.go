package llm

import (
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

const (
	ToolAddExpense        = "add_expense"
	ToolDeleteExpense     = "delete_expense"
	ToolEditExpense       = "edit_expense"
	ToolListExpenses      = "list_expenses"
	ToolSummarizeExpenses = "summarize_expenses"
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
	tool := responses.ToolParamOfFunction(ToolAddExpense, parameters, true)
	tool.OfFunction.Description = openai.String("Записать новую трату пользователя.")
	return tool
}

func deleteExpenseTool() responses.ToolUnionParam {
	parameters := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"id":       map[string]any{"type": []string{"integer", "null"}, "description": "ID траты после #, если указан, иначе null."},
			"amount":   map[string]any{"type": []string{"integer", "null"}, "description": "Сумма траты, если названа, иначе null."},
			"category": map[string]any{"type": []string{"string", "null"}, "enum": []any{"food", "transport", "housing", "entertainment", "health", "shopping", "other", nil}, "description": "Категория траты, если названа, иначе null."},
			"spent_at": map[string]any{"type": []string{"string", "null"}, "description": "Дата траты YYYY-MM-DD, если названа, иначе null."},
			"comment":  map[string]any{"type": []string{"string", "null"}, "description": "Предмет покупки или услуга, если названы, иначе null."},
		},
		"required":             []string{"id", "amount", "category", "spent_at", "comment"},
		"additionalProperties": false,
	}
	tool := responses.ToolParamOfFunction(ToolDeleteExpense, parameters, true)
	tool.OfFunction.Description = openai.String("Удалить трату по указанному ID или признакам из сообщения пользователя.")
	return tool
}

func editExpenseTool() responses.ToolUnionParam {
	category := map[string]any{
		"type": []string{"string", "null"},
		"enum": []any{"food", "transport", "housing", "entertainment", "health", "shopping", "other", nil},
	}
	parameters := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"target": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id":       map[string]any{"type": []string{"integer", "null"}, "description": "ID после #, иначе null."},
					"latest":   map[string]any{"type": "boolean", "description": "True только при явном указании на последнюю добавленную трату."},
					"amount":   map[string]any{"type": []string{"integer", "null"}, "description": "Старая сумма, если названа, иначе null."},
					"category": category,
					"spent_at": map[string]any{"type": []string{"string", "null"}, "description": "Старая дата YYYY-MM-DD, если названа, иначе null."},
					"comment":  map[string]any{"type": []string{"string", "null"}, "description": "Дословный фрагмент старого описания, если назван, иначе null."},
				},
				"required":             []string{"id", "latest", "amount", "category", "spent_at", "comment"},
				"additionalProperties": false,
			},
			"changes": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"amount":   map[string]any{"type": []string{"integer", "null"}, "description": "Новая сумма, если меняется, иначе null."},
					"category": category,
					"spent_at": map[string]any{"type": []string{"string", "null"}, "description": "Новая дата YYYY-MM-DD, если меняется, иначе null."},
					"comment":  map[string]any{"type": []string{"string", "null"}, "description": "Новое название или описание траты; null — не менять, пустая строка — очистить."},
				},
				"required":             []string{"amount", "category", "spent_at", "comment"},
				"additionalProperties": false,
			},
		},
		"required":             []string{"target", "changes"},
		"additionalProperties": false,
	}
	tool := responses.ToolParamOfFunction(ToolEditExpense, parameters, true)
	tool.OfFunction.Description = openai.String("Изменить поля существующей траты пользователя: target описывает старую запись, changes — новые значения.")
	return tool
}

func listExpensesTool() responses.ToolUnionParam {
	properties := viewFilterProperties()
	properties["limit"] = map[string]any{"type": []string{"integer", "null"}, "description": "Сколько трат показать: 0 — все, null — обычное число."}
	parameters := map[string]any{
		"type":                 "object",
		"properties":           properties,
		"required":             []string{"categories", "comment", "date_from", "date_to", "limit"},
		"additionalProperties": false,
	}
	tool := responses.ToolParamOfFunction(ToolListExpenses, parameters, true)
	tool.OfFunction.Description = openai.String("Показать траты пользователя с фильтром по категориям и периоду.")
	return tool
}

func summarizeExpensesTool() responses.ToolUnionParam {
	parameters := map[string]any{
		"type":                 "object",
		"properties":           viewFilterProperties(),
		"required":             []string{"categories", "comment", "date_from", "date_to"},
		"additionalProperties": false,
	}
	tool := responses.ToolParamOfFunction(ToolSummarizeExpenses, parameters, true)
	tool.OfFunction.Description = openai.String("Посчитать сумму и количество трат пользователя за период или по категориям.")
	return tool
}

func viewFilterProperties() map[string]any {
	return map[string]any{
		"categories": map[string]any{
			"type":        "array",
			"items":       map[string]any{"type": "string", "enum": []string{"food", "transport", "housing", "entertainment", "health", "shopping", "other"}},
			"description": "Все названные категории; если категорий нет, пустой массив.",
		},
		"comment":   map[string]any{"type": []string{"string", "null"}, "description": "Предмет покупки или услуга из запроса, иначе null."},
		"date_from": map[string]any{"type": []string{"string", "null"}, "description": "Начало периода YYYY-MM-DD включительно, иначе null."},
		"date_to":   map[string]any{"type": []string{"string", "null"}, "description": "Конец периода YYYY-MM-DD включительно, иначе null."},
	}
}
