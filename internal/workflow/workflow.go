package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/deca/voicepilot-eino/internal/executor"
	"github.com/deca/voicepilot-eino/internal/qiniu"
	"github.com/deca/voicepilot-eino/internal/security"
	"github.com/deca/voicepilot-eino/pkg/types"
)

// VoiceWorkflow represents the complete voice interaction workflow
type VoiceWorkflow struct {
	qiniuClient *qiniu.Client
	executor    *executor.Executor
	security    *security.SecurityManager
}

// NewVoiceWorkflow creates a new voice workflow
func NewVoiceWorkflow() *VoiceWorkflow {
	return &VoiceWorkflow{
		qiniuClient: qiniu.NewClient(),
		executor:    executor.NewExecutor(),
		security:    security.NewSecurityManager(),
	}
}

// Execute executes the complete voice interaction workflow
func (w *VoiceWorkflow) Execute(ctx context.Context, audioPath, sessionID string) (*types.VoiceResponse, error) {
	log.Printf("Starting workflow execution for session: %s", sessionID)

	// Create workflow context
	wfCtx := &types.WorkflowContext{
		SessionID: sessionID,
		AudioPath: audioPath,
		Context:   make(map[string]interface{}),
	}

	// Step 1: ASR Node - Speech to Text
	if err := w.asrNode(ctx, wfCtx); err != nil {
		return nil, fmt.Errorf("ASR node failed: %w", err)
	}

	// Step 2: Intent Recognition Node - Parse intent from text
	if err := w.intentNode(ctx, wfCtx); err != nil {
		return nil, fmt.Errorf("Intent node failed: %w", err)
	}

	// Step 3: Planner Node - Create task plan
	if err := w.plannerNode(ctx, wfCtx); err != nil {
		return nil, fmt.Errorf("Planner node failed: %w", err)
	}

	// Step 4: Security Check Node - Validate task safety
	if err := w.securityNode(ctx, wfCtx); err != nil {
		return nil, fmt.Errorf("Security node failed: %w", err)
	}

	// Step 5: Executor Node - Execute the task
	if err := w.executorNode(ctx, wfCtx); err != nil {
		return nil, fmt.Errorf("Executor node failed: %w", err)
	}

	// Step 6: Response Generation Node - Generate response text
	if err := w.responseNode(ctx, wfCtx); err != nil {
		return nil, fmt.Errorf("Response node failed: %w", err)
	}

	// Step 7: TTS Node - Convert response to speech
	if err := w.ttsNode(ctx, wfCtx); err != nil {
		return nil, fmt.Errorf("TTS node failed: %w", err)
	}

	// Build final response
	response := &types.VoiceResponse{
		RecognizedText: wfCtx.RecognizedText, // ASR识别的用户语音
		Text:           wfCtx.ResponseText,   // 系统响应
		AudioURL:       wfCtx.ResponseAudio,  // TTS音频
		SessionID:      sessionID,
		Success:        true,
	}

	log.Printf("Workflow execution completed successfully for session: %s", sessionID)
	return response, nil
}

// ExecuteText executes text-based interaction workflow (skip ASR)
func (w *VoiceWorkflow) ExecuteText(ctx context.Context, text, sessionID string) (*types.VoiceResponse, error) {
	log.Printf("Starting text workflow execution for session: %s", sessionID)

	// Create workflow context with pre-filled text
	wfCtx := &types.WorkflowContext{
		SessionID:      sessionID,
		RecognizedText: text,
		Context:        make(map[string]interface{}),
	}

	// Skip ASR, start from Intent Recognition
	if err := w.intentNode(ctx, wfCtx); err != nil {
		return nil, fmt.Errorf("Intent node failed: %w", err)
	}

	// Planner Node
	if err := w.plannerNode(ctx, wfCtx); err != nil {
		return nil, fmt.Errorf("Planner node failed: %w", err)
	}

	// Security Check Node
	if err := w.securityNode(ctx, wfCtx); err != nil {
		return nil, fmt.Errorf("Security node failed: %w", err)
	}

	// Executor Node
	if err := w.executorNode(ctx, wfCtx); err != nil {
		return nil, fmt.Errorf("Executor node failed: %w", err)
	}

	// Response Generation Node
	if err := w.responseNode(ctx, wfCtx); err != nil {
		return nil, fmt.Errorf("Response node failed: %w", err)
	}

	// TTS Node
	if err := w.ttsNode(ctx, wfCtx); err != nil {
		return nil, fmt.Errorf("TTS node failed: %w", err)
	}

	// Build final response
	response := &types.VoiceResponse{
		RecognizedText: wfCtx.RecognizedText, // 用户输入的文本
		Text:           wfCtx.ResponseText,   // 系统响应
		AudioURL:       wfCtx.ResponseAudio,  // TTS音频
		SessionID:      sessionID,
		Success:        true,
	}

	log.Printf("Text workflow execution completed successfully for session: %s", sessionID)
	return response, nil
}

