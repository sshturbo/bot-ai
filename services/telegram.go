package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/time/rate"

	"bot-ai/config"
	"bot-ai/database"
	"bot-ai/models"
)

type Update struct {
	UpdateID int                     `json:"update_id"`
	Message  *models.TelegramMessage `json:"message"`
}

type WebAppInfo struct {
	URL string `json:"url"`
}

type InlineKeyboardButton struct {
	Text   string      `json:"text"`
	URL    string      `json:"url,omitempty"`
	WebApp *WebAppInfo `json:"web_app,omitempty"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type SendMessageRequest struct {
	ChatID           int64                `json:"chat_id"`
	Text             string               `json:"text"`
	ParseMode        string               `json:"parse_mode,omitempty"`
	ReplyToMessageID int                  `json:"reply_to_message_id,omitempty"`
	ReplyMarkup      InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

type TelegramResponse struct {
	OK          bool            `json:"ok"`
	Result      json.RawMessage `json:"result"`
	Description string          `json:"description,omitempty"`
}

type TelegramService struct {
	baseURL string
	token   string
	client  *http.Client
	limiter *rate.Limiter
	config  *config.Config
	db      *database.Database
	ai      models.AIService
	botInfo *models.TelegramUser
}

func NewTelegramService(cfg *config.Config, db *database.Database, ai models.AIService) (*TelegramService, error) {
	client := &http.Client{
		Timeout: time.Second * 60,
	}

	service := &TelegramService{
		baseURL: fmt.Sprintf("https://api.telegram.org/bot%s", cfg.TelegramToken),
		token:   cfg.TelegramToken,
		client:  client,
		limiter: rate.NewLimiter(rate.Every(time.Second), 30),
		config:  cfg,
		db:      db,
		ai:      ai,
	}

	// Obtém informações do bot
	botInfo, err := service.getMe()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter informações do bot: %w", err)
	}
	service.botInfo = botInfo

	return service, nil
}

func (s *TelegramService) getMe() (*models.TelegramUser, error) {
	resp, err := s.makeRequest("getMe", nil)
	if err != nil {
		return nil, err
	}

	var user models.TelegramUser
	if err := json.Unmarshal(resp.Result, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *TelegramService) makeRequest(method string, payload interface{}) (*TelegramResponse, error) {
	var body []byte
	var err error

	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
	}

	url := fmt.Sprintf("%s/%s", s.baseURL, method)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tgResp TelegramResponse
	if err := json.Unmarshal(respBody, &tgResp); err != nil {
		return nil, err
	}

	if !tgResp.OK {
		return nil, fmt.Errorf("erro na API do Telegram: %s", tgResp.Description)
	}

	return &tgResp, nil
}

func (s *TelegramService) Start() {
	log.Printf("Bot iniciado: @%s", s.botInfo.UserName)

	offset := 0
	for {
		updates, err := s.getUpdates(offset)
		if err != nil {
			log.Printf("Erro ao obter atualizações: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, update := range updates {
			go s.handleUpdate(update)
			offset = update.UpdateID + 1
		}
	}
}

func (s *TelegramService) getUpdates(offset int) ([]Update, error) {
	payload := map[string]interface{}{
		"offset":  offset,
		"timeout": 60,
	}

	resp, err := s.makeRequest("getUpdates", payload)
	if err != nil {
		return nil, err
	}

	var updates []Update
	if err := json.Unmarshal(resp.Result, &updates); err != nil {
		return nil, err
	}

	return updates, nil
}

func (s *TelegramService) escapeMarkdown(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(text)
}

func (s *TelegramService) formatPreview(answer string, maxLength int) string {
	if len(answer) <= maxLength {
		return answer
	}

	lastSpace := strings.LastIndex(answer[:maxLength], " ")
	if lastSpace == -1 {
		lastSpace = maxLength
	}

	return answer[:lastSpace] + "..."
}

func (s *TelegramService) sendResponse(msg *models.TelegramMessage, answer string) {
	userName := msg.From.UserName
	if userName == "" {
		userName = msg.From.FirstName
	}

	hash, err := s.db.SaveMessage(answer)
	if err != nil {
		log.Printf("Erro ao salvar mensagem: %v", err)
		s.sendErrorMessage(msg)
		return
	}

	preview := s.formatPreview(answer, 200)
	escapedUserName := s.escapeMarkdown(userName)
	escapedPreview := s.escapeMarkdown(preview)
	response := fmt.Sprintf("Resposta para %s:\n\n%s", escapedUserName, escapedPreview)
	webAppURL := fmt.Sprintf("%s/message/%s", s.config.WebAppURL, hash)

	var button InlineKeyboardButton
	if msg.Chat.Type == "private" {
		// Em chats privados, usa o WebApp
		button = InlineKeyboardButton{
			Text: "📝 Ver Resposta Completa",
			WebApp: &WebAppInfo{
				URL: webAppURL,
			},
		}
	} else {
		// Em grupos e canais, usa URL normal
		button = InlineKeyboardButton{
			Text: "📝 Ver Resposta Completa",
			URL:  webAppURL,
		}
	}

	keyboard := InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{button},
		},
	}

	payload := SendMessageRequest{
		ChatID:           msg.Chat.ID,
		Text:             response,
		ParseMode:        "MarkdownV2",
		ReplyToMessageID: msg.MessageID,
		ReplyMarkup:      keyboard,
	}

	_, err = s.makeRequest("sendMessage", payload)
	if err != nil {
		log.Printf("Erro ao enviar mensagem: %v", err)
		s.sendErrorMessage(msg)
		return
	}
}

// keepTypingStatus mantém o status de "digitando" ativo até que o canal done seja fechado
func (s *TelegramService) keepTypingStatus(chatID int64, done chan struct{}) {
	ticker := time.NewTicker(4 * time.Second) // Telegram requer atualização a cada 5s, usamos 4s para garantir
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			s.sendChatAction(chatID, "typing")
		}
	}
}

func (s *TelegramService) handleUpdate(update Update) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recuperado de panic ao processar mensagem: %v", r)
		}
	}()

	if update.Message == nil {
		return
	}

	if !s.shouldProcessMessage(update.Message) {
		return
	}

	// Processa comando /newchat
	if update.Message.Text == "/newchat" {
		err := s.ai.NewChat(update.Message.From.ID)
		if err != nil {
			log.Printf("Erro ao criar novo chat: %v", err)
			s.sendErrorMessage(update.Message)
			return
		}

		s.sendResponse(update.Message, "✨ Novo chat iniciado! Pode começar a conversar.")
		return
	}

	question := s.extractQuestion(update.Message)
	if question == "" {
		return
	}

	// Criar canal para controlar o status de digitação
	typingDone := make(chan struct{})

	// Iniciar goroutine para manter o status de digitação
	go s.keepTypingStatus(update.Message.Chat.ID, typingDone)

	// Enviar ação de "digitando" inicial
	s.sendChatAction(update.Message.Chat.ID, "typing")

	// Obter resposta da IA, agora passando o ID do usuário
	answer, err := s.ai.AskWithRetry(update.Message.From.ID, question)

	// Fechar o canal para parar o status de digitação
	close(typingDone)

	if err != nil {
		log.Printf("Erro ao obter resposta: %v", err)
		s.sendErrorMessage(update.Message)
		return
	}

	s.sendResponse(update.Message, answer)
}

func (s *TelegramService) sendChatAction(chatID int64, action string) {
	payload := map[string]interface{}{
		"chat_id": chatID,
		"action":  action,
	}

	_, err := s.makeRequest("sendChatAction", payload)
	if err != nil {
		log.Printf("Erro ao enviar ação de chat: %v", err)
	}
}

func (s *TelegramService) shouldProcessMessage(msg *models.TelegramMessage) bool {
	return msg.Chat.Type == "private" ||
		strings.Contains(msg.Text, "@"+s.botInfo.UserName)
}

func (s *TelegramService) extractQuestion(msg *models.TelegramMessage) string {
	question := msg.Text
	if msg.Chat.Type != "private" {
		question = strings.ReplaceAll(question, "@"+s.botInfo.UserName, "")
	}
	return strings.TrimSpace(question)
}

func (s *TelegramService) sendErrorMessage(msg *models.TelegramMessage) {
	payload := SendMessageRequest{
		ChatID:           msg.Chat.ID,
		Text:             "Desculpe, ocorreu um erro ao processar sua mensagem. Tente novamente mais tarde.",
		ReplyToMessageID: msg.MessageID,
	}

	s.makeRequest("sendMessage", payload)
}
