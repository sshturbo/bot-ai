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
	"bot-ai/database"
	"bot-ai/models"
)

type AzureOpenAIService struct {
	client *http.Client
	config *config.Config
	db     *database.Database
}

type AzureRequest struct {
	Messages    []models.ChatMessage `json:"messages"`
	Model       string               `json:"model"`
	MaxTokens   int                  `json:"max_tokens,omitempty"`
	Temperature float64              `json:"temperature,omitempty"`
}

type AzureResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func NewAzureOpenAIService(cfg *config.Config, db *database.Database) models.AIService {
	return &AzureOpenAIService{
		client: &http.Client{
			Timeout: cfg.HTTPTimeout,
		},
		config: cfg,
		db:     db,
	}
}

func (s *AzureOpenAIService) AskWithRetry(userID int64, question string) (string, error) {
	var lastErr error
	for attempt := 1; attempt <= s.config.MaxRetries; attempt++ {
		answer, err := s.Ask(userID, question)
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

func (s *AzureOpenAIService) Ask(userID int64, question string) (string, error) {
	// Busca ou cria um chat ativo para o usuário
	chat, err := s.db.GetActiveChat(userID)
	if err != nil {
		return "", fmt.Errorf("erro ao buscar chat ativo: %w", err)
	}

	if chat == nil {
		chat, err = s.db.CreateNewChat(userID)
		if err != nil {
			return "", fmt.Errorf("erro ao criar novo chat: %w", err)
		}
	}

	// Recupera o histórico de mensagens
	messages, err := s.db.GetChatMessages(chat.ID)
	if err != nil {
		return "", fmt.Errorf("erro ao recuperar histórico: %w", err)
	}

	// Prepara a requisição com o histórico
	url := fmt.Sprintf("%s/chat/completions", s.config.AzureOpenAIEndpoint)

	reqMessages := []models.ChatMessage{
		{
			Role:    "system",
			Content: "You are a helpful assistant.",
		},
	}

	// Adiciona o histórico de mensagens
	reqMessages = append(reqMessages, messages...)

	// Adiciona a pergunta atual
	reqMessages = append(reqMessages, models.ChatMessage{
		Role:    "user",
		Content: question,
	})

	reqBody := AzureRequest{
		Messages:    reqMessages,
		Model:       s.config.AzureOpenAIModel,
		MaxTokens:   s.config.AzureOpenAIMaxTokens,
		Temperature: s.config.AzureOpenAITemperature,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("erro ao serializar requisição: %w", err)
	}

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.config.AzureOpenAIKey))

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

	var azureResp AzureResponse
	if err := json.Unmarshal(body, &azureResp); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	if len(azureResp.Choices) == 0 {
		return "", fmt.Errorf("resposta vazia do Azure OpenAI")
	}

	answer := azureResp.Choices[0].Message.Content

	// Salva a pergunta e a resposta no histórico
	if err := s.db.AddMessageToChat(chat.ID, "user", question); err != nil {
		return "", fmt.Errorf("erro ao salvar pergunta no histórico: %w", err)
	}
	if err := s.db.AddMessageToChat(chat.ID, "assistant", answer); err != nil {
		return "", fmt.Errorf("erro ao salvar resposta no histórico: %w", err)
	}

	return answer, nil
}

func (s *AzureOpenAIService) NewChat(userID int64) error {
	_, err := s.db.CreateNewChat(userID)
	if err != nil {
		return fmt.Errorf("erro ao criar novo chat: %w", err)
	}
	return nil
}