// asrNode performs speech-to-text conversion
func (w *VoiceWorkflow) asrNode(ctx context.Context, wfCtx *types.WorkflowContext) error {
	log.Printf("ASR Node: Processing audio file")

	text, err := w.qiniuClient.ASR(ctx, wfCtx.AudioPath)
	if err != nil {
		return fmt.Errorf("ASR failed: %w", err)
	}

	wfCtx.RecognizedText = text
	log.Printf("ASR Node: Recognized text: %s", text)
	return nil
}

// intentNode performs intent recognition using LLM
func (w *VoiceWorkflow) intentNode(ctx context.Context, wfCtx *types.WorkflowContext) error {
	log.Printf("Intent Node: Parsing intent from text")

	systemPrompt := `你是一个语音助手的意图识别模块。请分析用户的语音输入，并将其转换为结构化的意图JSON格式。

输出格式：
{
  "intent": "意图类型（如：play_music, write_article, open_app, summarize_file等）",
  "parameters": {"参数名": "参数值"},
  "confidence": 0.95
}

如果无法识别意图，请输出：
{
  "intent": "unknown",
  "parameters": {},
  "confidence": 0.0
}

只输出JSON，不要输出其他内容。`

	messages := []qiniu.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: wfCtx.RecognizedText},
	}

	response, err := w.qiniuClient.ChatCompletion(ctx, messages)
	if err != nil {
		return fmt.Errorf("intent recognition failed: %w", err)
	}

	// Parse intent JSON
	var intent types.Intent
	if err := json.Unmarshal([]byte(response), &intent); err != nil {
		log.Printf("Failed to parse intent JSON: %v, raw response: %s", err, response)
		// Fallback: treat as unknown intent
		intent = types.Intent{
			Intent:     "unknown",
			Parameters: make(map[string]interface{}),
			Confidence: 0.0,
		}
	}

	wfCtx.Intent = &intent
	log.Printf("Intent Node: Recognized intent: %s (confidence: %.2f)", intent.Intent, intent.Confidence)
	return nil
}

// plannerNode creates a task execution plan
func (w *VoiceWorkflow) plannerNode(ctx context.Context, wfCtx *types.WorkflowContext) error {
	log.Printf("Planner Node: Creating task plan")

	// If intent is unknown or confidence is low, ask for clarification
	if wfCtx.Intent.Intent == "unknown" || wfCtx.Intent.Confidence < 0.5 {
		wfCtx.TaskPlan = &types.TaskPlan{
			Steps: []types.TaskStep{
				{
					Action: "clarify",
					Parameters: map[string]interface{}{
						"message": "抱歉，我没有理解您的意思，能否请您再说一遍？",
					},
				},
			},
		}
		return nil
	}

	// Use LLM to create a detailed task plan
	systemPrompt := `你是一个任务规划模块。根据用户的意图，生成详细的执行计划。

输出格式：
{
  "steps": [
    {"action": "动作类型", "parameters": {"参数名": "参数值"}},
    ...
  ]
}

支持的动作类型：
- execute_command: 执行系统命令
- open_app: 打开应用程序
- play_music: 播放音乐
- generate_text: 生成文本
- save_file: 保存文件
- clarify: 请求用户澄清

只输出JSON，不要输出其他内容。`

	intentJSON, _ := json.Marshal(wfCtx.Intent)
	userPrompt := fmt.Sprintf("用户意图：%s\n用户原始输入：%s", string(intentJSON), wfCtx.RecognizedText)

	messages := []qiniu.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	response, err := w.qiniuClient.ChatCompletion(ctx, messages)
	if err != nil {
		return fmt.Errorf("task planning failed: %w", err)
	}

	// Parse task plan JSON
	var taskPlan types.TaskPlan
	if err := json.Unmarshal([]byte(response), &taskPlan); err != nil {
		log.Printf("Failed to parse task plan JSON: %v, raw response: %s", err, response)
		// Fallback: single step execution
		taskPlan = types.TaskPlan{
			Steps: []types.TaskStep{
				{
					Action:     wfCtx.Intent.Intent,
					Parameters: wfCtx.Intent.Parameters,
				},
			},
		}
	}

	wfCtx.TaskPlan = &taskPlan
	log.Printf("Planner Node: Created plan with %d steps", len(taskPlan.Steps))
	return nil
}

