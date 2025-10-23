package context

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewContextManager(t *testing.T) {
	tempDir := t.TempDir()
	cm := NewContextManager(tempDir, 100, 24*time.Hour)

	if cm == nil {
		t.Fatal("NewContextManager returned nil")
	}

	if cm.storagePath != tempDir {
		t.Errorf("Expected storage path %s, got %s", tempDir, cm.storagePath)
	}

	if cm.maxHistory != 100 {
		t.Errorf("Expected maxHistory 100, got %d", cm.maxHistory)
	}

	if cm.sessionExpiry != 24*time.Hour {
		t.Errorf("Expected sessionExpiry 24h, got %v", cm.sessionExpiry)
	}
}

func TestGetSession(t *testing.T) {
	tempDir := t.TempDir()
	cm := NewContextManager(tempDir, 100, 24*time.Hour)

	sessionID := "test-session-1"
	session := cm.GetSession(sessionID)

	if session == nil {
		t.Fatal("GetSession returned nil")
	}

	if session.ID != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, session.ID)
	}

	if len(session.Messages) != 0 {
		t.Errorf("Expected empty messages, got %d messages", len(session.Messages))
	}
}

func TestAddUserMessage(t *testing.T) {
	tempDir := t.TempDir()
	cm := NewContextManager(tempDir, 100, 24*time.Hour)

	sessionID := "test-session-2"
	content := "Hello, can you help me?"
	intent := "greeting"

	err := cm.AddUserMessage(sessionID, content, intent)
	if err != nil {
		t.Fatalf("AddUserMessage failed: %v", err)
	}

	history := cm.GetHistory(sessionID, 0)
	if len(history) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(history))
	}

	msg := history[0]
	if msg.Role != "user" {
		t.Errorf("Expected role 'user', got %s", msg.Role)
	}

	if msg.Content != content {
		t.Errorf("Expected content %s, got %s", content, msg.Content)
	}

	if msg.Intent != intent {
		t.Errorf("Expected intent %s, got %s", intent, msg.Intent)
	}
}

func TestAddAssistantMessage(t *testing.T) {
	tempDir := t.TempDir()
	cm := NewContextManager(tempDir, 100, 24*time.Hour)

	sessionID := "test-session-3"
	content := "Sure, I can help you!"

	err := cm.AddAssistantMessage(sessionID, content)
	if err != nil {
		t.Fatalf("AddAssistantMessage failed: %v", err)
	}

	history := cm.GetHistory(sessionID, 0)
	if len(history) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(history))
	}

	msg := history[0]
	if msg.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got %s", msg.Role)
	}

	if msg.Content != content {
		t.Errorf("Expected content %s, got %s", content, msg.Content)
	}
}

func TestGetHistory(t *testing.T) {
	tempDir := t.TempDir()
	cm := NewContextManager(tempDir, 100, 24*time.Hour)

	sessionID := "test-session-4"

	// Add multiple messages
	cm.AddUserMessage(sessionID, "Message 1", "intent1")
	cm.AddAssistantMessage(sessionID, "Response 1")
	cm.AddUserMessage(sessionID, "Message 2", "intent2")
	cm.AddAssistantMessage(sessionID, "Response 2")

	// Get all history
	history := cm.GetHistory(sessionID, 0)
	if len(history) != 4 {
		t.Errorf("Expected 4 messages, got %d", len(history))
	}

	// Get limited history
	limitedHistory := cm.GetHistory(sessionID, 2)
	if len(limitedHistory) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(limitedHistory))
	}

	// Verify it gets the most recent messages
	if limitedHistory[0].Content != "Message 2" {
		t.Errorf("Expected first limited message to be 'Message 2', got %s", limitedHistory[0].Content)
	}
}

