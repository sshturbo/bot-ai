package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken string
	GeminiApiKey  string
	WebAppURL     string
	Debug         bool
	HTTPTimeout   time.Duration
	ServerAddr    string
	MaxRetries    int
	RetryDelay    time.Duration

	// Configurações de retenção de mensagens
	MessageRetention time.Duration
	CleanupInterval  time.Duration

	// Seleção do serviço de IA
	AIService string

	// Configurações do Gemini
	GeminiModel           string
	GeminiTemperature     float64
	GeminiTopK            int
	GeminiTopP            float64
	GeminiMaxOutputTokens int

	// Azure OpenAI Configuration
	AzureOpenAIKey         string
	AzureOpenAIEndpoint    string
	AzureOpenAIModel       string
	AzureOpenAIMaxTokens   int
	AzureOpenAITemperature float64
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

	// Obtém o endereço do servidor (padrão: "localhost:8080")
	serverAddr := getEnvWithDefault("SERVER_ADDR", "localhost:8080")

	// Azure OpenAI configuration
	maxTokens := getEnvAsInt("AZURE_OPENAI_MAX_TOKENS", 4096)
	temperature := getEnvAsFloat("AZURE_OPENAI_TEMPERATURE", 1.0)

	return &Config{
		TelegramToken:    os.Getenv("TELEGRAM_BOT_TOKEN"),
		GeminiApiKey:     os.Getenv("GEMINI_API_KEY"),
		WebAppURL:        os.Getenv("WEBAPP_URL"),
		HTTPTimeout:      30 * time.Second,
		ServerAddr:       serverAddr,
		MaxRetries:       3,
		RetryDelay:       2 * time.Second,
		MessageRetention: messageRetention,
		CleanupInterval:  cleanupInterval,

		// Seleção do serviço de IA (padrão: google)
		AIService: strings.ToLower(getEnvWithDefault("AI_SERVICE", "google")),

		// Configurações do Gemini
		GeminiModel:           getEnvWithDefault("GEMINI_MODEL", "gemini-2.5-pro-exp-03-25"),
		GeminiTemperature:     getEnvAsFloat("GEMINI_TEMPERATURE", 1.0),
		GeminiTopK:            getEnvAsInt("GEMINI_TOP_K", 64),
		GeminiTopP:            getEnvAsFloat("GEMINI_TOP_P", 0.95),
		GeminiMaxOutputTokens: getEnvAsInt("GEMINI_MAX_OUTPUT_TOKENS", 65536),

		// Azure OpenAI settings
		AzureOpenAIKey:         os.Getenv("AZURE_OPENAI_API_KEY"),
		AzureOpenAIEndpoint:    getEnvWithDefault("AZURE_OPENAI_ENDPOINT", "https://models.inference.ai.azure.com"),
		AzureOpenAIModel:       getEnvWithDefault("AZURE_OPENAI_MODEL", "gpt-4"),
		AzureOpenAIMaxTokens:   maxTokens,
		AzureOpenAITemperature: temperature,
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

// getEnvAsFloat obtém uma variável de ambiente e a converte para float64,
// usando o valor padrão caso a variável não exista ou seja inválida
func getEnvAsFloat(name string, defaultValue float64) float64 {
	valueStr := os.Getenv(name)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		log.Printf("Aviso: valor inválido para %s, usando padrão (%.2f): %v", name, defaultValue, err)
		return defaultValue
	}

	return value
}

// getEnvWithDefault obtém uma variável de ambiente ou retorna o valor padrão
// caso a variável não exista
func getEnvWithDefault(name string, defaultValue string) string {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue
	}
	return value
}