// securityNode validates task safety
func (w *VoiceWorkflow) securityNode(ctx context.Context, wfCtx *types.WorkflowContext) error {
	log.Printf("Security Node: Validating task safety")

	for i, step := range wfCtx.TaskPlan.Steps {
		if err := w.security.ValidateAction(step.Action, step.Parameters); err != nil {
			log.Printf("Security check failed for step %d: %v", i, err)
			// Replace dangerous action with a safe error message
			wfCtx.TaskPlan.Steps = []types.TaskStep{
				{
					Action: "error",
					Parameters: map[string]interface{}{
						"message": fmt.Sprintf("出于安全考虑，无法执行该操作：%s", err.Error()),
					},
				},
			}
			break
		}
	}

	log.Printf("Security Node: Validation passed")
	return nil
}

// executorNode executes the task plan
func (w *VoiceWorkflow) executorNode(ctx context.Context, wfCtx *types.WorkflowContext) error {
	log.Printf("Executor Node: Executing task plan")

	result := w.executor.Execute(ctx, wfCtx.TaskPlan)
	wfCtx.ExecutionResult = result

	log.Printf("Executor Node: Execution completed (success: %v)", result.Success)
	return nil
}

// responseNode generates response text based on execution result
func (w *VoiceWorkflow) responseNode(ctx context.Context, wfCtx *types.WorkflowContext) error {
	log.Printf("Response Node: Generating response text")

	// If execution failed, use error message
	if !wfCtx.ExecutionResult.Success {
		wfCtx.ResponseText = wfCtx.ExecutionResult.Error
		return nil
	}

	// Use LLM to generate a natural response
	systemPrompt := `你是一个友好的语音助手。根据任务执行结果，生成简洁、友好的回复。回复应该：
1. 确认任务已完成
2. 简要说明执行结果
3. 语气自然、友好

直接输出回复文本，不要包含额外的格式或标记。`

	resultJSON, _ := json.Marshal(wfCtx.ExecutionResult)
	userPrompt := fmt.Sprintf("用户请求：%s\n执行结果：%s", wfCtx.RecognizedText, string(resultJSON))

	messages := []qiniu.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	response, err := w.qiniuClient.ChatCompletion(ctx, messages)
	if err != nil {
		log.Printf("Response generation failed: %v, using fallback", err)
		response = wfCtx.ExecutionResult.Message
	}

	wfCtx.ResponseText = response
	log.Printf("Response Node: Generated response: %s", response)
	return nil
}

// ttsNode converts response text to speech
func (w *VoiceWorkflow) ttsNode(ctx context.Context, wfCtx *types.WorkflowContext) error {
	log.Printf("TTS Node: Converting response to speech")

	audioURL, err := w.qiniuClient.TTS(ctx, wfCtx.ResponseText)
	if err != nil {
		log.Printf("TTS failed: %v, continuing without audio", err)
		// TTS is optional, continue even if it fails
		return nil
	}

	wfCtx.ResponseAudio = audioURL
	log.Printf("TTS Node: Audio URL: %s", audioURL)
	return nil
}
