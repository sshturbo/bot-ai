package database

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"bot-ai/models"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir banco de dados: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("erro ao conectar ao banco de dados: %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func createTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			hash TEXT UNIQUE,
			content TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS chat_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS chat_messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			chat_history_id INTEGER,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (chat_history_id) REFERENCES chat_history(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_history_user_id ON chat_history(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_history_is_active ON chat_history(is_active)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_messages_chat_history_id ON chat_messages(chat_history_id)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("erro ao criar tabela/índice: %w", err)
		}
	}

	return nil
}

// SaveMessage salva uma mensagem normal
func (d *Database) SaveMessage(content string) (string, error) {
	hasher := sha256.New()
	hasher.Write([]byte(content + time.Now().String()))
	hash := hex.EncodeToString(hasher.Sum(nil))[:8]

	_, err := d.db.Exec(
		"INSERT INTO messages (hash, content) VALUES (?, ?)",
		hash, content,
	)
	if err != nil {
		return "", fmt.Errorf("erro ao salvar mensagem: %w", err)
	}

	return hash, nil
}

// GetMessage recupera uma mensagem pelo hash
func (d *Database) GetMessage(hash string) (*models.Message, error) {
	var msg models.Message
	err := d.db.QueryRow(
		"SELECT id, hash, content, created_at FROM messages WHERE hash = ?",
		hash,
	).Scan(&msg.ID, &msg.Hash, &msg.Content, &msg.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// GetActiveChat recupera o chat ativo de um usuário
func (d *Database) GetActiveChat(userID int64) (*models.ChatHistory, error) {
	var chat models.ChatHistory
	err := d.db.QueryRow(`
		SELECT id, user_id, is_active, created_at, updated_at 
		FROM chat_history 
		WHERE user_id = ? AND is_active = true
		ORDER BY created_at DESC LIMIT 1`,
		userID,
	).Scan(&chat.ID, &chat.UserID, &chat.IsActive, &chat.CreatedAt, &chat.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar chat ativo: %w", err)
	}
	return &chat, nil
}

// CreateNewChat cria um novo chat para o usuário e desativa os anteriores
func (d *Database) CreateNewChat(userID int64) (*models.ChatHistory, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback()

	// Desativa chats anteriores
	_, err = tx.Exec("UPDATE chat_history SET is_active = false WHERE user_id = ?", userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao desativar chats anteriores: %w", err)
	}

	// Cria novo chat
	result, err := tx.Exec(`
		INSERT INTO chat_history (user_id, is_active, created_at, updated_at) 
		VALUES (?, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar novo chat: %w", err)
	}

	chatID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter ID do novo chat: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("erro ao confirmar transação: %w", err)
	}

	return &models.ChatHistory{
		ID:        chatID,
		UserID:    userID,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// AddMessageToChat adiciona uma mensagem ao histórico do chat
func (d *Database) AddMessageToChat(chatID int64, role, content string) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback()

	// Primeiro salva na tabela messages e gera um hash
	hasher := sha256.New()
	hasher.Write([]byte(content + time.Now().String()))
	hash := hex.EncodeToString(hasher.Sum(nil))[:8]

	_, err = tx.Exec(
		"INSERT INTO messages (hash, content) VALUES (?, ?)",
		hash, content,
	)
	if err != nil {
		return fmt.Errorf("erro ao salvar mensagem: %w", err)
	}

	// Depois salva na tabela chat_messages
	_, err = tx.Exec(`
		INSERT INTO chat_messages (chat_history_id, role, content) 
		VALUES (?, ?, ?)`,
		chatID, role, content,
	)
	if err != nil {
		return fmt.Errorf("erro ao adicionar mensagem ao chat: %w", err)
	}

	// Atualiza o timestamp do chat
	_, err = tx.Exec("UPDATE chat_history SET updated_at = CURRENT_TIMESTAMP WHERE id = ?", chatID)
	if err != nil {
		return fmt.Errorf("erro ao atualizar timestamp do chat: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("erro ao confirmar transação: %w", err)
	}

	return nil
}

// GetChatMessages recupera todas as mensagens de um chat
func (d *Database) GetChatMessages(chatID int64) ([]models.ChatMessage, error) {
	rows, err := d.db.Query(`
		SELECT id, chat_history_id, role, content, created_at 
		FROM chat_messages 
		WHERE chat_history_id = ? 
		ORDER BY created_at ASC`,
		chatID,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar mensagens do chat: %w", err)
	}
	defer rows.Close()

	var messages []models.ChatMessage
	for rows.Next() {
		var msg models.ChatMessage
		err := rows.Scan(&msg.ID, &msg.ChatHistoryID, &msg.Role, &msg.Content, &msg.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler mensagem do chat: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// ListUserChats lista todos os chats de um usuário
func (d *Database) ListUserChats(userID int64) ([]models.ChatHistory, error) {
	rows, err := d.db.Query(`
		SELECT id, user_id, is_active, created_at, updated_at 
		FROM chat_history 
		WHERE user_id = ? 
		ORDER BY updated_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar chats do usuário: %w", err)
	}
	defer rows.Close()

	var chats []models.ChatHistory
	for rows.Next() {
		var chat models.ChatHistory
		err := rows.Scan(&chat.ID, &chat.UserID, &chat.IsActive, &chat.CreatedAt, &chat.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler chat: %w", err)
		}
		chats = append(chats, chat)
	}

	return chats, nil
}

// CleanupOldMessages remove mensagens mais antigas que o período especificado
func (d *Database) CleanupOldMessages(retentionPeriod time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-retentionPeriod)

	result, err := d.db.Exec(
		"DELETE FROM messages WHERE created_at < ?",
		cutoffTime,
	)
	if err != nil {
		return 0, fmt.Errorf("erro ao limpar mensagens antigas: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("erro ao obter contagem de mensagens removidas: %w", err)
	}

	return rowsAffected, nil
}

// ScheduleCleanup inicia uma rotina em segundo plano para limpar mensagens antigas periodicamente
func (d *Database) ScheduleCleanup(cleanupInterval, retentionPeriod time.Duration) {
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()

		for range ticker.C {
			count, err := d.CleanupOldMessages(retentionPeriod)
			if err != nil {
				log.Printf("Erro durante limpeza automática de mensagens: %v", err)
				continue
			}

			if count > 0 {
				log.Printf("Limpeza automática: removidas %d mensagens antigas", count)
			}
		}
	}()

	log.Printf("Limpeza automática de mensagens agendada: intervalo=%s, retenção=%s",
		cleanupInterval, retentionPeriod)
}

// GetMessagesByUser recupera todas as mensagens associadas a um usuário
func (d *Database) GetMessagesByUser(userID int64) ([]models.Message, error) {
	rows, err := d.db.Query(`
		SELECT DISTINCT m.id, m.hash, m.content, m.created_at 
		FROM chat_history ch
		JOIN chat_messages cm ON cm.chat_history_id = ch.id
		JOIN messages m ON m.content = cm.content
		WHERE ch.user_id = ?
		ORDER BY m.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar mensagens do usuário: %w", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(&msg.ID, &msg.Hash, &msg.Content, &msg.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler mensagem: %w", err)
		}
		messages = append(messages, msg)
	}

	// Se não encontrou mensagens, retorna uma lista vazia em vez de erro
	if len(messages) == 0 {
		return []models.Message{}, nil
	}

	return messages, nil
}

// NewChat cria um novo chat para o usuário
func (d *Database) NewChat(userID int64) error {
	_, err := d.CreateNewChat(userID)
	if err != nil {
		return fmt.Errorf("erro ao criar novo chat: %w", err)
	}
	return nil
}

func (d *Database) Close() error {
	return d.db.Close()
}
