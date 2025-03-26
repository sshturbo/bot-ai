package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"bot-ai/config"
	"bot-ai/database"
	"bot-ai/models"
)

type GeminiService struct {
	client *genai.Client
	config *config.Config
	model  *genai.GenerativeModel
	db     *database.Database
}

func NewGeminiService(cfg *config.Config, db *database.Database) models.AIService {
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
		db:     db,
	}
}

func (s *GeminiService) AskWithRetry(userID int64, question string) (string, error) {
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

func (s *GeminiService) Ask(userID int64, question string) (string, error) {
	ctx := context.Background()

	// Busca o chat ativo do usuário
	chat, err := s.db.GetActiveChat(userID)
	if err != nil {
		return "", fmt.Errorf("erro ao buscar chat ativo: %w", err)
	}

	// Se não houver chat ativo, cria um novo
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

	// Prepara o histórico para o Gemini
	cs := s.model.StartChat()
	for _, msg := range messages {
		cs.History = append(cs.History, &genai.Content{
			Parts: []genai.Part{
				genai.Text(msg.Content),
			},
			Role: msg.Role,
		})
	}

	// Envia a pergunta
	resp, err := cs.SendMessage(ctx, genai.Text(question))
	if err != nil {
		return "", fmt.Errorf("erro ao gerar conteúdo: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("resposta vazia do Gemini")
	}

	answer := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	// Salva a pergunta e a resposta no histórico
	if err := s.db.AddMessageToChat(chat.ID, "user", question); err != nil {
		return "", fmt.Errorf("erro ao salvar pergunta no histórico: %w", err)
	}
	if err := s.db.AddMessageToChat(chat.ID, "model", answer); err != nil {
		return "", fmt.Errorf("erro ao salvar resposta no histórico: %w", err)
	}

	return answer, nil
}

func (s *GeminiService) NewChat(userID int64) error {
	_, err := s.db.CreateNewChat(userID)
	if err != nil {
		return fmt.Errorf("erro ao criar novo chat: %w", err)
	}
	return nil
}
