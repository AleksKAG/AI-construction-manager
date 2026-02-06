package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/ai/foundation_models/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

// LLMParser использует Yandex LLM для парсинга документов
type LLMParser struct {
	client *foundation_models.TextGenerationServiceClient
	folderID string
}

func NewLLMParser(iamToken, folderID string) (*LLMParser, error) {
	ctx := context.Background()
	sdk, err := ycsdk.Build(ctx, ycsdk.Config{
		Credentials: ycsdk.OAuthToken(iamToken),
	})
	if err != nil {
		return nil, err
	}
	return &LLMParser{
		client: sdk.FoundationModels().TextGeneration(),
		folderID: folderID,
	}, nil
}

// ParseDocument парсит текст документа через LLM
func (p *LLMParser) ParseDocument(text string, instructions string) (map[string]interface{}, error) {
	req := &foundation_models.TextGenerationRequest{
		ModelUri: fmt.Sprintf("gpt://%s/yandexgpt-lite", p.folderID),
		CompletionOptions: &foundation_models.CompletionOptions{
			Temperature: 0.6,
			MaxTokens: 2000,
		},
		Messages: []*foundation_models.Message{
			{Role: "system", Text: instructions},
			{Role: "user", Text: text},
		},
	}
	resp, err := p.client.Generate(context.Background(), req)
	if err != nil {
		return nil, err
	}
	// Парсинг ответа как JSON (предполагая, что LLM возвращает JSON)
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(resp.GetResult().GetAlternatives()[0].GetMessage().GetText()), &parsed); err != nil {
		return nil, err
	}
	return parsed, nil
}

// Пример использования: парсинг для рисков/стоимости
func ParseForRisks(text string) (map[string]float64, error) {
	// ...
	return nil, nil
}
