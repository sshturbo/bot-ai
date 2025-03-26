package main

import (
	"log"
	"strings"

	"bot-ai/config"
	"bot-ai/database"
	"bot-ai/models"
	"bot-ai/services"
)

func initializeAIService(cfg *config.Config, db *database.Database) models.AIService {
	// Verifica qual serviço deve ser usado com base na configuração
	switch cfg.AIService {
	case "azure":
		if strings.TrimSpace(cfg.AzureOpenAIKey) == "" {
			log.Fatal("Serviço Azure OpenAI selecionado mas AZURE_OPENAI_API_KEY não está configurada")
		}
		log.Println("Usando Azure OpenAI como serviço de IA")
		return services.NewAzureOpenAIService(cfg, db)

	case "google":
		if strings.TrimSpace(cfg.GeminiApiKey) == "" {
			log.Fatal("Serviço Google Gemini selecionado mas GEMINI_API_KEY não está configurada")
		}
		log.Println("Usando Google Gemini como serviço de IA")
		return services.NewGeminiService(cfg, db)

	default:
		log.Fatalf("Serviço de IA '%s' não suportado. Use 'google' ou 'azure' na variável AI_SERVICE", cfg.AIService)
		return nil
	}
}

func main() {
	// Carregar configurações
	cfg := config.LoadConfig()

	// Inicializar banco de dados
	db, err := database.NewDatabase("messages.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Configurar limpeza automática de mensagens antigas
	db.ScheduleCleanup(cfg.CleanupInterval, cfg.MessageRetention)

	// Inicializar o serviço de IA apropriado
	aiService := initializeAIService(cfg, db)

	// Inicializar serviço do Telegram
	telegramService, err := services.NewTelegramService(cfg, db, aiService)
	if err != nil {
		log.Fatal(err)
	}

	// Inicializar servidor HTTP
	httpServer := services.NewHTTPServer(cfg, db)

	// Iniciar servidor HTTP em uma goroutine
	go httpServer.Start()

	// Iniciar o bot do Telegram
	telegramService.Start()
}
