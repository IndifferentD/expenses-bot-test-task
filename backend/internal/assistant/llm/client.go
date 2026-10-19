package llm

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

type Call struct {
	Name      string
	Arguments string
}

type Interpreter struct {
	client openai.Client
	model  string
}

func NewInterpreter(client openai.Client, model string) (*Interpreter, error) {
	if strings.TrimSpace(model) == "" {
		return nil, errors.New("OPENAI_MODEL is required")
	}
	return &Interpreter{client: client, model: model}, nil
}

func (s *Interpreter) Interpret(ctx context.Context, message string, sentAt time.Time) (*Call, error) {
	instructions := fmt.Sprintf(
		"Сегодня %s, часовой пояс %s. Относительные даты и периоды преобразуй в YYYY-MM-DD. Не придумывай сумму, дату или ID. Число после # означает ID траты.\n"+
			"add_expense: если сообщение содержит описание траты и сумму без явного другого действия, считай его просьбой добавить трату. Категорию определяй по покупке или услуге, при неясности выбирай other. В comment передавай только предмет покупки или услугу.\n"+
			"delete_expense: выбирай только при явной просьбе удалить.\n"+
			"edit_expense: выбирай при просьбе изменить существующую трату. В target передавай старые признаки, в changes — новые значения. latest=true только при явном указании на последнюю добавленную трату.\n"+
			"list_expenses: выбирай при просьбе показать траты. limit=0 означает все подходящие траты, в том числе если указан фильтр без ограничения числа; при запросе последних N передавай N, иначе null.\n"+
			"summarize_expenses: выбирай при вопросе о сумме или числе трат.\n"+
			"Для просмотра и сводки передавай все названные категории в categories, а если категорий нет — пустой массив; comment — только конкретный предмет или услугу. Границы периода включительны; конец текущего периода — сегодня.\n"+
			"Для удаления и поиска изменяемой траты извлекай только явно названные старые признаки: comment — короткая дословная часть описания; category — только если категория названа прямо.",
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
			return nil, fmt.Errorf("failed to request OpenAI response (code: %q, message: %q): %w", apiErr.Code, apiErr.Message, err)
		}
		return nil, fmt.Errorf("failed to request OpenAI response: %w", err)
	}
	slog.Debug("received OpenAI response", "model", s.model, "input", message, "response", json.RawMessage(response.RawJSON()))
	if response.Status != responses.ResponseStatusCompleted {
		return nil, fmt.Errorf("failed to complete OpenAI response: status %s", response.Status)
	}
	var selected *Call
	for _, output := range response.Output {
		if output.Type != "function_call" {
			continue
		}
		if selected != nil {
			return nil, errors.New("failed to select one tool: model returned multiple tool calls")
		}
		call := output.AsFunctionCall()
		selected = &Call{Name: call.Name, Arguments: call.Arguments}
	}
	if selected == nil {
		slog.Warn("could not select any tool", "model", s.model, "input", message, "output_text", response.OutputText())
	}
	return selected, nil
}
