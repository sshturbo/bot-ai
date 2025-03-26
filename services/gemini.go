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

func (s *GeminiService) AskWithRetry(userID int64, question string) (string, string, error) {
	var lastErr error
	for attempt := 1; attempt <= s.config.MaxRetries; attempt++ {
		answer, hash, err := s.Ask(userID, question)
		if err == nil {
			return answer, hash, nil
		}

		lastErr = err
		if attempt < s.config.MaxRetries {
			time.Sleep(s.config.RetryDelay)
		}
	}
	return "", "", fmt.Errorf("todas as tentativas falharam: %v", lastErr)
}

func (s *GeminiService) Ask(userID int64, question string) (string, string, error) {
	ctx := context.Background()

	// Busca o chat ativo do usuário
	chat, err := s.db.GetActiveChat(userID)
	if err != nil {
		return "", "", fmt.Errorf("erro ao buscar chat ativo: %w", err)
	}

	// Se não houver nenhum chat, cria um novo
	if chat == nil {
		chat, err = s.db.CreateNewChat(userID)
		if err != nil {
			return "", "", fmt.Errorf("erro ao criar novo chat: %w", err)
		}
	}

	// Recupera o histórico de mensagens
	messages, err := s.db.GetChatMessages(chat.ID)
	if err != nil {
		return "", "", fmt.Errorf("erro ao recuperar histórico: %w", err)
	}

	// Prepara o histórico para o Gemini
	cs := s.model.StartChat()
	for _, msg := range messages {
		// Mapeia os roles do nosso sistema para os roles aceitos pelo Gemini
		role := "user"
		if msg.Role == "assistant" {
			role = "model"
		}

		cs.History = append(cs.History, &genai.Content{
			Role:  role,
			Parts: []genai.Part{genai.Text(msg.Content)},
		})
	}

	// Adiciona a nova pergunta ao histórico
	if err := s.db.AddMessageToChat(chat.ID, "user", question); err != nil {
		return "", "", fmt.Errorf("erro ao adicionar pergunta ao histórico: %w", err)
	}

	// Envia a pergunta para o Gemini
	resp, err := cs.SendMessage(ctx, genai.Text(question))
	if err != nil {
		return "", "", fmt.Errorf("erro ao obter resposta: %w", err)
	}

	answer := resp.Candidates[0].Content.Parts[0].(genai.Text)

	// Salva a resposta no banco
	hash, err := s.db.SaveMessage(string(answer))
	if err != nil {
		return "", "", fmt.Errorf("erro ao salvar resposta: %w", err)
	}

	// Adiciona a resposta ao histórico do chat
	if err := s.db.AddMessageToChatWithExistingHash(chat.ID, "assistant", string(answer), hash); err != nil {
		return "", "", fmt.Errorf("erro ao adicionar resposta ao histórico: %w", err)
	}

	return string(answer), hash, nil
}

func (s *GeminiService) NewChat(userID int64) error {
	if err := s.db.NewChat(userID); err != nil {
		return fmt.Errorf("erro ao criar novo chat: %w", err)
	}
	return nil
}
