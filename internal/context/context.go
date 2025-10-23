package context

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Message represents a conversation message
type Message struct {
	Role      string    `json:"role"`      // "user" or "assistant"
	Content   string    `json:"content"`   // message content
	Timestamp time.Time `json:"timestamp"` // message timestamp
	Intent    string    `json:"intent,omitempty"`
}

// Session represents a conversation session
type Session struct {
	ID        string    `json:"id"`
	Messages  []Message `json:"messages"`
	Context   map[string]interface{} `json:"context,omitempty"` // Additional context data
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ContextManager manages conversation context and sessions
type ContextManager struct {
	sessions      map[string]*Session
	mu            sync.RWMutex
	storagePath   string
	maxHistory    int           // Maximum number of messages to keep per session
	sessionExpiry time.Duration // Session expiration time
}

// NewContextManager creates a new context manager
func NewContextManager(storagePath string, maxHistory int, sessionExpiry time.Duration) *ContextManager {
	cm := &ContextManager{
		sessions:      make(map[string]*Session),
		storagePath:   storagePath,
		maxHistory:    maxHistory,
		sessionExpiry: sessionExpiry,
	}

	// Ensure storage directory exists
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		fmt.Printf("Warning: failed to create storage directory: %v\n", err)
	}

	return cm
}

// GetSession retrieves a session by ID, creates new one if not exists
func (cm *ContextManager) GetSession(sessionID string) *Session {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	session, exists := cm.sessions[sessionID]
	if !exists {
		// Try to load from storage
		session = cm.loadSessionFromStorage(sessionID)
		if session == nil {
			// Create new session
			session = &Session{
				ID:        sessionID,
				Messages:  []Message{},
				Context:   make(map[string]interface{}),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}
		cm.sessions[sessionID] = session
	}

	return session
}

// AddUserMessage adds a user message to the session
func (cm *ContextManager) AddUserMessage(sessionID, content string, intent string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	session := cm.getOrCreateSession(sessionID)

	message := Message{
		Role:      "user",
		Content:   content,
		Timestamp: time.Now(),
		Intent:    intent,
	}

	session.Messages = append(session.Messages, message)
	session.UpdatedAt = time.Now()

	// Trim history if exceeds max
	cm.trimSessionHistory(session)

	// Persist to storage
	return cm.saveSessionToStorage(session)
}

// AddAssistantMessage adds an assistant message to the session
func (cm *ContextManager) AddAssistantMessage(sessionID, content string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	session := cm.getOrCreateSession(sessionID)

	message := Message{
		Role:      "assistant",
		Content:   content,
		Timestamp: time.Now(),
	}

	session.Messages = append(session.Messages, message)
	session.UpdatedAt = time.Now()

	// Trim history if exceeds max
	cm.trimSessionHistory(session)

	// Persist to storage
	return cm.saveSessionToStorage(session)
}

// GetHistory retrieves conversation history for a session
func (cm *ContextManager) GetHistory(sessionID string, limit int) []Message {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	session, exists := cm.sessions[sessionID]
	if !exists {
		// Try to load from storage
		session = cm.loadSessionFromStorage(sessionID)
		if session == nil {
			return []Message{}
		}
	}

	messages := session.Messages
	if limit > 0 && len(messages) > limit {
		return messages[len(messages)-limit:]
	}

	return messages
}

// GetContextData retrieves custom context data for a session
func (cm *ContextManager) GetContextData(sessionID string, key string) (interface{}, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	session, exists := cm.sessions[sessionID]
	if !exists {
		return nil, false
	}

	value, exists := session.Context[key]
	return value, exists
}

// SetContextData sets custom context data for a session
func (cm *ContextManager) SetContextData(sessionID string, key string, value interface{}) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	session := cm.getOrCreateSession(sessionID)
	session.Context[key] = value
	session.UpdatedAt = time.Now()

	return cm.saveSessionToStorage(session)
}

// ClearSession clears a specific session
func (cm *ContextManager) ClearSession(sessionID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.sessions, sessionID)

	// Remove from storage
	sessionPath := filepath.Join(cm.storagePath, fmt.Sprintf("%s.json", sessionID))
	if err := os.Remove(sessionPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove session file: %w", err)
	}

	return nil
}

