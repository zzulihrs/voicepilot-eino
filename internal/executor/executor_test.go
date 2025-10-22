package executor

import (
	"context"
	"testing"

	"github.com/deca/voicepilot-eino/pkg/types"
)

func TestNewExecutor(t *testing.T) {
	exec := NewExecutor()

	if exec == nil {
		t.Fatal("Executor should not be nil")
	}

	if len(exec.handlers) == 0 {
		t.Error("handlers should not be empty")
	}

	// Check required handlers are registered
	requiredHandlers := []string{"open_app", "play_music", "execute_command", "generate_text", "clarify", "error"}
	for _, handler := range requiredHandlers {
		if _, exists := exec.handlers[handler]; !exists {
			t.Errorf("Required handler '%s' not registered", handler)
		}
	}
}

func TestRegisterHandler(t *testing.T) {
	exec := NewExecutor()

	customHandler := func(ctx context.Context, params map[string]interface{}) *types.ExecutionResult {
		return &types.ExecutionResult{Success: true, Message: "custom"}
	}

	exec.RegisterHandler("custom_action", customHandler)

	if _, exists := exec.handlers["custom_action"]; !exists {
		t.Error("custom_action should be registered")
	}
}

func TestExecute(t *testing.T) {
	exec := NewExecutor()
	ctx := context.Background()

	tests := []struct {
		name      string
		plan      *types.TaskPlan
		wantError bool
	}{
		{
			name: "clarify action",
			plan: &types.TaskPlan{
				Steps: []types.TaskStep{
					{
						Action: "clarify",
						Parameters: map[string]interface{}{
							"message": "test message",
						},
					},
				},
			},
			wantError: false,
		},
		{
			name: "error action",
			plan: &types.TaskPlan{
				Steps: []types.TaskStep{
					{
						Action: "error",
						Parameters: map[string]interface{}{
							"message": "test error",
						},
					},
				},
			},
			wantError: true,
		},
		{
			name: "unknown action",
			plan: &types.TaskPlan{
				Steps: []types.TaskStep{
					{
						Action:     "unknown_action",
						Parameters: map[string]interface{}{},
					},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exec.Execute(ctx, tt.plan)

			if result.Success == tt.wantError {
				t.Errorf("Execute() success = %v, wantError %v", result.Success, tt.wantError)
			}
		})
	}
}

func TestHandleClarify(t *testing.T) {
	exec := NewExecutor()
	ctx := context.Background()

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantMsg string
	}{
		{
			name:    "with custom message",
			params:  map[string]interface{}{"message": "custom clarification"},
			wantMsg: "custom clarification",
		},
		{
			name:    "without message",
			params:  map[string]interface{}{},
			wantMsg: "抱歉，我没有理解您的意思，能否请您再说一遍？",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exec.handleClarify(ctx, tt.params)

			if !result.Success {
				t.Error("handleClarify should always succeed")
			}

			if result.Message != tt.wantMsg {
				t.Errorf("Expected message '%s', got '%s'", tt.wantMsg, result.Message)
			}
		})
	}
}

func TestHandleError(t *testing.T) {
	exec := NewExecutor()
	ctx := context.Background()

	tests := []struct {
		name     string
		params   map[string]interface{}
		wantMsg  string
	}{
		{
			name:    "with custom error",
			params:  map[string]interface{}{"message": "custom error"},
			wantMsg: "custom error",
		},
		{
			name:    "without message",
			params:  map[string]interface{}{},
			wantMsg: "执行过程中发生错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exec.handleError(ctx, tt.params)

			if result.Success {
				t.Error("handleError should always fail")
			}

			if result.Error != tt.wantMsg {
				t.Errorf("Expected error '%s', got '%s'", tt.wantMsg, result.Error)
			}
		})
	}
}

func TestHandleGenerateText(t *testing.T) {
	exec := NewExecutor()
	ctx := context.Background()

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name:      "with topic",
			params:    map[string]interface{}{"topic": "AI"},
			wantError: false,
		},
		{
			name:      "without topic",
			params:    map[string]interface{}{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exec.handleGenerateText(ctx, tt.params)

			if result.Success == tt.wantError {
				t.Errorf("handleGenerateText() success = %v, wantError %v", result.Success, tt.wantError)
			}
		})
	}
}

func TestHandlePlayMusic(t *testing.T) {
	exec := NewExecutor()
	ctx := context.Background()

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name:      "with song name",
			params:    map[string]interface{}{"song": "test song"},
			wantError: false,
		},
		{
			name:      "without song name",
			params:    map[string]interface{}{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exec.handlePlayMusic(ctx, tt.params)

			if result.Success == tt.wantError {
				t.Errorf("handlePlayMusic() success = %v, wantError %v", result.Success, tt.wantError)
			}
		})
	}
}

func TestHandleOpenApp(t *testing.T) {
	exec := NewExecutor()
	ctx := context.Background()

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name:      "without app name",
			params:    map[string]interface{}{},
			wantError: true,
		},
		{
			name:      "with invalid app name",
			params:    map[string]interface{}{"name": "NonExistentApp12345"},
			wantError: true, // Will likely fail as app doesn't exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exec.handleOpenApp(ctx, tt.params)

			if result.Success == tt.wantError {
				t.Errorf("handleOpenApp() success = %v, wantError %v", result.Success, tt.wantError)
			}
		})
	}
}

func TestHandleExecuteCommand(t *testing.T) {
	exec := NewExecutor()
	ctx := context.Background()

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name:      "without command",
			params:    map[string]interface{}{},
			wantError: true,
		},
		{
			name:      "with empty command",
			params:    map[string]interface{}{"command": ""},
			wantError: true,
		},
		{
			name:      "with simple command",
			params:    map[string]interface{}{"command": "echo hello"},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exec.handleExecuteCommand(ctx, tt.params)

			if result.Success == tt.wantError {
				t.Errorf("handleExecuteCommand() success = %v, wantError %v", result.Success, tt.wantError)
			}
		})
	}
}
