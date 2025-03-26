package models

import (
	"time"
)

// AIService interface comum para serviços de IA
type AIService interface {
	AskWithRetry(userID int64, question string) (string, string, error) // Retorna (resposta, hash, erro)
	NewChat(userID int64) error
}

// Message representa uma mensagem armazenada no banco de dados
type Message struct {
	ID        int64     `json:"id"`
	Hash      string    `json:"hash"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Role      string    `json:"role,omitempty"`
}

// ChatHistory representa o histórico de chat de um usuário
type ChatHistory struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	IsActive       bool      `json:"is_active"`
	PreviewMessage string    `json:"preview_message"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ChatMessage representa uma mensagem no histórico
type ChatMessage struct {
	ID            int64     `json:"id"`
	ChatHistoryID int64     `json:"chat_history_id"`
	Role          string    `json:"role"`
	Content       string    `json:"content"`
	CreatedAt     time.Time `json:"created_at"`
}

// TelegramMessage representa uma mensagem do Telegram
type TelegramMessage struct {
	MessageID int           `json:"message_id"`
	From      *TelegramUser `json:"from"`
	Chat      *TelegramChat `json:"chat"`
	Text      string        `json:"text"`
}

// TelegramUser representa um usuário do Telegram
type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	UserName  string `json:"username,omitempty"`
}

// TelegramChat representa um chat do Telegram
type TelegramChat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

// ChatHistoryResponse representa a resposta da API com o histórico
type ChatHistoryResponse struct {
	ID        int64         `json:"id"`
	IsActive  bool          `json:"is_active"`
	Messages  []ChatMessage `json:"messages"`
	CreatedAt time.Time     `json:"created_at"`
}

// GeminiRequest representa a estrutura correta para a API do Gemini
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiResponse representa a estrutura de resposta da API do Gemini
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}
