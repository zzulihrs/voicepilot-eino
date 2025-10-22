package security

import (
	"testing"

	"github.com/deca/voicepilot-eino/internal/config"
)

func TestNewSecurityManager(t *testing.T) {
	sm := NewSecurityManager()

	if sm == nil {
		t.Fatal("SecurityManager should not be nil")
	}

	if len(sm.allowedActions) == 0 {
		t.Error("allowedActions should not be empty")
	}

	if len(sm.dangerousKeywords) == 0 {
		t.Error("dangerousKeywords should not be empty")
	}
}

func TestValidateAction(t *testing.T) {
	// Setup config
	config.AppConfig = &config.Config{
		EnableSafeMode: true,
	}

	sm := NewSecurityManager()

	tests := []struct {
		name      string
		action    string
		params    map[string]interface{}
		safeMode  bool
		wantError bool
	}{
		{
			name:      "allow open_app in safe mode",
			action:    "open_app",
			params:    map[string]interface{}{"name": "Music"},
			safeMode:  true,
			wantError: false,
		},
		{
			name:      "block execute_command in safe mode",
			action:    "execute_command",
			params:    map[string]interface{}{"command": "ls"},
			safeMode:  true,
			wantError: true,
		},
		{
			name:      "allow play_music",
			action:    "play_music",
			params:    map[string]interface{}{"song": "test"},
			safeMode:  true,
			wantError: false,
		},
		{
			name:      "block unknown action",
			action:    "unknown_action",
			params:    map[string]interface{}{},
			safeMode:  true,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.AppConfig.EnableSafeMode = tt.safeMode
			err := sm.ValidateAction(tt.action, tt.params)

			if (err != nil) != tt.wantError {
				t.Errorf("ValidateAction() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateCommand(t *testing.T) {
	sm := NewSecurityManager()

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name:      "safe command",
			params:    map[string]interface{}{"command": "ls -la"},
			wantError: false,
		},
		{
			name:      "dangerous rm command",
			params:    map[string]interface{}{"command": "rm -rf /"},
			wantError: true,
		},
		{
			name:      "dangerous sudo command",
			params:    map[string]interface{}{"command": "sudo reboot"},
			wantError: true,
		},
		{
			name:      "command with path traversal",
			params:    map[string]interface{}{"command": "cat ../../secret"},
			wantError: true,
		},
		{
			name:      "command with pipe",
			params:    map[string]interface{}{"command": "ls | grep test"},
			wantError: true,
		},
		{
			name:      "command chain",
			params:    map[string]interface{}{"command": "echo hello && rm file"},
			wantError: true,
		},
		{
			name:      "missing command parameter",
			params:    map[string]interface{}{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sm.validateCommand(tt.params)

			if (err != nil) != tt.wantError {
				t.Errorf("validateCommand() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateAppName(t *testing.T) {
	sm := NewSecurityManager()

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name:      "valid app name",
			params:    map[string]interface{}{"name": "Music"},
			wantError: false,
		},
		{
			name:      "app name with path",
			params:    map[string]interface{}{"name": "/Applications/Music.app"},
			wantError: true,
		},
		{
			name:      "app name with path traversal",
			params:    map[string]interface{}{"name": "../Music"},
			wantError: true,
		},
		{
			name:      "missing name parameter",
			params:    map[string]interface{}{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sm.validateAppName(tt.params)

			if (err != nil) != tt.wantError {
				t.Errorf("validateAppName() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestAddAllowedAction(t *testing.T) {
	sm := NewSecurityManager()

	sm.AddAllowedAction("new_action")

	if !sm.allowedActions["new_action"] {
		t.Error("new_action should be in allowed list")
	}
}

func TestRemoveAllowedAction(t *testing.T) {
	sm := NewSecurityManager()

	sm.AddAllowedAction("test_action")
	sm.RemoveAllowedAction("test_action")

	if sm.allowedActions["test_action"] {
		t.Error("test_action should not be allowed after removal")
	}
}

func TestAddDangerousKeyword(t *testing.T) {
	sm := NewSecurityManager()

	initialCount := len(sm.dangerousKeywords)
	sm.AddDangerousKeyword("dangerous_command")

	if len(sm.dangerousKeywords) != initialCount+1 {
		t.Error("dangerous keyword should be added")
	}
}
