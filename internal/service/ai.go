package service

import (
 "bytes"
 "encoding/json"
 "errors"
 "net/http"
 "os"
)

type AIService struct {
 apiKey string
}

func NewAIService() *AIService {
 return &AIService{
  apiKey: os.Getenv("OPENAI_API_KEY"),
 }
}

type AIRequest struct {
 Prompt string `json:"prompt"`
}

type AIResponse struct {
 Result string `json:"result"`
}

func (s *AIService) Analyze(prompt string) (string, error) {
 if s.apiKey == "" {
  return "AI отключен (нет API ключа)", nil
 }

 body := map[string]interface{}{
  "model": "gpt-4o-mini",
  "messages": []map[string]string{
   {
    "role": "user",
    "content": prompt,
   },
  },
 }

 jsonData, _ := json.Marshal(body)

 req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
 if err != nil {
  return "", err
 }

 req.Header.Set("Authorization", "Bearer "+s.apiKey)
 req.Header.Set("Content-Type", "application/json")

 client := &http.Client{}
 resp, err := client.Do(req)
 if err != nil {
  return "", err
 }
 defer resp.Body.Close()

 var result map[string]interface{}
 json.NewDecoder(resp.Body).Decode(&result)

 choices := result["choices"].([]interface{})
 if len(choices) == 0 {
  return "", errors.New("no response")
 }

 msg := choices[0].(map[string]interface{})["message"].(map[string]interface{})
 content := msg["content"].(string)

 return content, nil
}
