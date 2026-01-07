package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"wtfsiw/internal/ai"
	"wtfsiw/internal/config"
)

// Session represents a chat session
type Session struct {
	ID        string           `json:"id"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	Title     string           `json:"title,omitempty"` // Auto-generated from first message
	Messages  []ai.ChatMessage `json:"messages"`
}

// New creates a new empty session
func New() *Session {
	return &Session{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages:  []ai.ChatMessage{},
	}
}

// AddMessage adds a message to the session and updates the timestamp
func (s *Session) AddMessage(msg ai.ChatMessage) {
	msg.Timestamp = time.Now()
	s.Messages = append(s.Messages, msg)
	s.UpdatedAt = time.Now()

	// Auto-generate title from first user message if not set
	if s.Title == "" && msg.Role == "user" {
		s.Title = truncateTitle(msg.Content, 50)
	}
}

// Save persists the session to disk
func (s *Session) Save() error {
	sessionsDir := config.GetSessionsDir()
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		return fmt.Errorf("failed to create sessions directory: %w", err)
	}

	filename := fmt.Sprintf("%s_%s.json",
		s.CreatedAt.Format("20060102_150405"),
		s.ID[:8])
	filepath := filepath.Join(sessionsDir, filename)

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// Load loads a session from disk by ID
func Load(id string) (*Session, error) {
	sessionsDir := config.GetSessionsDir()

	files, err := os.ReadDir(sessionsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read sessions directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		// Check if this file matches the ID
		if strings.Contains(file.Name(), id[:min(8, len(id))]) {
			filepath := filepath.Join(sessionsDir, file.Name())
			return loadFromFile(filepath)
		}
	}

	return nil, fmt.Errorf("session not found: %s", id)
}

// LoadLatest loads the most recent session
func LoadLatest() (*Session, error) {
	sessions, err := List()
	if err != nil {
		return nil, err
	}

	if len(sessions) == 0 {
		return nil, fmt.Errorf("no sessions found")
	}

	return Load(sessions[0].ID)
}

// List returns all sessions, sorted by most recent first
func List() ([]*Session, error) {
	sessionsDir := config.GetSessionsDir()

	if _, err := os.Stat(sessionsDir); os.IsNotExist(err) {
		return []*Session{}, nil
	}

	files, err := os.ReadDir(sessionsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read sessions directory: %w", err)
	}

	sessions := make([]*Session, 0)
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filepath := filepath.Join(sessionsDir, file.Name())
		session, err := loadFromFile(filepath)
		if err != nil {
			continue // Skip corrupted files
		}
		sessions = append(sessions, session)
	}

	// Sort by UpdatedAt descending (most recent first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
	})

	return sessions, nil
}

// Delete removes a session from disk
func Delete(id string) error {
	sessionsDir := config.GetSessionsDir()

	files, err := os.ReadDir(sessionsDir)
	if err != nil {
		return fmt.Errorf("failed to read sessions directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		if strings.Contains(file.Name(), id[:min(8, len(id))]) {
			filepath := filepath.Join(sessionsDir, file.Name())
			return os.Remove(filepath)
		}
	}

	return fmt.Errorf("session not found: %s", id)
}

// DeleteAll removes all sessions
func DeleteAll() error {
	sessionsDir := config.GetSessionsDir()
	return os.RemoveAll(sessionsDir)
}

// Helper functions

func loadFromFile(path string) (*Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

func truncateTitle(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	// Replace newlines with spaces
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")

	if len(s) <= maxLen {
		return s
	}

	// Find last space before maxLen
	s = s[:maxLen]
	if idx := strings.LastIndex(s, " "); idx > maxLen/2 {
		s = s[:idx]
	}
	return s + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
