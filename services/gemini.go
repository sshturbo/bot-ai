package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"bot-ai/config"
	"bot-ai/models"
)

type GeminiService struct {
	client *http.Client
	config *config.Config
}

func NewGeminiService(cfg *config.Config) *GeminiService {
	return &GeminiService{
		client: &http.Client{
			Timeout: cfg.HTTPTimeout,
		},
		config: cfg,
	}
}

func (s *GeminiService) AskWithRetry(question string) (string, error) {
	var lastErr error
	for attempt := 1; attempt <= s.config.MaxRetries; attempt++ {
		answer, err := s.Ask(question)
		if err == nil {
			return answer, nil
		}

		lastErr = err
		if attempt < s.config.MaxRetries {
			time.Sleep(s.config.RetryDelay)
		}
	}
	return "", fmt.Errorf("todas as tentativas falharam: %v", lastErr)
}

func (s *GeminiService) Ask(question string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=%s", s.config.GeminiApiKey)

	reqBody := &models.GeminiRequest{
		Contents: []models.GeminiContent{
			{
				Parts: []models.GeminiPart{
					{
						Text: question,
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("erro ao serializar requisição: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.config.HTTPTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro na requisição HTTP: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao ler resposta: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API retornou status %d: %s", resp.StatusCode, string(body))
	}

	var geminiResp models.GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("resposta vazia do Gemini")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}
