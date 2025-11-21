package data

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/connordoman/doman/internal/data/db"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed sql/schema/*.sql
var schemaFS embed.FS

var (
	database *sql.DB
	queries  *db.Queries
)

// Conversation represents a conversation record (re-exported from db package)
type Conversation = db.Conversation

// Message represents a message record (re-exported from db package)
type Message = db.Message

// InitDB initializes the SQLite database and runs migrations
func InitDB(dbPath string) error {
	// Ensure the directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	var err error
	database, err = sql.Open("sqlite3", dbPath+"?_foreign_keys=1")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := database.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Read and execute schema files
	schemaFiles, err := schemaFS.ReadDir("sql/schema")
	if err != nil {
		return fmt.Errorf("failed to read schema directory: %w", err)
	}

	for _, file := range schemaFiles {
		if filepath.Ext(file.Name()) == ".sql" {
			schemaSQL, err := schemaFS.ReadFile(fmt.Sprintf("sql/schema/%s", file.Name()))
			if err != nil {
				return fmt.Errorf("failed to read schema file %s: %w", file.Name(), err)
			}

			if _, err := database.Exec(string(schemaSQL)); err != nil {
				return fmt.Errorf("failed to execute schema file %s: %w", file.Name(), err)
			}
		}
	}

	// Initialize queries
	queries = db.New(database)

	return nil
}

// GetDB returns the database connection
func GetDB() *sql.DB {
	return database
}

// GetQueries returns the queries instance
func GetQueries() *db.Queries {
	return queries
}

// CloseDB closes the database connection
func CloseDB() error {
	if database != nil {
		return database.Close()
	}
	return nil
}

// CreateConversation creates a new conversation
func CreateConversation(id, model, service string) (*Conversation, error) {
	if queries == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	ctx := context.Background()
	conv, err := queries.CreateConversation(ctx, db.CreateConversationParams{
		ID:      id,
		Model:   model,
		Service: service,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	return &conv, nil
}

// GetConversation retrieves a conversation by ID
func GetConversation(id string) (*Conversation, error) {
	if queries == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	ctx := context.Background()
	conv, err := queries.GetConversation(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("conversation not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	return &conv, nil
}

// UpdateConversationTimestamp updates the updated_at timestamp
func UpdateConversationTimestamp(id string) error {
	if queries == nil {
		return fmt.Errorf("database not initialized")
	}

	ctx := context.Background()
	if err := queries.UpdateConversationTimestamp(ctx, id); err != nil {
		return fmt.Errorf("failed to update conversation timestamp: %w", err)
	}

	return nil
}

// CreateMessage creates a new message in a conversation
func CreateMessage(conversationID, role, content string) (*Message, error) {
	if queries == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	ctx := context.Background()
	msg, err := queries.CreateMessage(ctx, db.CreateMessageParams{
		ConversationID: conversationID,
		Role:           role,
		Content:        content,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return &msg, nil
}

// GetMessagesByConversationID retrieves all messages for a conversation
func GetMessagesByConversationID(conversationID string) ([]*Message, error) {
	if queries == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	ctx := context.Background()
	messages, err := queries.GetMessagesByConversationId(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	// Convert to slice of pointers
	result := make([]*Message, len(messages))
	for i := range messages {
		result[i] = &messages[i]
	}

	return result, nil
}

// ListConversations lists recent conversations
func ListConversations(limit, offset int) ([]*Conversation, error) {
	if queries == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	ctx := context.Background()
	conversations, err := queries.ListConversations(ctx, db.ListConversationsParams{
		Limit:  int64(limit),
		Offset: int64(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}

	// Convert to slice of pointers
	result := make([]*Conversation, len(conversations))
	for i := range conversations {
		result[i] = &conversations[i]
	}

	return result, nil
}
