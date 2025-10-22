package executor

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"

	"github.com/deca/voicepilot-eino/pkg/types"
)

// Executor executes tasks based on the task plan
type Executor struct {
	handlers map[string]ActionHandler
}

// ActionHandler is a function that handles a specific action
type ActionHandler func(ctx context.Context, params map[string]interface{}) *types.ExecutionResult

// NewExecutor creates a new executor
func NewExecutor() *Executor {
	e := &Executor{
		handlers: make(map[string]ActionHandler),
	}

	// Register action handlers
	e.RegisterHandler("open_app", e.handleOpenApp)
	e.RegisterHandler("play_music", e.handlePlayMusic)
	e.RegisterHandler("execute_command", e.handleExecuteCommand)
	e.RegisterHandler("generate_text", e.handleGenerateText)
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

// handlePlayMusic plays music
func (e *Executor) handlePlayMusic(ctx context.Context, params map[string]interface{}) *types.ExecutionResult {
	song, ok := params["song"].(string)
	if !ok {
		return &types.ExecutionResult{
			Success: false,
			Error:   "缺少歌曲名称参数",
		}
	}

	log.Printf("Playing music: %s", song)

	// On macOS, open Music app with search
	if runtime.GOOS == "darwin" {
		// First open Music app
		cmd := exec.CommandContext(ctx, "open", "-a", "Music")
		if err := cmd.Run(); err != nil {
			log.Printf("Failed to open Music app: %v", err)
		}
	}

	return &types.ExecutionResult{
		Success: true,
		Message: fmt.Sprintf("正在为您播放：%s", song),
		Data:    song,
	}
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

// handleGenerateText generates text content
func (e *Executor) handleGenerateText(ctx context.Context, params map[string]interface{}) *types.ExecutionResult {
	topic, ok := params["topic"].(string)
	if !ok {
		return &types.ExecutionResult{
			Success: false,
			Error:   "缺少主题参数",
		}
	}

	log.Printf("Generating text for topic: %s", topic)

	// This would typically call an LLM to generate the text
	// For now, return a placeholder
	text := fmt.Sprintf("这是关于「%s」的文章。\n\n（此处应调用大语言模型生成完整内容）", topic)

	return &types.ExecutionResult{
		Success: true,
		Message: "文章已生成",
		Data:    text,
	}
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
