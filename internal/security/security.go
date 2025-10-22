package security

import (
	"fmt"
	"log"
	"strings"

	"github.com/deca/voicepilot-eino/internal/config"
)

// SecurityManager manages security and permission checks
type SecurityManager struct {
	allowedActions    map[string]bool
	dangerousKeywords []string
}

// NewSecurityManager creates a new security manager
func NewSecurityManager() *SecurityManager {
	return &SecurityManager{
		allowedActions: map[string]bool{
			"open_app":        true,
			"play_music":      true,
			"generate_text":   true,
			"write_article":   true, // Same as generate_text
			"clarify":         true,
			"error":           true,
			"execute_command": false, // Only allowed in non-safe mode
		},
		dangerousKeywords: []string{
			"rm -rf", "del", "format", "shutdown", "reboot",
			"kill", "pkill", "killall",
			"sudo", "su",
			"chmod", "chown",
			"dd if=", "mkfs",
			"> /dev/", "curl", "wget",
			"passwd", "useradd", "userdel",
		},
	}
}

// ValidateAction validates if an action is allowed
func (s *SecurityManager) ValidateAction(action string, params map[string]interface{}) error {
	log.Printf("Security check for action: %s", action)

	// Check if safe mode is enabled
	if config.AppConfig.EnableSafeMode {
		// In safe mode, only allow explicitly safe actions
		if action == "execute_command" {
			return fmt.Errorf("在安全模式下不允许执行系统命令")
		}
	}

	// Check if action is in allowed list
	allowed, exists := s.allowedActions[action]
	if !exists {
		return fmt.Errorf("未知的操作类型：%s", action)
	}

	if !allowed && config.AppConfig.EnableSafeMode {
		return fmt.Errorf("操作 %s 在安全模式下被禁止", action)
	}

	// Additional validation for specific actions
	switch action {
	case "execute_command":
		if err := s.validateCommand(params); err != nil {
			return err
		}
	case "open_app":
		if err := s.validateAppName(params); err != nil {
			return err
		}
	}

	return nil
}

// validateCommand validates if a command is safe to execute
func (s *SecurityManager) validateCommand(params map[string]interface{}) error {
	command, ok := params["command"].(string)
	if !ok {
		return fmt.Errorf("命令参数无效")
	}

	command = strings.ToLower(command)

	// Check for dangerous keywords
	for _, keyword := range s.dangerousKeywords {
		if strings.Contains(command, strings.ToLower(keyword)) {
			log.Printf("Blocked dangerous command: %s (keyword: %s)", command, keyword)
			return fmt.Errorf("命令包含危险关键字：%s", keyword)
		}
	}

	// Check for dangerous patterns
	if strings.Contains(command, "..") {
		return fmt.Errorf("命令包含危险路径模式")
	}

	if strings.Contains(command, "|") || strings.Contains(command, ";") || strings.Contains(command, "&&") {
		return fmt.Errorf("不允许使用管道或命令链")
	}

	return nil
}

// validateAppName validates if an application name is safe
func (s *SecurityManager) validateAppName(params map[string]interface{}) error {
	appName, ok := params["name"].(string)
	if !ok {
		return fmt.Errorf("应用程序名称参数无效")
	}

	// Check for path traversal attempts
	if strings.Contains(appName, "..") || strings.Contains(appName, "/") || strings.Contains(appName, "\\") {
		return fmt.Errorf("应用程序名称包含非法字符")
	}

	return nil
}

// AddAllowedAction adds an action to the allowed list
func (s *SecurityManager) AddAllowedAction(action string) {
	s.allowedActions[action] = true
	log.Printf("Added allowed action: %s", action)
}

// RemoveAllowedAction removes an action from the allowed list
func (s *SecurityManager) RemoveAllowedAction(action string) {
	s.allowedActions[action] = false
	log.Printf("Removed allowed action: %s", action)
}

// AddDangerousKeyword adds a keyword to the dangerous list
func (s *SecurityManager) AddDangerousKeyword(keyword string) {
	s.dangerousKeywords = append(s.dangerousKeywords, keyword)
	log.Printf("Added dangerous keyword: %s", keyword)
}
