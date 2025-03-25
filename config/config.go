package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken string
	GeminiApiKey  string
	WebAppURL     string
	Debug         bool
	HTTPTimeout   time.Duration
	MaxRetries    int
	RetryDelay    time.Duration

	// Configurações de retenção de mensagens
	MessageRetention time.Duration
	CleanupInterval  time.Duration
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Erro ao carregar arquivo .env")
	}

	// Obtém o período de retenção de mensagens (padrão: 30 dias)
	retentionDays := getEnvAsInt("MESSAGE_RETENTION_DAYS", 30)
	messageRetention := time.Duration(retentionDays) * 24 * time.Hour

	// Obtém o intervalo de limpeza (padrão: 24 horas)
	cleanupHours := getEnvAsInt("CLEANUP_INTERVAL_HOURS", 24)
	cleanupInterval := time.Duration(cleanupHours) * time.Hour

	return &Config{
		TelegramToken:    os.Getenv("TELEGRAM_BOT_TOKEN"),
		GeminiApiKey:     os.Getenv("GEMINI_API_KEY"),
		WebAppURL:        os.Getenv("WEBAPP_URL"),
		HTTPTimeout:      30 * time.Second,
		MaxRetries:       3,
		RetryDelay:       2 * time.Second,
		MessageRetention: messageRetention,
		CleanupInterval:  cleanupInterval,
	}
}

// getEnvAsInt obtém uma variável de ambiente e a converte para inteiro,
// usando o valor padrão caso a variável não exista ou seja inválida
func getEnvAsInt(name string, defaultValue int) int {
	valueStr := os.Getenv(name)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Aviso: valor inválido para %s, usando padrão (%d): %v", name, defaultValue, err)
		return defaultValue
	}

	return value
}
