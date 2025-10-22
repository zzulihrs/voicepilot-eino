package types

// Intent represents the structured intent parsed from user voice input
type Intent struct {
	Intent     string                 `json:"intent"`
	Parameters map[string]interface{} `json:"parameters"`
	Confidence float64                `json:"confidence,omitempty"`
}

// TaskPlan represents the execution plan for a task
type TaskPlan struct {
	Steps []TaskStep `json:"steps"`
}

// TaskStep represents a single step in a task plan
type TaskStep struct {
	Action     string                 `json:"action"`
	Parameters map[string]interface{} `json:"parameters"`
}

// ExecutionResult represents the result of task execution
type ExecutionResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// VoiceRequest represents a voice interaction request
type VoiceRequest struct {
	AudioPath string `json:"audio_path"`
	SessionID string `json:"session_id,omitempty"`
}

// VoiceResponse represents a voice interaction response
type VoiceResponse struct {
	Text      string `json:"text"`
	AudioURL  string `json:"audio_url,omitempty"`
	SessionID string `json:"session_id"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

// WorkflowContext represents the context passed through the workflow
type WorkflowContext struct {
	SessionID       string                 `json:"session_id"`
	AudioPath       string                 `json:"audio_path,omitempty"`
	RecognizedText  string                 `json:"recognized_text,omitempty"`
	Intent          *Intent                `json:"intent,omitempty"`
	TaskPlan        *TaskPlan              `json:"task_plan,omitempty"`
	ExecutionResult *ExecutionResult       `json:"execution_result,omitempty"`
	ResponseText    string                 `json:"response_text,omitempty"`
	ResponseAudio   string                 `json:"response_audio,omitempty"`
	Context         map[string]interface{} `json:"context,omitempty"`
}
