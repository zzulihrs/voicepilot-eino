# 上下文管理模块 (Context Manager)

## 概述

上下文管理模块为 VoicePilot-Eino 提供完整的会话管理和对话历史记录功能，支持多轮对话的上下文维护和本地持久化存储。

## 功能特性

- ✅ **会话管理**：支持多会话并发管理
- ✅ **对话历史**：自动记录用户和助手的对话消息
- ✅ **本地持久化**：会话数据自动保存到本地 JSON 文件
- ✅ **自动清理**：支持过期会话的自动清理
- ✅ **历史限制**：可配置单个会话的最大消息数量
- ✅ **上下文数据**：支持自定义键值对存储额外上下文信息
- ✅ **LLM 集成**：提供直接生成 LLM 上下文消息的方法
- ✅ **线程安全**：使用读写锁保证并发安全

## 数据结构

### Message (消息)

```go
type Message struct {
    Role      string    `json:"role"`       // "user" 或 "assistant"
    Content   string    `json:"content"`    // 消息内容
    Timestamp time.Time `json:"timestamp"`  // 时间戳
    Intent    string    `json:"intent,omitempty"` // 用户意图（可选）
}
```

### Session (会话)

```go
type Session struct {
    ID        string                 `json:"id"`
    Messages  []Message              `json:"messages"`
    Context   map[string]interface{} `json:"context,omitempty"`
    CreatedAt time.Time              `json:"created_at"`
    UpdatedAt time.Time              `json:"updated_at"`
}
```

## 使用方法

### 1. 初始化上下文管理器

```go
import (
    "time"
    ctx "github.com/deca/voicepilot-eino/internal/context"
)

// 创建上下文管理器
// 参数：存储路径、最大历史消息数、会话过期时间
cm := ctx.NewContextManager(
    "./data/sessions",  // 存储路径
    100,                // 最多保留 100 条消息
    24 * time.Hour,     // 24 小时后过期
)
```

### 2. 添加对话消息

```go
sessionID := "user-123-session-1"

// 添加用户消息
err := cm.AddUserMessage(sessionID, "播放音乐", "play_music")
if err != nil {
    log.Printf("Error: %v", err)
}

// 添加助手回复
err = cm.AddAssistantMessage(sessionID, "好的，正在为您播放音乐")
if err != nil {
    log.Printf("Error: %v", err)
}

// 或者使用便捷方法同时添加一对交互
err = cm.AddInteraction(
    sessionID,
    "播放音乐",
    "play_music",
    "好的，正在为您播放音乐",
)
```

### 3. 获取对话历史

```go
// 获取所有历史消息
history := cm.GetHistory(sessionID, 0)

// 获取最近 10 条消息
recentHistory := cm.GetHistory(sessionID, 10)

for _, msg := range history {
    fmt.Printf("%s: %s\n", msg.Role, msg.Content)
}
```

### 4. 为 LLM 构建上下文

```go
// 构建 LLM 可用的上下文消息
// 获取最近 5 条消息作为上下文
llmContext := cm.BuildLLMContext(sessionID, 5)

// llmContext 格式：
// [
//   {"role": "user", "content": "..."},
//   {"role": "assistant", "content": "..."},
//   ...
// ]
```

### 5. 自定义上下文数据

```go
// 设置自定义上下文数据
cm.SetContextData(sessionID, "user_preference", "prefer_music_type")
cm.SetContextData(sessionID, "last_played_song", "稻香")

// 获取自定义上下文数据
value, exists := cm.GetContextData(sessionID, "last_played_song")
if exists {
    fmt.Printf("Last played: %v\n", value)
}
```

### 6. 会话管理

```go
// 获取会话信息
info := cm.GetSessionInfo(sessionID)
fmt.Printf("Session info: %+v\n", info)

// 清除特定会话
err := cm.ClearSession(sessionID)

// 清除所有会话
err := cm.ClearAllSessions()

// 清理过期会话
err := cm.CleanupExpiredSessions()
```

### 7. 获取会话摘要

```go
// 获取最近对话的摘要（用于快速了解上下文）
summary := cm.GetSessionSummary(sessionID)
fmt.Printf("Session summary:\n%s\n", summary)
```

## 在 Workflow 中集成

### 修改 VoiceWorkflow 结构

```go
package workflow

import (
    ctx "github.com/deca/voicepilot-eino/internal/context"
    // ... 其他导入
)

type VoiceWorkflow struct {
    qiniuClient    *qiniu.Client
    executor       *executor.Executor
    security       *security.SecurityManager
    contextManager *ctx.ContextManager  // 添加上下文管理器
}

func NewVoiceWorkflow() *VoiceWorkflow {
    return &VoiceWorkflow{
        qiniuClient:    qiniu.NewClient(),
        executor:       executor.NewExecutor(),
        security:       security.NewSecurityManager(),
        contextManager: ctx.NewContextManager(
            "./data/sessions",
            50,  // 保留最近 50 条消息
            24 * time.Hour,
        ),
    }
}
```

