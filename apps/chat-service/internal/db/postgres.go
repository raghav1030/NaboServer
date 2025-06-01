package db

import (
	"context"
	"database/sql"
	"time"

	chatv1 "github.com/raghav1030/NaboServer/libs/proto/gen/go/chat/v1"
	_ "github.com/lib/pq"
)

type PostgresManager struct {
	db *sql.DB
}

func NewPostgresManager(connStr string) (*PostgresManager, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return &PostgresManager{db: db}, nil
}

func (p *PostgresManager) SaveMessage(ctx context.Context, conversationID, senderID, text, msgType string, attachment *chatv1.FileAttachment) (string, error) {
	var messageID string
	err := p.db.QueryRowContext(ctx, `
		INSERT INTO messages (conversation_id, sender_id, text, type, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, conversationID, senderID, text, msgType, time.Now()).Scan(&messageID)
	if err != nil {
		return "", err
	}
	// Save attachment if present
	if attachment != nil {
		_, err = p.db.ExecContext(ctx, `
			INSERT INTO message_attachments (message_id, url, file_name, mime_type, size_bytes, thumbnail_url)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, messageID, attachment.Url, attachment.FileName, attachment.MimeType, attachment.SizeBytes, attachment.ThumbnailUrl)
		if err != nil {
			return messageID, err
		}
	}
	return messageID, nil
}
