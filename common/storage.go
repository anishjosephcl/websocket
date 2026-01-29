package common

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/tidwall/buntdb"
)

type FileClient struct {
	db     *buntdb.DB
	dbPath string
}

func NewFileClient(dbPath string) (*FileClient, error) {
	db, err := buntdb.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return &FileClient{
		db:     db,
		dbPath: dbPath,
	}, nil
}

func (fc *FileClient) WriteMessageToFile(message string) error {
	timestamp := fmt.Sprintf("%d", time.Now().UnixNano())

	return fc.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set("msg:"+timestamp, message, nil)
		if err != nil {
			return fmt.Errorf("failed to write message: %w", err)
		}
		fmt.Printf("Wrote message: %s\n", message)
		return nil
	})
}

func (fc *FileClient) RetrieveMessageFromFile(limit int) ([]string, error) {
	var messages []string

	err := fc.db.View(func(tx *buntdb.Tx) error {
		return tx.Descend("", func(key, value string) bool {
			if len(messages) >= limit {
				return false
			}
			messages = append(messages, value)
			return true
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve messages: %w", err)
	}

	// Reverse to chronological order within limit
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}
func (fc *FileClient) Close() error {
	if err := fc.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
}
func GetFilePath() string {
	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed to get current directory:", err)
	}

	// Get parent directory (-1 level up)
	parentDir := filepath.Dir(currentDir)

	// Join parent directory with messages.db
	return filepath.Join(parentDir, "messages.db")
}
