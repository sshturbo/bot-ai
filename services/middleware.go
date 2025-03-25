package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

type TelegramAuthMiddleware struct {
	botToken string
}

func NewTelegramAuthMiddleware(botToken string) *TelegramAuthMiddleware {
	return &TelegramAuthMiddleware{
		botToken: botToken,
	}
}

func (m *TelegramAuthMiddleware) ValidateInitData(initData string) bool {
	// Log para debug
	log.Printf("Validando initData: %s", initData)

	// Decodifica a URL encoded string
	decodedData, err := url.QueryUnescape(initData)
	if err != nil {
		log.Printf("Erro ao decodificar initData: %v", err)
		return false
	}

	// Extrai os parâmetros
	data, err := url.ParseQuery(decodedData)
	if err != nil {
		log.Printf("Erro ao parsear initData: %v", err)
		return false
	}

	// Obtém e remove o hash dos parâmetros
	hash := data.Get("hash")
	data.Del("hash")
	if hash == "" {
		log.Printf("Hash não encontrado no initData")
		return false
	}

	// Ordena os parâmetros alfabeticamente
	params := make([]string, 0)
	for k, values := range data {
		for _, v := range values {
			params = append(params, fmt.Sprintf("%s=%s", k, v))
		}
	}
	sort.Strings(params)

	// Cria a string de verificação
	checkString := strings.Join(params, "\n")
	log.Printf("String de verificação: %s", checkString)

	// Calcula o HMAC-SHA256
	secret := sha256.Sum256([]byte(m.botToken))
	h := hmac.New(sha256.New, secret[:])
	h.Write([]byte(checkString))
	expectedHash := hex.EncodeToString(h.Sum(nil))

	// Log do resultado da validação
	log.Printf("Hash recebido: %s", hash)
	log.Printf("Hash esperado: %s", expectedHash)
	log.Printf("Validação: %v", hash == expectedHash)

	return hash == expectedHash
}

func (m *TelegramAuthMiddleware) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Configura CORS para desenvolvimento
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Telegram-Init-Data")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Obtém o initData do cabeçalho
		initData := r.Header.Get("X-Telegram-Init-Data")
		log.Printf("InitData recebido: %s", initData)

		if initData == "" {
			log.Printf("X-Telegram-Init-Data não encontrado nos headers")
			http.Error(w, "Unauthorized: Missing init data", http.StatusUnauthorized)
			return
		}

		// Valida o initData
		if !m.ValidateInitData(initData) {
			log.Printf("Validação do initData falhou")
			http.Error(w, "Unauthorized: Invalid init data", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}
