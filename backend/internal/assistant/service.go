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

	"expenses-bot/backend/internal/assistant/llm"
	"expenses-bot/backend/internal/expense"
)

type Interpreter interface {
	Interpret(context.Context, string, time.Time) (*llm.Call, error)
}

// Service выполняет выбранное моделью действие и формирует ответы для Telegram.
type Service struct {
	interpreter Interpreter
	expenses    *expense.Service
	location    *time.Location
}

type Reply struct {
	Text     string
	RichHTML string
}

func NewService(interpreter Interpreter, expenses *expense.Service, location *time.Location) *Service {
	return &Service{interpreter: interpreter, expenses: expenses, location: location}
}

func (s *Service) ProcessMessage(ctx context.Context, userID int64, message string, sentAt time.Time) ([]Reply, error) {
	sentAt = sentAt.In(s.location)
	call, err := s.interpreter.Interpret(ctx, message, sentAt)
	if err != nil {
		return nil, fmt.Errorf("failed to interpret message: %w", err)
	}
	if call == nil {
		// TODO: Учитывать текстовый ответ модели: общая фраза скрывает причину отказа или просьбу уточнить запрос.
		return []Reply{{Text: "Я могу только добавлять, просматривать, изменять и удалять траты. Пожалуйста, уточните запрос."}}, nil
	}
	slog.DebugContext(ctx, "selected assistant tool", "user_id", userID, "tool", call.Name, "arguments", call.Arguments)
	switch call.Name {
	case llm.ToolAddExpense:
		var input addExpenseInput
		if err := decodeArguments(call.Arguments, &input); err != nil {
			return nil, fmt.Errorf("failed to decode expense details: %w", err)
		}
		reply, err := s.addExpense(ctx, userID, input, sentAt)
		if err != nil {
			return nil, fmt.Errorf("failed to execute add_expense: %w", err)
		}
		return []Reply{reply}, nil
	case llm.ToolDeleteExpense:
		var input deleteExpenseInput
		if err := decodeArguments(call.Arguments, &input); err != nil {
			return nil, fmt.Errorf("failed to decode deletion details: %w", err)
		}
		reply, err := s.deleteExpense(ctx, userID, input)
		if err != nil {
			return nil, fmt.Errorf("failed to execute delete_expense: %w", err)
		}
		return []Reply{reply}, nil
	case llm.ToolEditExpense:
		var input editExpenseInput
		if err := decodeArguments(call.Arguments, &input); err != nil {
			return nil, fmt.Errorf("failed to decode edit details: %w", err)
		}
		reply, err := s.editExpense(ctx, userID, input)
		if err != nil {
			return nil, fmt.Errorf("failed to execute edit_expense: %w", err)
		}
		return []Reply{reply}, nil
	case llm.ToolListExpenses:
		var input viewExpensesInput
		if err := decodeArguments(call.Arguments, &input); err != nil {
			return nil, fmt.Errorf("failed to decode expense list request: %w", err)
		}
		replies, err := s.listExpenses(ctx, userID, input)
		if err != nil {
			return nil, fmt.Errorf("failed to execute list_expenses: %w", err)
		}
		return replies, nil
	case llm.ToolSummarizeExpenses:
		var input viewExpensesInput
		if err := decodeArguments(call.Arguments, &input); err != nil {
			return nil, fmt.Errorf("failed to decode expense summary request: %w", err)
		}
		reply, err := s.summarizeExpenses(ctx, userID, input)
		if err != nil {
			return nil, fmt.Errorf("failed to execute summarize_expenses: %w", err)
		}
		return []Reply{reply}, nil
	default:
		return nil, fmt.Errorf("failed to execute unknown tool %q", call.Name)
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
