package services

import (
	"encoding/json"
	"log"
	"net/http"
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
	// Desabilitando temporariamente o middleware para debug
	// http.HandleFunc("/api/messages/", s.authMiddleware.Middleware(s.handleGetMessage))
	http.HandleFunc("/api/messages/", s.corsMiddleware(s.handleGetMessage))

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
		log.Printf("Erro ao buscar mensagem %s: %v", hash, err)
		http.Error(w, "Mensagem não encontrada", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msg)
}