func TestContextData(t *testing.T) {
	tempDir := t.TempDir()
	cm := NewContextManager(tempDir, 100, 24*time.Hour)

	sessionID := "test-session-5"
	key := "user_preference"
	value := "dark_mode"

	// Set context data
	err := cm.SetContextData(sessionID, key, value)
	if err != nil {
		t.Fatalf("SetContextData failed: %v", err)
	}

	// Get context data
	retrievedValue, exists := cm.GetContextData(sessionID, key)
	if !exists {
		t.Fatal("Context data not found")
	}

	if retrievedValue != value {
		t.Errorf("Expected value %s, got %v", value, retrievedValue)
	}

	// Get non-existent key
	_, exists = cm.GetContextData(sessionID, "non_existent")
	if exists {
		t.Error("Expected non-existent key to return false")
	}
}

func TestClearSession(t *testing.T) {
	tempDir := t.TempDir()
	cm := NewContextManager(tempDir, 100, 24*time.Hour)

	sessionID := "test-session-6"

	// Add some data
	cm.AddUserMessage(sessionID, "Test message", "test")

	// Verify session exists
	history := cm.GetHistory(sessionID, 0)
	if len(history) != 1 {
		t.Fatal("Session should have 1 message")
	}

	// Clear session
	err := cm.ClearSession(sessionID)
	if err != nil {
		t.Fatalf("ClearSession failed: %v", err)
	}

	// Verify session is cleared
	history = cm.GetHistory(sessionID, 0)
	if len(history) != 0 {
		t.Errorf("Expected empty history after clear, got %d messages", len(history))
	}
}

func TestSessionPersistence(t *testing.T) {
	tempDir := t.TempDir()
	sessionID := "test-session-7"

	// Create first context manager and add data
	cm1 := NewContextManager(tempDir, 100, 24*time.Hour)
	cm1.AddUserMessage(sessionID, "Persistent message", "test")
	cm1.AddAssistantMessage(sessionID, "Persistent response")

	// Create second context manager (simulating app restart)
	cm2 := NewContextManager(tempDir, 100, 24*time.Hour)

	// Verify data persisted
	history := cm2.GetHistory(sessionID, 0)
	if len(history) != 2 {
		t.Fatalf("Expected 2 persisted messages, got %d", len(history))
	}

	if history[0].Content != "Persistent message" {
		t.Errorf("Expected persisted message, got %s", history[0].Content)
	}
}

func TestMaxHistoryLimit(t *testing.T) {
	tempDir := t.TempDir()
	maxHistory := 5
	cm := NewContextManager(tempDir, maxHistory, 24*time.Hour)

	sessionID := "test-session-8"

	// Add more messages than the limit
	for i := 0; i < 10; i++ {
		cm.AddUserMessage(sessionID, "Message", "test")
	}

	history := cm.GetHistory(sessionID, 0)
	if len(history) != maxHistory {
		t.Errorf("Expected %d messages (max history), got %d", maxHistory, len(history))
	}
}

func TestCleanupExpiredSessions(t *testing.T) {
	tempDir := t.TempDir()
	shortExpiry := 100 * time.Millisecond
	cm := NewContextManager(tempDir, 100, shortExpiry)

	sessionID := "test-session-9"

	// Add a message
	cm.AddUserMessage(sessionID, "Test message", "test")

	// Wait for session to expire
	time.Sleep(200 * time.Millisecond)

	// Run cleanup
	err := cm.CleanupExpiredSessions()
	if err != nil {
		t.Fatalf("CleanupExpiredSessions failed: %v", err)
	}

	// Verify session was removed
	cm.mu.RLock()
	_, exists := cm.sessions[sessionID]
	cm.mu.RUnlock()

	if exists {
		t.Error("Expected expired session to be removed")
	}

	// Verify storage file was removed
	sessionPath := filepath.Join(tempDir, sessionID+".json")
	if _, err := os.Stat(sessionPath); !os.IsNotExist(err) {
		t.Error("Expected session file to be removed")
	}
}

func TestAddInteraction(t *testing.T) {
	tempDir := t.TempDir()
	cm := NewContextManager(tempDir, 100, 24*time.Hour)

	sessionID := "test-session-10"
	userInput := "What's the weather?"
	intent := "weather_query"
	assistantResponse := "It's sunny today!"

	err := cm.AddInteraction(sessionID, userInput, intent, assistantResponse)
	if err != nil {
		t.Fatalf("AddInteraction failed: %v", err)
	}

	history := cm.GetHistory(sessionID, 0)
	if len(history) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(history))
	}

	if history[0].Role != "user" || history[0].Content != userInput {
		t.Error("First message should be user message")
	}

	if history[1].Role != "assistant" || history[1].Content != assistantResponse {
		t.Error("Second message should be assistant message")
	}
}

