package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/AleksKAG/ai-construction-manager/internal/domain"
	"github.com/sirupsen/logrus"
)

type LLMParser struct {
	apiKey     string
	folderID   string
	httpClient *http.Client
	logger     *logrus.Logger
}

type LLMRequest struct {
	ModelUri          string   `json:"modelUri"`
	CompletionOptions Options  `json:"completionOptions"`
	Messages          []Message `json:"messages"`
}

type Options struct {
	Stream      bool    `json:"stream"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"maxTokens"`
}

type Message struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type LLMResponse struct {
	Result struct {
		Alternatives []struct {
			Message struct {
				Text string `json:"text"`
			} `json:"message"`
		} `json:"alternatives"`
	} `json:"result"`
}

type ParsedTask struct {
	Name         string `json:"name"`
	DurationDays int    `json:"duration_days"`
	Dependencies []string `json:"dependencies"`
	Resources    []struct {
		Type string `json:"type"`
		Name string `json:"name"`
		Qty  int    `json:"quantity"`
	} `json:"resources"`
}

func NewLLMParser(apiKey, folderID string, logger *logrus.Logger) *LLMParser {
	return &LLMParser{
		apiKey:     apiKey,
		folderID:   folderID,
		httpClient: &http.Client{Timeout: 60 * time.Second},
		logger:     logger,
	}
}

func (p *LLMParser) ParseMeetingTranscript(ctx context.Context, text string) ([]ParsedTask, error) {
	prompt := fmt.Sprintf(`Проанализируй протокол совещания строительного проекта и извлеки задачи в формате JSON.

Пример ответа:
{
  "tasks": [
    {
      "name": "Монтаж опалубки",
      "duration_days": 3,
      "dependencies": ["Подготовка площадки"],
      "resources": [
        {"type": "labor", "name": "Монтажники", "quantity": 4},
        {"type": "equipment", "name": "Подъёмный кран", "quantity": 1}
      ]
    }
  ]
}

Текст протокола:
%s`, text)

	reqBody := LLMRequest{
		ModelUri: fmt.Sprintf("gpt://%s/yandexgpt/latest", p.folderID),
		CompletionOptions: Options{
			Stream:      false,
			Temperature: 0.3,
			MaxTokens:   2000,
		},
		Messages: []Message{
			{Role: "system", Text: "Ты — эксперт по строительному менеджменту. Извлекай задачи из текста протокола и возвращай только JSON."},
			{Role: "user", Text: prompt},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		"https://llm.api.cloud.yandex.net/foundationModels/v1/completion",
		bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Api-Key "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("LLM API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var llmResp LLMResponse
	if err := json.NewDecoder(resp.Body).Decode(&llmResp); err != nil {
		return nil, err
	}

	if len(llmResp.Result.Alternatives) == 0 {
		return nil, errors.New("empty response from YandexGPT")
	}

	jsonText := llmResp.Result.Alternatives[0].Message.Text
	p.logger.Debugf("LLM raw response: %s", jsonText)

	// Извлекаем JSON из ответа
	start := bytes.IndexByte([]byte(jsonText), '{')
	end := bytes.LastIndexByte([]byte(jsonText), '}')
	if start == -1 || end == -1 {
		return nil, fmt.Errorf("no JSON found in response")
	}

	var result struct {
		Tasks []ParsedTask `json:"tasks"`
	}
	if err := json.Unmarshal([]byte(jsonText)[start:end+1], &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from LLM: %w", err)
	}

	return result.Tasks, nil
}
