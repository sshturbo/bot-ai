package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"bot-ai/config"
	"bot-ai/models"
)

type GeminiService struct {
	client *genai.Client
	config *config.Config
	model  *genai.GenerativeModel
}

func NewGeminiService(cfg *config.Config) models.AIService {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.GeminiApiKey))
	if err != nil {
		panic(fmt.Sprintf("Erro ao criar cliente Gemini: %v", err))
	}

	model := client.GenerativeModel(cfg.GeminiModel)
	model.SetTemperature(float32(cfg.GeminiTemperature))
	model.SetTopK(int32(cfg.GeminiTopK))
	model.SetTopP(float32(cfg.GeminiTopP))
	model.SetMaxOutputTokens(int32(cfg.GeminiMaxOutputTokens))
	model.ResponseMIMEType = "text/plain"

	return &GeminiService{
		client: client,
		config: cfg,
		model:  model,
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
	ctx := context.Background()

	resp, err := s.model.GenerateContent(ctx, genai.Text(question))
	if err != nil {
		return "", fmt.Errorf("erro ao gerar conteúdo: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("resposta vazia do Gemini")
	}

	return fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0]), nil
}