### 在意图识别节点中使用上下文

```go
func (w *VoiceWorkflow) intentNode(ctx context.Context, wfCtx *types.WorkflowContext) error {
    log.Printf("Intent Node: Parsing intent from text")

    systemPrompt := `你是一个语音助手的意图识别模块...`

    // 构建包含历史上下文的消息列表
    messages := []qiniu.Message{
        {Role: "system", Content: systemPrompt},
    }

    // 添加历史上下文（最近 5 轮对话）
    historyContext := w.contextManager.BuildLLMContext(wfCtx.SessionID, 5)
    for _, msg := range historyContext {
        messages = append(messages, qiniu.Message{
            Role:    msg["role"],
            Content: msg["content"],
        })
    }

    // 添加当前用户输入
    messages = append(messages, qiniu.Message{
        Role:    "user",
        Content: wfCtx.RecognizedText,
    })

    // 调用 LLM
    response, err := w.qiniuClient.ChatCompletion(ctx, messages)
    // ... 处理响应
}
```

### 在工作流结束时保存对话

```go
func (w *VoiceWorkflow) Execute(ctx context.Context, audioPath, sessionID string) (*types.VoiceResponse, error) {
    // ... 执行工作流 ...

    // 工作流结束后，保存对话到上下文
    err := w.contextManager.AddInteraction(
        sessionID,
        wfCtx.RecognizedText,     // 用户输入
        wfCtx.Intent.Intent,      // 识别的意图
        wfCtx.ResponseText,       // 助手回复
    )
    if err != nil {
        log.Printf("Warning: failed to save interaction to context: %v", err)
    }

    return response, nil
}
```

## 数据存储

会话数据以 JSON 格式存储在指定目录下：

```
./data/sessions/
├── user-123-session-1.json
├── user-456-session-2.json
└── ...
```

每个会话文件示例：

```json
{
  "id": "user-123-session-1",
  "messages": [
    {
      "role": "user",
      "content": "播放音乐",
      "timestamp": "2024-01-01T10:00:00Z",
      "intent": "play_music"
    },
    {
      "role": "assistant",
      "content": "好的，正在为您播放音乐",
      "timestamp": "2024-01-01T10:00:01Z"
    }
  ],
  "context": {
    "last_played_song": "稻香",
    "user_preference": "prefer_music_type"
  },
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:01Z"
}
```

## 配置建议

### 开发环境

```go
cm := ctx.NewContextManager(
    "./data/sessions",
    20,              // 较少的历史消息，方便调试
    1 * time.Hour,   // 1 小时过期
)
```

### 生产环境

```go
cm := ctx.NewContextManager(
    "/var/app/sessions",
    100,             // 更多的历史消息
    72 * time.Hour,  // 3 天过期
)
```

## 定期清理

建议在应用启动时设置定期清理任务：

```go
import "time"

// 每 1 小时清理一次过期会话
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()

    for range ticker.C {
        if err := cm.CleanupExpiredSessions(); err != nil {
            log.Printf("Failed to cleanup expired sessions: %v", err)
        }
    }
}()
```

## 性能考虑

- **并发安全**：使用 `sync.RWMutex` 保证线程安全
- **内存管理**：通过 `maxHistory` 限制单个会话的内存占用
- **自动清理**：定期清理过期会话释放资源
- **懒加载**：会话数据按需从磁盘加载

## 测试

运行单元测试：

```bash
go test -v ./internal/context/
```

测试覆盖：
- ✅ 会话创建和获取
- ✅ 消息添加和检索
- ✅ 历史限制
- ✅ 持久化存储
- ✅ 过期清理
- ✅ 上下文数据管理
- ✅ LLM 上下文构建

## 常见问题

### Q: 如何限制磁盘空间占用？

A: 可以通过以下方式：
1. 设置合理的 `maxHistory` 限制每个会话的消息数
2. 设置较短的 `sessionExpiry` 时间
3. 定期调用 `CleanupExpiredSessions()`

### Q: 如何迁移现有会话数据？

A: 会话数据是标准的 JSON 格式，可以直接读取、修改和写入。

### Q: 支持跨进程共享会话吗？

A: 当前版本通过文件系统持久化，理论上可以跨进程，但建议单进程使用。如需多进程共享，建议使用 Redis 等外部存储。

## 未来改进

- [ ] 支持 Redis 等外部存储后端
- [ ] 添加会话搜索功能
- [ ] 支持会话导出和分析
- [ ] 添加会话分组管理
- [ ] 支持更细粒度的权限控制
- [ ] 通过langfuse对会话过程进行追踪
