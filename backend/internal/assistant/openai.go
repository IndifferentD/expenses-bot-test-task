package assistant

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

func NewOpenAIClient(apiKey, baseURL, proxyAddress string) (openai.Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return openai.Client{}, errors.New("OPENAI_API_KEY is required")
	}
	if strings.TrimSpace(baseURL) == "" {
		return openai.Client{}, errors.New("OPENAI_BASE_URL is required")
	}
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		return openai.Client{}, fmt.Errorf("failed to parse OpenAI base URL: %w", err)
	}
	if parsedBaseURL.Host == "" || (parsedBaseURL.Scheme != "http" && parsedBaseURL.Scheme != "https") {
		return openai.Client{}, errors.New("OPENAI_BASE_URL must be an HTTP URL with a host")
	}
	httpClient := &http.Client{Timeout: 45 * time.Second}
	if proxyAddress != "" {
		proxyURL, err := url.Parse(proxyAddress)
		if err != nil {
			return openai.Client{}, fmt.Errorf("failed to parse OpenAI proxy URL: %w", err)
		}
		if proxyURL.Host == "" {
			return openai.Client{}, errors.New("OpenAI proxy URL must include a host")
		}
		switch proxyURL.Scheme {
		case "http", "https", "socks5", "socks5h":
		default:
			return openai.Client{}, fmt.Errorf("unsupported OpenAI proxy scheme %q", proxyURL.Scheme)
		}
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.Proxy = http.ProxyURL(proxyURL)
		httpClient.Transport = transport
	}
	return openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
		option.WithHTTPClient(httpClient),
	), nil
}

type toolCall struct {
	name      string
	arguments string
}

func (s *Service) selectTool(ctx context.Context, message string, sentAt time.Time) (*toolCall, error) {
	instructions := fmt.Sprintf(
		"Сегодня %s, часовой пояс %s. Выбирай add_expense только если пользователь просит записать новую трату и указал сумму. "+
			"Не вызывай инструмент для вопросов, просмотра, изменения, удаления или отрицания записи. "+
			"Не придумывай сумму, категорию или дату. Относительную дату преобразуй в YYYY-MM-DD; если дата не указана, передай null. "+
			"В comment укажи только предмет покупки или услугу; если их нельзя определить, передай null.",
		sentAt.Format(time.DateOnly), sentAt.Location(),
	)
	response, err := s.client.Responses.New(ctx, responses.ResponseNewParams{
		Model:             s.model,
		Instructions:      openai.String(instructions),
		Input:             responses.ResponseNewParamsInputUnion{OfString: openai.String(message)},
		Tools:             []responses.ToolUnionParam{addExpenseTool()},
		ParallelToolCalls: openai.Bool(false),
		Store:             openai.Bool(false),
	})
	if err != nil {
		var apiErr *openai.Error
		if errors.As(err, &apiErr) {
			return nil, fmt.Errorf("failed to request OpenAI response (code: %q, message: %q): %w", apiErr.Code, apiErr.Message, err)
		}
		return nil, fmt.Errorf("failed to request OpenAI response: %w", err)
	}
	if response.Status != responses.ResponseStatusCompleted {
		return nil, fmt.Errorf("failed to complete OpenAI response: status %s", response.Status)
	}
	var selected *toolCall
	for _, output := range response.Output {
		if output.Type != "function_call" {
			continue
		}
		if selected != nil {
			return nil, fmt.Errorf("failed to select one tool: %w", ErrUnrecognized)
		}
		call := output.AsFunctionCall()
		selected = &toolCall{name: call.Name, arguments: call.Arguments}
	}
	return selected, nil
}

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