func TestGetSessionInfo(t *testing.T) {
	tempDir := t.TempDir()
	cm := NewContextManager(tempDir, 100, 24*time.Hour)

	sessionID := "test-session-11"

	// Non-existent session
	info := cm.GetSessionInfo(sessionID)
	if info["exists"].(bool) {
		t.Error("Expected session to not exist")
	}

	// Create session
	cm.AddUserMessage(sessionID, "Test", "test")

	// Get session info
	info = cm.GetSessionInfo(sessionID)
	if !info["exists"].(bool) {
		t.Error("Expected session to exist")
	}

	if info["id"].(string) != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, info["id"].(string))
	}

	if info["message_count"].(int) != 1 {
		t.Errorf("Expected message count 1, got %d", info["message_count"].(int))
	}
}

func TestBuildLLMContext(t *testing.T) {
	tempDir := t.TempDir()
	cm := NewContextManager(tempDir, 100, 24*time.Hour)

	sessionID := "test-session-12"

	// Add some conversation
	cm.AddUserMessage(sessionID, "Hello", "greeting")
	cm.AddAssistantMessage(sessionID, "Hi there!")
	cm.AddUserMessage(sessionID, "How are you?", "question")

	// Build LLM context
	llmContext := cm.BuildLLMContext(sessionID, 0)

	if len(llmContext) != 3 {
		t.Fatalf("Expected 3 messages in LLM context, got %d", len(llmContext))
	}

	if llmContext[0]["role"] != "user" || llmContext[0]["content"] != "Hello" {
		t.Error("First message should be user greeting")
	}

	// Test with limit
	limitedContext := cm.BuildLLMContext(sessionID, 2)
	if len(limitedContext) != 2 {
		t.Fatalf("Expected 2 messages in limited context, got %d", len(limitedContext))
	}
}

func TestGetSessionSummary(t *testing.T) {
	tempDir := t.TempDir()
	cm := NewContextManager(tempDir, 100, 24*time.Hour)

	sessionID := "test-session-13"

	// Empty session
	summary := cm.GetSessionSummary(sessionID)
	if summary != "" {
		t.Error("Expected empty summary for non-existent session")
	}

	// Add messages
	cm.AddUserMessage(sessionID, "Test message 1", "test")
	cm.AddAssistantMessage(sessionID, "Response 1")

	summary = cm.GetSessionSummary(sessionID)
	if summary == "" {
		t.Error("Expected non-empty summary")
	}

	// Summary should contain both messages
	if !contains(summary, "Test message 1") || !contains(summary, "Response 1") {
		t.Error("Summary should contain both messages")
	}
}

func TestClearAllSessions(t *testing.T) {
	tempDir := t.TempDir()
	cm := NewContextManager(tempDir, 100, 24*time.Hour)

	// Create multiple sessions
	cm.AddUserMessage("session-1", "Message 1", "test")
	cm.AddUserMessage("session-2", "Message 2", "test")
	cm.AddUserMessage("session-3", "Message 3", "test")

	// Verify sessions exist
	if len(cm.sessions) != 3 {
		t.Fatalf("Expected 3 sessions, got %d", len(cm.sessions))
	}

	// Clear all sessions
	err := cm.ClearAllSessions()
	if err != nil {
		t.Fatalf("ClearAllSessions failed: %v", err)
	}

	// Verify sessions are cleared
	if len(cm.sessions) != 0 {
		t.Errorf("Expected 0 sessions after clear, got %d", len(cm.sessions))
	}

	// Verify storage files are cleared
	files, _ := filepath.Glob(filepath.Join(tempDir, "*.json"))
	if len(files) != 0 {
		t.Errorf("Expected 0 storage files, got %d", len(files))
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
