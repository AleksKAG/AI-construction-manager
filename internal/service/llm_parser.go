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

	"github.com/sirupsen/logrus"
)

// LLMParser — сервис для работы с Yandex GPT через HTTP API
type LLMParser struct {
	apiKey     string
	folderID   string
	httpClient *http.Client
	logger     *logrus.Logger
}

type LLMRequest struct {
	ModelUri          string    `json:"modelUri"`
	CompletionOptions Options   `json:"completionOptions"`
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
	Name         string   `json:"name"`
	DurationDays int      `json:"duration_days"`
	Dependencies []string `json:"dependencies"`
	Resources    []struct {
		Type string `json:"type"`
		Name string `json:"name"`
		Qty  int    `json:"quantity"`
	} `json:"resources"`
}

func NewLLMParser(apiKey, folderID string, logger *logrus.Logger) *LLMParser {
	return &LLMParser{
		apiKey:   apiKey,
		folderID: folderID,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		logger: logger,
	}
}

func (p *LLMParser) ParseMeetingTranscript(ctx context.Context, text string) ([]ParsedTask, error) {
	prompt := fmt.Sprintf(`Проанализируй протокол совещания строительного проекта и извлеки задачи в формате JSON.

Верни строго JSON без пояснений.

Формат:
{
  "tasks": [
    {
      "name": "Название задачи",
      "duration_days": 1,
      "dependencies": [],
      "resources": [
        {"type": "labor", "name": "Рабочие", "quantity": 1}
      ]
    }
  ]
}

Текст:
%s`, text)

	reqBody := LLMRequest{
		ModelUri: fmt.Sprintf("gpt://%s/yandexgpt/latest", p.folderID),
		CompletionOptions: Options{
			Stream:      false,
			Temperature: 0.2, // уменьшаем креативность → больше стабильности JSON
			MaxTokens:   2000,
		},
		Messages: []Message{
			{
				Role: "system",
				Text: "Ты эксперт по строительному менеджменту. Возвращай только валидный JSON.",
			},
			{
				Role: "user",
				Text: prompt,
			},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		"https://llm.api.cloud.yandex.net/foundationModels/v1/completion",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Api-Key "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LLM API error %d: %s", resp.StatusCode, string(respBody))
	}

	var llmResp LLMResponse
	if err := json.Unmarshal(respBody, &llmResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(llmResp.Result.Alternatives) == 0 {
		return nil, errors.New("empty response from LLM")
	}

	rawText := llmResp.Result.Alternatives[0].Message.Text
	p.logger.Debugf("LLM raw response: %s", rawText)

	jsonStr, err := extractJSON(rawText)
	if err != nil {
		return nil, err
	}

	var result struct {
		Tasks []ParsedTask `json:"tasks"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("parse JSON: %w\nraw: %s", err, jsonStr)
	}

	return result.Tasks, nil
}

// extractJSON — аккуратно достаёт JSON из ответа модели
func extractJSON(text string) (string, error) {
	start := bytes.IndexByte([]byte(text), '{')
	end := bytes.LastIndexByte([]byte(text), '}')

	if start == -1 || end == -1 || start >= end {
		return "", fmt.Errorf("no valid JSON found in response: %s", text)
	}

	return text[start : end+1], nil
}

вт, 24 мар. 2026 г., 13:31 Квачев Александр Григорьевич <a.kvachev@r7-service.ru>:
# github.com/AleksKAG/ai-construction-manager/internal/service 
Ошибка: internal/service/llm_parser.go:13:2: "github.com/AleksKAG/ai-construction-manager/internal/domain" импортировано, но не используется 
github.com/AleksKAG/ai-construction-manager/internal/delivery/telegram 
github.com/ghodss/yaml 
Ошибка: internal/llm_parser.go:15:28: undefined: foundation_models.TextGenerationServiceClient 
Ошибка: internal/llm_parser.go:28:15: sdk.FoundationModels undefined (тип *ycsdk.SDK не имеет поля или метода FoundationModels)
Ошибка: internal/llm_parser.go:35:28: undefined: foundation_models.TextGenerationRequest 
Ошибка: internal/llm_parser.go:38:17: нельзя использовать 0.6 (нетипизированную константу с плавающей запятой) в качестве значения *wrapperspb.DoubleValue в литерале структуры 
Ошибка: internal/llm_parser.go:39:15: нельзя использовать 2000 (нетипизированную константу int) в качестве значения *wrapperspb.Int64Value в литерале структуры 
Ошибка: internal/llm_parser.go:42:21: неизвестное поле Text в литерале структуры типа "github.com/yandex-cloud/go-genproto/yandex/cloud/ai/foundation_models/v1".Сообщение 
Ошибка: internal/llm_parser.go:43:19: неизвестное поле Text в литерале структуры типа "github.com/yandex-cloud/go-genproto/yandex/cloud/ai/foundation_models/v1".Сообщение 
Ошибка: Процесс завершен с кодом выхода 1.
 
 
From: Александр К <kvachev.aleks@gmail.com> 
Sent: Tuesday, March 24, 2026 1:27 PM
To: Квачев Александр Григорьевич <a.kvachev@r7-service.ru>
Subject: Re:
 
**Проблема очень простая и понятная:**
 
Ты **не заменил** файл `internal/llm_parser.go` на ту версию, которую я давал в предыдущем сообщении.  
В репозитории до сих пор лежит **старый код** с protobuf (`foundation_models.TextGenerationServiceClient`, `wrapperspb`, `sdk.FoundationModels` и т.д.).
 
Go продолжает пытаться компилировать старую версию, поэтому все эти ошибки остаются.
 
---
 
### ✅ Решение: Замени файл `internal/llm_parser.go` **полностью**
 
**Удалите всё содержимое файла** и вставьте **этот код** (это финальная рабочая версия через прямой HTTP-запрос к YandexGPT):
 
```go
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
 
// LLMParser — сервис для работы с Yandex GPT через HTTP API
type LLMParser struct {
                apiKey     string
                folderID   string
                httpClient *http.Client
                logger     *logrus.Logger
}
 
type LLMRequest struct {
                ModelUri          string    `json:"modelUri"`
                CompletionOptions Options   `json:"completionOptions"`
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
 
                // Извлекаем JSON из ответа (модель иногда добавляет пояснения)
                start := bytes.IndexByte([]byte(jsonText), '{')
                end := bytes.LastIndexByte([]byte(jsonText), '}')
                if start == -1 || end == -1 {
                               return nil, fmt.Errorf("no JSON found in response: %s", jsonText)
                }
 
                var result struct {
                               Tasks []ParsedTask `json:"tasks"`
                }
                if err := json.Unmarshal([]byte(jsonText)[start:end+1], &result); err != nil {
                               return nil, fmt.Errorf("failed to parse JSON from LLM: %w", err)
                }
 
                return result.Tasks, nil
}