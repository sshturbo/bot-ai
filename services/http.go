package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"bot-ai/config"
	"bot-ai/database"
)

type HTTPServer struct {
	config         *config.Config
	db             *database.Database
	authMiddleware *TelegramAuthMiddleware
}

func NewHTTPServer(cfg *config.Config, db *database.Database) *HTTPServer {
	return &HTTPServer{
		config:         cfg,
		db:             db,
		authMiddleware: NewTelegramAuthMiddleware(cfg.TelegramToken),
	}
}

// Middleware CORS para lidar com requisições cross-origin
func (s *HTTPServer) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Configura cabeçalhos CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Telegram-Init-Data")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 horas

		// Trata requisições OPTIONS (preflight)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func (s *HTTPServer) Start() {
	// Rotas da API
	http.HandleFunc("/api/messages/", s.corsMiddleware(s.handleGetMessage))
	http.HandleFunc("/api/messages", s.corsMiddleware(s.handleMessages))
	http.HandleFunc("/api/chat/new", s.corsMiddleware(s.handleNewChat))

	// Frontend static files handler
	http.HandleFunc("/", s.corsMiddleware(s.handleFrontend))

	addr := s.config.ServerAddr
	log.Printf("Servidor HTTP iniciado em http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func (s *HTTPServer) handleFrontend(w http.ResponseWriter, r *http.Request) {
	// Caminho para a pasta dist do frontend
	distPath := "./frontend/dist"

	// Se o caminho solicitado começa com /api, não processe
	if strings.HasPrefix(r.URL.Path, "/api/") {
		http.NotFound(w, r)
		return
	}

	// Tenta encontrar o arquivo no diretório dist
	filePath := filepath.Join(distPath, r.URL.Path)

	// Se o arquivo não existe, serve o index.html para suportar rotas do frontend
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		filePath = filepath.Join(distPath, "index.html")
	}

	http.ServeFile(w, r, filePath)
}

func (s *HTTPServer) handleGetMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodOptions {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	hash := strings.TrimPrefix(r.URL.Path, "/api/messages/")
	msg, err := s.db.GetMessage(hash)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			// Este é um caso comum quando a mensagem não existe
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "Mensagem não encontrada",
				"details": "A mensagem solicitada não existe ou foi removida",
			})
			return
		}

		// Para outros tipos de erros, registramos no log
		log.Printf("Erro ao buscar mensagem %s: %v", hash, err)
		http.Error(w, "Erro ao buscar mensagem", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msg)
}

func extractUserID(initData string) (int64, error) {
	// Decodifica o initData URL encoded
	data, err := url.QueryUnescape(initData)
	if err != nil {
		return 0, fmt.Errorf("erro ao decodificar initData: %w", err)
	}

	// Parse os parâmetros
	params, err := url.ParseQuery(data)
	if err != nil {
		return 0, fmt.Errorf("erro ao parsear initData: %w", err)
	}

	// Obtém o objeto user como string JSON
	userStr := params.Get("user")
	if userStr == "" {
		return 0, fmt.Errorf("user não encontrado no initData")
	}

	// Define uma struct para decodificar o JSON do usuário
	var user struct {
		ID int64 `json:"id"`
	}

	// Decodifica o JSON
	if err := json.Unmarshal([]byte(userStr), &user); err != nil {
		return 0, fmt.Errorf("erro ao decodificar JSON do usuário: %w", err)
	}

	if user.ID == 0 {
		return 0, fmt.Errorf("ID do usuário inválido")
	}

	return user.ID, nil
}

func (s *HTTPServer) handleMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Obtém o initData do cabeçalho para identificar o usuário
	initData := r.Header.Get("X-Telegram-Init-Data")
	if initData == "" {
		http.Error(w, "Unauthorized: Missing init data", http.StatusUnauthorized)
		return
	}

	// Valida o initData
	if !s.authMiddleware.ValidateInitData(initData) {
		http.Error(w, "Unauthorized: Invalid init data", http.StatusUnauthorized)
		return
	}

	// Extrai o user_id do initData
	userID, err := extractUserID(initData)
	if err != nil {
		log.Printf("Erro ao extrair user_id: %v", err)
		http.Error(w, "Erro ao identificar usuário", http.StatusBadRequest)
		return
	}

	// Busca mensagens do usuário
	messages, err := s.db.GetMessagesByUser(userID)
	if err != nil {
		log.Printf("Erro ao buscar mensagens do usuário %d: %v", userID, err)
		http.Error(w, "Erro ao buscar mensagens", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func (s *HTTPServer) handleNewChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Obtém o initData do cabeçalho para identificar o usuário
	initData := r.Header.Get("X-Telegram-Init-Data")
	if initData == "" {
		http.Error(w, "Unauthorized: Missing init data", http.StatusUnauthorized)
		return
	}

	// Valida o initData
	if !s.authMiddleware.ValidateInitData(initData) {
		http.Error(w, "Unauthorized: Invalid init data", http.StatusUnauthorized)
		return
	}

	// Extrai o user_id do initData
	userID, err := extractUserID(initData)
	if err != nil {
		log.Printf("Erro ao extrair user_id: %v", err)
		http.Error(w, "Erro ao identificar usuário", http.StatusBadRequest)
		return
	}

	// Cria novo chat para o usuário
	if err := s.db.NewChat(userID); err != nil {
		log.Printf("Erro ao criar novo chat para usuário %d: %v", userID, err)
		http.Error(w, "Erro ao criar novo chat", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Novo chat criado com sucesso",
	})
}
