package executor

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os/exec"
	"runtime"
	"strings"

	"github.com/deca/voicepilot-eino/internal/qiniu"
	"github.com/deca/voicepilot-eino/pkg/types"
)

// Executor executes tasks based on the task plan
type Executor struct {
	handlers    map[string]ActionHandler
	qiniuClient *qiniu.Client
}

// ActionHandler is a function that handles a specific action
type ActionHandler func(ctx context.Context, params map[string]interface{}) *types.ExecutionResult

// NewExecutor creates a new executor
func NewExecutor() *Executor {
	e := &Executor{
		handlers:    make(map[string]ActionHandler),
		qiniuClient: qiniu.NewClient(),
	}

	// Register action handlers
	e.RegisterHandler("open_app", e.handleOpenApp)
	e.RegisterHandler("play_music", e.handlePlayMusic)
	e.RegisterHandler("execute_command", e.handleExecuteCommand)
	e.RegisterHandler("generate_text", e.handleGenerateText)
	e.RegisterHandler("write_article", e.handleGenerateText) // Alias for generate_text
	e.RegisterHandler("clarify", e.handleClarify)
	e.RegisterHandler("error", e.handleError)

	return e
}

// RegisterHandler registers a handler for a specific action
func (e *Executor) RegisterHandler(action string, handler ActionHandler) {
	e.handlers[action] = handler
}

// Execute executes a task plan
func (e *Executor) Execute(ctx context.Context, plan *types.TaskPlan) *types.ExecutionResult {
	log.Printf("Executing task plan with %d steps", len(plan.Steps))

	var results []string
	for i, step := range plan.Steps {
		log.Printf("Executing step %d: %s", i+1, step.Action)

		handler, exists := e.handlers[step.Action]
		if !exists {
			return &types.ExecutionResult{
				Success: false,
				Error:   fmt.Sprintf("未知的操作类型：%s", step.Action),
			}
		}

		result := handler(ctx, step.Parameters)
		if !result.Success {
			return result
		}

		if result.Message != "" {
			results = append(results, result.Message)
		}
	}

	return &types.ExecutionResult{
		Success: true,
		Message: strings.Join(results, "\n"),
	}
}

// handleOpenApp opens an application
func (e *Executor) handleOpenApp(ctx context.Context, params map[string]interface{}) *types.ExecutionResult {
	appName, ok := params["name"].(string)
	if !ok {
		return &types.ExecutionResult{
			Success: false,
			Error:   "缺少应用程序名称参数",
		}
	}

	log.Printf("Opening application: %s", appName)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = exec.CommandContext(ctx, "open", "-a", appName)
	case "windows":
		cmd = exec.CommandContext(ctx, "start", appName)
	case "linux":
		cmd = exec.CommandContext(ctx, "xdg-open", appName)
	default:
		return &types.ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("不支持的操作系统：%s", runtime.GOOS),
		}
	}

	if err := cmd.Run(); err != nil {
		return &types.ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("打开应用程序失败：%v", err),
		}
	}

	return &types.ExecutionResult{
		Success: true,
		Message: fmt.Sprintf("已打开应用程序：%s", appName),
	}
}

// handlePlayMusic opens NetEase Cloud Music app and searches for the song
func (e *Executor) handlePlayMusic(ctx context.Context, params map[string]interface{}) *types.ExecutionResult {
	// Try multiple parameter names for compatibility
	song, ok := params["song"].(string)
	if !ok {
		song, ok = params["song_name"].(string)
	}
	if !ok {
		song, ok = params["name"].(string)
	}
	if !ok {
		return &types.ExecutionResult{
			Success: false,
			Error:   "缺少歌曲名称参数",
		}
	}

	log.Printf("Opening NetEase Music app and searching for: %s", song)

	// Open NetEase Music app and search (macOS only)
	if runtime.GOOS == "darwin" {
		err := e.searchInNeteaseApp(ctx, song)
		if err != nil {
			log.Printf("Failed to search in NetEase app: %v", err)
			return &types.ExecutionResult{
				Success: false,
				Error:   fmt.Sprintf("无法在网易云音乐中搜索：%v", err),
			}
		}

		return &types.ExecutionResult{
			Success: true,
			Message: fmt.Sprintf("已在网易云音乐中搜索：%s", song),
			Data:    song,
		}
	}

	// Fallback for other OS: open web search
	err := e.openNeteaseWebSearch(ctx, song)
	if err != nil {
		return &types.ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("无法打开网易云音乐搜索：%v", err),
		}
	}

	return &types.ExecutionResult{
		Success: true,
		Message: fmt.Sprintf("已为您打开网易云音乐搜索：%s", song),
		Data:    song,
	}
}

