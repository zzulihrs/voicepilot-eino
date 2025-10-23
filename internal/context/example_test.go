package context_test

import (
	"fmt"
	"time"

	ctx "github.com/deca/voicepilot-eino/internal/context"
)

// Example_basicUsage demonstrates basic context manager usage
func Example_basicUsage() {
	// Create a context manager
	cm := ctx.NewContextManager("./temp/sessions", 100, 24*time.Hour)

	sessionID := "example-session-1"

	// Add user message
	cm.AddUserMessage(sessionID, "播放音乐", "play_music")

	// Add assistant response
	cm.AddAssistantMessage(sessionID, "好的，正在为您播放音乐")

	// Get conversation history
	history := cm.GetHistory(sessionID, 0)

	for _, msg := range history {
		fmt.Printf("%s: %s\n", msg.Role, msg.Content)
	}

	// Output:
	// user: 播放音乐
	// assistant: 好的，正在为您播放音乐
}

// Example_contextData demonstrates custom context data usage
func Example_contextData() {
	cm := ctx.NewContextManager("./temp/sessions", 100, 24*time.Hour)
	sessionID := "example-session-2"

	// Set custom context data
	cm.SetContextData(sessionID, "user_name", "张三")
	cm.SetContextData(sessionID, "last_song", "稻香")

	// Retrieve context data
	userName, _ := cm.GetContextData(sessionID, "user_name")
	lastSong, _ := cm.GetContextData(sessionID, "last_song")

	fmt.Printf("User: %v, Last Song: %v\n", userName, lastSong)

	// Output:
	// User: 张三, Last Song: 稻香
}

// Example_llmIntegration demonstrates LLM context building
func Example_llmIntegration() {
	cm := ctx.NewContextManager("./temp/sessions", 100, 24*time.Hour)
	sessionID := "example-session-3"

	// Simulate a multi-turn conversation
	cm.AddInteraction(sessionID, "你好", "greeting", "你好！有什么可以帮助你的吗？")
	cm.AddInteraction(sessionID, "播放音乐", "play_music", "好的，正在播放音乐")
	cm.AddInteraction(sessionID, "换一首", "next_song", "已切换到下一首")

	// Build context for LLM (last 4 messages = 2 interactions)
	llmContext := cm.BuildLLMContext(sessionID, 4)

	for _, msg := range llmContext {
		fmt.Printf("%s: %s\n", msg["role"], msg["content"])
	}

	// Output:
	// user: 播放音乐
	// assistant: 好的，正在播放音乐
	// user: 换一首
	// assistant: 已切换到下一首
}

// Example_sessionManagement demonstrates session management operations
func Example_sessionManagement() {
	cm := ctx.NewContextManager("./temp/sessions", 100, 24*time.Hour)

	// Create multiple sessions
	cm.AddUserMessage("session-1", "Message 1", "test")
	cm.AddUserMessage("session-2", "Message 2", "test")

	// Get session info
	info := cm.GetSessionInfo("session-1")
	fmt.Printf("Session exists: %v\n", info["exists"])
	fmt.Printf("Message count: %v\n", info["message_count"])

	// Clear specific session
	cm.ClearSession("session-1")

	// Check if session still exists
	info = cm.GetSessionInfo("session-1")
	fmt.Printf("Session exists after clear: %v\n", info["exists"])

	// Output:
	// Session exists: true
	// Message count: 1
	// Session exists after clear: false
}

// Example_persistence demonstrates data persistence across instances
func Example_persistence() {
	storagePath := "./temp/sessions"
	sessionID := "persistent-session"

	// First instance - write data
	cm1 := ctx.NewContextManager(storagePath, 100, 24*time.Hour)
	cm1.AddUserMessage(sessionID, "Hello", "greeting")

	// Second instance - read data (simulates app restart)
	cm2 := ctx.NewContextManager(storagePath, 100, 24*time.Hour)
	history := cm2.GetHistory(sessionID, 0)

	if len(history) > 0 {
		fmt.Printf("Persisted message: %s\n", history[0].Content)
	}

	// Output:
	// Persisted message: Hello
}
