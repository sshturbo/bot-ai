package models

import "time"

type Message struct {
	ID        int64     `json:"id"`
	Hash      string    `json:"hash"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
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
