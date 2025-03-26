package services

import (
	"log"
	"net/http"
)

type TelegramAuthMiddleware struct {
	botToken string
}

func NewTelegramAuthMiddleware(botToken string) *TelegramAuthMiddleware {
	return &TelegramAuthMiddleware{
		botToken: botToken,
	}
}

// Desabilitar temporariamente a validação para desenvolvimento
func (m *TelegramAuthMiddleware) ValidateInitData(initData string) bool {
	// Log para debug
	// log.Printf("Validando initData: %s", initData)

	// Para desenvolvimento, vamos aceitar todas as requisições
	// Atenção: REMOVER em produção!
	return true

	/* Implementação correta:
	// Divide a string em pares chave-valor mantendo o formato original
	params := make(map[string]string)
	pairs := strings.Split(initData, "&")
	hash := ""

	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		if key == "hash" {
			hash = value
			continue
		}
		params[key] = value
	}

	if hash == "" {
		log.Printf("Hash não encontrado no initData")
		return false
	}

	// Ordena as chaves
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Cria a string de verificação mantendo os valores exatamente como recebidos
	var lines []string
	for _, k := range keys {
		lines = append(lines, fmt.Sprintf("%s=%s", k, params[k]))
	}
	checkString := strings.Join(lines, "\n")
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
	*/
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