// searchInNeteaseApp opens NetEase Cloud Music app and searches for the song
func (e *Executor) searchInNeteaseApp(ctx context.Context, songName string) error {
	// AppleScript to:
	// 1. Open and activate NetEase Music app
	// 2. Open search with Cmd+F
	// 3. Use clipboard to input Chinese song name
	// 4. Press Enter to search
	// Note: We don't auto-play, user can choose which song to play
	script := fmt.Sprintf(`
		-- Open NetEase Music app
		tell application "NeteaseMusic"
			activate
		end tell

		delay 1.5

		-- Copy song name to clipboard (supports Chinese)
		set the clipboard to "%s"

		tell application "System Events"
			-- Open search with Cmd+F
			keystroke "f" using command down
			delay 0.5

			-- Paste song name from clipboard
			keystroke "v" using command down
			delay 0.5

			-- Press Enter to search
			keystroke return
		end tell
	`, songName)

	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("AppleScript execution failed: %v (output: %s)", err, string(output))
	}

	log.Printf("Successfully opened NetEase Music and searched for: %s", songName)
	return nil
}

// openNeteaseWebSearch opens NetEase Cloud Music web search page for the song
func (e *Executor) openNeteaseWebSearch(ctx context.Context, songName string) error {
	// Build NetEase Cloud Music search URL
	// Format: https://music.163.com/#/search/m/?s=ENCODED_SONG_NAME

	// URL encode the song name (properly handles Chinese characters)
	encodedSong := url.QueryEscape(songName)
	searchURL := fmt.Sprintf("https://music.163.com/#/search/m/?s=%s", encodedSong)

	log.Printf("Opening NetEase Music search for: %s (URL: %s)", songName, searchURL)

	// Open URL in default browser
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = exec.CommandContext(ctx, "open", searchURL)
	case "linux":
		cmd = exec.CommandContext(ctx, "xdg-open", searchURL)
	case "windows":
		cmd = exec.CommandContext(ctx, "cmd", "/c", "start", searchURL)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open browser: %w", err)
	}

	log.Printf("Successfully opened NetEase Music search for: %s", songName)
	return nil
}

// handleExecuteCommand executes a system command
func (e *Executor) handleExecuteCommand(ctx context.Context, params map[string]interface{}) *types.ExecutionResult {
	command, ok := params["command"].(string)
	if !ok {
		return &types.ExecutionResult{
			Success: false,
			Error:   "缺少命令参数",
		}
	}

	log.Printf("Executing command: %s", command)

	// Parse command and arguments
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return &types.ExecutionResult{
			Success: false,
			Error:   "命令为空",
		}
	}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &types.ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("命令执行失败：%v\n输出：%s", err, string(output)),
		}
	}

	return &types.ExecutionResult{
		Success: true,
		Message: fmt.Sprintf("命令执行成功"),
		Data:    string(output),
	}
}

// handleGenerateText generates text content using LLM
func (e *Executor) handleGenerateText(ctx context.Context, params map[string]interface{}) *types.ExecutionResult {
	topic, ok := params["topic"].(string)
	if !ok {
		// Try other parameter names
		if content, ok := params["content"].(string); ok {
			topic = content
		} else if subject, ok := params["subject"].(string); ok {
			topic = subject
		} else {
			return &types.ExecutionResult{
				Success: false,
				Error:   "缺少主题参数",
			}
		}
	}

	log.Printf("Generating text for topic: %s", topic)

	// Get additional parameters
	length := "适中"
	if l, ok := params["length"].(string); ok {
		length = l
	}

	contentType := "文章"
	if ct, ok := params["content_type"].(string); ok {
		contentType = ct
	}

	// Construct prompt for LLM
	systemPrompt := "你是一个专业的内容创作助手。请根据用户的要求生成高质量的文本内容。"
	userPrompt := fmt.Sprintf("请写一篇关于「%s」的%s，长度要求：%s。", topic, contentType, length)

	messages := []qiniu.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	// Call LLM to generate content
	generatedText, err := e.qiniuClient.ChatCompletion(ctx, messages)
	if err != nil {
		log.Printf("Failed to generate text: %v", err)
		return &types.ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("文本生成失败：%v", err),
		}
	}

	log.Printf("Generated text: %s", generatedText[:min(100, len(generatedText))])

	return &types.ExecutionResult{
		Success: true,
		Message: generatedText,
		Data:    generatedText,
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// handleClarify handles clarification requests
func (e *Executor) handleClarify(ctx context.Context, params map[string]interface{}) *types.ExecutionResult {
	message, ok := params["message"].(string)
	if !ok {
		message = "抱歉，我没有理解您的意思，能否请您再说一遍？"
	}

	return &types.ExecutionResult{
		Success: true,
		Message: message,
	}
}

// handleError handles error actions
func (e *Executor) handleError(ctx context.Context, params map[string]interface{}) *types.ExecutionResult {
	message, ok := params["message"].(string)
	if !ok {
		message = "执行过程中发生错误"
	}

	return &types.ExecutionResult{
		Success: false,
		Error:   message,
	}
}
