package assistant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
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
		"Сегодня %s, часовой пояс %s. Если сообщение содержит описание траты и сумму, но не указывает явно другое действие, считай его просьбой добавить трату. "+
			"В таком случае сразу вызывай add_expense без запроса подтверждения. "+
			"Выбирай delete_expense, только если пользователь явно просит удалить трату. "+
			"Выбирай edit_expense, если пользователь просит исправить существующую трату. В target передавай старые признаки, в changes — только новые значения. "+
			"latest=true ставь только при явном указании на последнюю добавленную трату; новую сумму не используй для её поиска. "+
			"Для неизменяемых полей в changes передавай null; changes.comment=\"\" только при явной просьбе убрать название или описание траты. "+
			"Выбирай list_expenses, когда пользователь просит показать записи, и summarize_expenses, когда спрашивает сумму или число трат. "+
			"Для списка передавай all=true, если пользователь просит все записи или указывает фильтр без ограничения числа. При запросе последних N передавай all=false и limit=N; иначе all=false и limit=null. "+
			"Для просмотра и сводки передавай все названные категории в categories, а если категорий нет — пустой массив; comment — если назван конкретный предмет или услуга. "+
			"Границы запрошенного периода включительны; для текущего периода конец — сегодня, а если период не указан, передавай null. Не вызывай инструменты для отрицания записи и посторонних вопросов. "+
			"Не придумывай сумму, конкретную дату траты или ID. Категорию новой траты определяй по покупке или услуге, а если она неясна — выбирай other. Относительные даты и периоды преобразуй в YYYY-MM-DD. "+
			"В comment укажи только предмет покупки или услугу; если их нельзя определить, передай null. "+
			"Для удаления и поиска изменяемой траты извлекай только явно названные старые признаки: comment — короткая дословная часть описания, без перефразирования; "+
			"category заполняй только если категория названа прямо. Число после # означает ID траты.",
		sentAt.Format(time.DateOnly), sentAt.Location(),
	)
	response, err := s.client.Responses.New(ctx, responses.ResponseNewParams{
		Model:             s.model,
		Instructions:      openai.String(instructions),
		Input:             responses.ResponseNewParamsInputUnion{OfString: openai.String(message)},
		Tools:             []responses.ToolUnionParam{addExpenseTool(), editExpenseTool(), deleteExpenseTool(), listExpensesTool(), summarizeExpensesTool()},
		ParallelToolCalls: openai.Bool(false),
		Store:             openai.Bool(false),
	})
	if err != nil {
		var apiErr *openai.Error
		if errors.As(err, &apiErr) {
			if slog.Default().Enabled(ctx, slog.LevelDebug) {
				slog.DebugContext(ctx, "received OpenAI error response", "model", s.model, "input", message, "response", json.RawMessage(apiErr.RawJSON()))
			}
			return nil, fmt.Errorf("failed to request OpenAI response (code: %q, message: %q): %w", apiErr.Code, apiErr.Message, err)
		}
		return nil, fmt.Errorf("failed to request OpenAI response: %w", err)
	}
	if slog.Default().Enabled(ctx, slog.LevelDebug) {
		slog.DebugContext(ctx, "received OpenAI response", "model", s.model, "input", message, "response", json.RawMessage(response.RawJSON()))
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
			return nil, errors.New("failed to select one tool: model returned multiple tool calls")
		}
		call := output.AsFunctionCall()
		selected = &toolCall{name: call.Name, arguments: call.Arguments}
	}
	if selected == nil {
		slog.DebugContext(ctx, "OpenAI response selected no tool", "model", s.model, "input", message)
	}
	return selected, nil
}
