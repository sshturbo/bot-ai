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
	query := `
		CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			hash TEXT UNIQUE,
			content TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("erro ao criar tabela: %w", err)
	}

	return nil
}

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

// CleanupOldMessages remove mensagens mais antigas que o período especificado
func (d *Database) CleanupOldMessages(retentionPeriod time.Duration) (int64, error) {
	// Calcula a data limite para manter mensagens
	cutoffTime := time.Now().Add(-retentionPeriod)

	// Executa a consulta para excluir mensagens antigas
	result, err := d.db.Exec(
		"DELETE FROM messages WHERE created_at < ?",
		cutoffTime,
	)
	if err != nil {
		return 0, fmt.Errorf("erro ao limpar mensagens antigas: %w", err)
	}

	// Retorna o número de linhas afetadas
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

func (d *Database) Close() error {
	return d.db.Close()
}
