package main

import (
	"log"

	"bot-ai/config"
	"bot-ai/database"
	"bot-ai/services"
)

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

	// Inicializar serviços
	geminiService := services.NewGeminiService(cfg)

	telegramService, err := services.NewTelegramService(cfg, db, geminiService)
	if err != nil {
		log.Fatal(err)
	}

	httpServer := services.NewHTTPServer(cfg, db)

	// Iniciar servidor HTTP em uma goroutine
	go httpServer.Start()

	// Iniciar o bot do Telegram
	telegramService.Start()
}
