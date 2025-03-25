package models

import "time"

// AIService interface comum para serviços de IA
type AIService interface {
	AskWithRetry(question string) (string, error)
}

// Message representa uma mensagem armazenada no banco de dados
type Message struct {
	ID        int64     `json:"id"`
	Hash      string    `json:"hash"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
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

// ChatMessage representa uma mensagem para APIs de chat (Gemini, Azure)
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
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