// ClearAllSessions clears all sessions
func (cm *ContextManager) ClearAllSessions() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.sessions = make(map[string]*Session)

	// Clear storage directory
	files, err := filepath.Glob(filepath.Join(cm.storagePath, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to list session files: %w", err)
	}

	for _, file := range files {
		if err := os.Remove(file); err != nil {
			fmt.Printf("Warning: failed to remove session file %s: %v\n", file, err)
		}
	}

	return nil
}

// CleanupExpiredSessions removes sessions that haven't been updated within the expiry duration
func (cm *ContextManager) CleanupExpiredSessions() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	expiredSessions := []string{}

	for id, session := range cm.sessions {
		if now.Sub(session.UpdatedAt) > cm.sessionExpiry {
			expiredSessions = append(expiredSessions, id)
		}
	}

	// Remove expired sessions
	for _, id := range expiredSessions {
		delete(cm.sessions, id)
		sessionPath := filepath.Join(cm.storagePath, fmt.Sprintf("%s.json", id))
		if err := os.Remove(sessionPath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Warning: failed to remove expired session file %s: %v\n", sessionPath, err)
		}
	}

	fmt.Printf("Cleaned up %d expired sessions\n", len(expiredSessions))
	return nil
}

// GetSessionSummary generates a summary of the conversation for context
func (cm *ContextManager) GetSessionSummary(sessionID string) string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	session, exists := cm.sessions[sessionID]
	if !exists || len(session.Messages) == 0 {
		return ""
	}

	// Build a summary of recent conversation
	var summary string
	recentMessages := session.Messages
	if len(recentMessages) > 5 {
		recentMessages = recentMessages[len(recentMessages)-5:]
	}

	for _, msg := range recentMessages {
		summary += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
	}

	return summary
}

// BuildLLMContext builds context messages for LLM from session history
func (cm *ContextManager) BuildLLMContext(sessionID string, maxMessages int) []map[string]string {
	messages := cm.GetHistory(sessionID, maxMessages)
	llmMessages := make([]map[string]string, 0, len(messages))

	for _, msg := range messages {
		llmMessages = append(llmMessages, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	return llmMessages
}

// Internal helper methods

func (cm *ContextManager) getOrCreateSession(sessionID string) *Session {
	session, exists := cm.sessions[sessionID]
	if !exists {
		session = &Session{
			ID:        sessionID,
			Messages:  []Message{},
			Context:   make(map[string]interface{}),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		cm.sessions[sessionID] = session
	}
	return session
}

func (cm *ContextManager) trimSessionHistory(session *Session) {
	if cm.maxHistory > 0 && len(session.Messages) > cm.maxHistory {
		// Keep only the most recent messages
		session.Messages = session.Messages[len(session.Messages)-cm.maxHistory:]
	}
}

func (cm *ContextManager) saveSessionToStorage(session *Session) error {
	sessionPath := filepath.Join(cm.storagePath, fmt.Sprintf("%s.json", session.ID))

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := os.WriteFile(sessionPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

func (cm *ContextManager) loadSessionFromStorage(sessionID string) *Session {
	sessionPath := filepath.Join(cm.storagePath, fmt.Sprintf("%s.json", sessionID))

	data, err := os.ReadFile(sessionPath)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("Warning: failed to read session file: %v\n", err)
		}
		return nil
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		fmt.Printf("Warning: failed to unmarshal session: %v\n", err)
		return nil
	}

	return &session
}

// AddInteraction is a convenience method to add both user and assistant messages
func (cm *ContextManager) AddInteraction(sessionID string, userInput string, intent string, assistantResponse string) error {
	if err := cm.AddUserMessage(sessionID, userInput, intent); err != nil {
		return fmt.Errorf("failed to add user message: %w", err)
	}

	if err := cm.AddAssistantMessage(sessionID, assistantResponse); err != nil {
		return fmt.Errorf("failed to add assistant message: %w", err)
	}

	return nil
}

// GetSessionInfo returns basic information about a session
func (cm *ContextManager) GetSessionInfo(sessionID string) map[string]interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	session, exists := cm.sessions[sessionID]
	if !exists {
		return map[string]interface{}{
			"exists": false,
		}
	}

	return map[string]interface{}{
		"exists":       true,
		"id":           session.ID,
		"message_count": len(session.Messages),
		"created_at":   session.CreatedAt,
		"updated_at":   session.UpdatedAt,
	}
}
