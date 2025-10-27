# VoicePilot-Eino

基于七牛云 API 与 Eino 工作流的语音控制电脑助手

- 💡 [问题回答](./docs/PRODUCT_DESIGN.md) - 议题问题回答文档


## 项目简介

VoicePilot-Eino 是一个智能语音控制系统，支持通过语音交互控制电脑执行各种操作。系统集成了七牛云的 ASR（语音识别）、LLM（大语言模型）和 TTS（语音合成）服务，通过工作流式编排实现完整的语音交互闭环。

### 主要功能

- 🎤 **语音识别**：将用户语音转换为文本（支持 WebSocket 和 HTTP 双策略）
- 🧠 **意图识别**：理解用户意图并生成结构化指令
- 📋 **任务规划**：自动分解多步任务
- 🔐 **安全控制**：白名单机制防止危险操作
- 🤖 **任务执行**：根据意图执行相应的系统操作
- 💬 **上下文管理**：支持多轮对话，自动维护对话上下文
- 🔊 **语音反馈**：将执行结果转换为语音输出
- 🌐 **Web 界面**：浏览器端语音交互界面
- 📊 **完整测试**：单元测试覆盖率 89.8%+

### 支持的操作

- 🖥️ 打开应用程序
- 🎵 播放音乐（网易云音乐集成）
- ✍️ 生成文本内容（AI 写作）
- 💻 系统命令执行（安全模式下受限）
- 💬 多轮对话（支持上下文理解）

## 技术架构

### 核心技术栈

- **Go 1.21+**: 主要开发语言
- **Gin**: HTTP Web 框架
- **七牛云 API**: ASR、LLM、TTS 服务
- **工作流引擎**: 自定义节点式工作流

### 系统架构

```
[用户语音输入]
      ↓
【ASR节点（七牛云语音识别）】
      ↓
【LLM节点（七牛云大模型解析意图）】
      ↓
【任务规划节点】
      ↓
【安全检查节点】
      ↓
【执行节点（系统操作）】
      ↓
【反馈生成节点】
      ↓
【TTS节点（七牛云语音输出）】
```

### 项目结构

```
VoicePilot-Eino/
├── cmd/
│   └── server/          # 服务入口
│       └── main.go
├── internal/
│   ├── config/          # 配置管理
│   ├── context/         # 上下文管理模块（多轮对话）
│   ├── qiniu/           # 七牛云 API 客户端
│   ├── workflow/        # 工作流节点（7节点编排）
│   ├── executor/        # 任务执行器
│   ├── security/        # 安全模块
│   └── handler/         # HTTP 处理器
├── pkg/
│   └── types/           # 公共类型定义
├── web/                 # Web 前端
│   ├── index.html       # 主页面
│   └── static/
│       ├── css/         # 样式文件
│       └── js/          # JavaScript
├── docs/                # 项目文档
│   ├── README.md        # 文档索引
│   ├── 设计方案.md      # 总体设计
│   ├── PRODUCT_DESIGN.md # 产品设计
│   ├── IMPLEMENTATION_VERIFICATION.md # 实现验证
│   ├── ASR_README.md    # ASR 技术文档
│   └── CONTEXT_INTEGRATION.md # 上下文集成文档
├── data/
│   └── sessions/        # 会话数据存储（本地持久化）
├── static/
│   └── audio/           # 音频文件存储
├── temp/                # 临时文件
├── tests/               # 测试文件
├── .env.example         # 环境变量模板
├── Makefile             # 构建脚本
└── README.md            # 本文档
```

## 快速开始

### 前置要求

- **Go 1.21 或更高版本**（必需）
  - macOS: `brew install go@1.21` 或从 https://go.dev/dl/ 下载
  - 验证版本: `go version`
- 七牛云账号和 API Key

### 安装步骤

1. 克隆项目

```bash
git clone https://github.com/deca/voicepilot-eino.git
cd voicepilot-eino
```

2. 安装依赖

```bash
make install-deps
```

3. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件，填入你的七牛云 API Key
```

必需配置项：
```bash
QINIU_API_KEY=your-qiniu-api-key-here
```

4. 初始化项目

```bash
make init
```

5. 运行服务

```bash
make run
```

服务将在 `http://localhost:8080` 启动。

6. 访问 Web 界面

打开浏览器访问 `http://localhost:8080` 即可使用 Web 界面进行语音交互。

### 构建可执行文件

```bash
make build
```

生成的可执行文件位于 `bin/voicepilot-eino`。

生产环境构建：

```bash
make build-prod
```

## API 文档

### 1. 健康检查

```
GET /api/health
```

响应：
```json
{
  "status": "ok",
  "time": 1234567890
}
```

### 2. 语音交互

```
POST /api/voice
Content-Type: multipart/form-data
```

参数：
- `audio`: 音频文件（WAV 格式，最大 10MB）
- `session_id`: 会话 ID（可选）

响应：
```json
{
  "text": "已打开应用程序：微信",
  "audio_url": "/static/audio/tts_1234567890.mp3",
  "session_id": "uuid-here",
  "success": true
}
```

### 3. 文本交互

```
POST /api/text
Content-Type: application/json
```

请求体：
```json
{
  "text": "打开微信",
  "session_id": "uuid-here"
}
```

### 4. 获取音频文件

```
GET /static/audio/:filename
```

## 配置说明

### 环境变量

#### 服务器配置
| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| PORT | 服务端口 | 8080 |

#### 七牛云 API 配置
| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| QINIU_API_KEY | 七牛云 API Key | 必填 |
| QINIU_BASE_URL | 七牛云 API 地址 | https://openai.qiniu.com/v1 |

#### TTS 配置
| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| TTS_VOICE_TYPE | TTS 音色类型 | qiniu_zh_female_wwxkjx |
| TTS_ENCODING | TTS 音频格式 | mp3 |
| TTS_SPEED_RATIO | TTS 语速比例 | 1.0 |

#### ASR 配置
| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| ASR_MODEL | ASR 模型 | asr |
| ASR_FORMAT | ASR 音频格式 | wav |

#### LLM 配置
| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| LLM_MODEL | LLM 模型 | deepseek/deepseek-v3.1-terminus |
| LLM_MAX_TOKENS | LLM 最大 Token 数 | 2000 |
| LLM_TEMPERATURE | LLM 温度参数 | 0.7 |

#### 会话和上下文管理
| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| SESSION_STORAGE_PATH | 会话存储路径 | ./data/sessions |
| SESSION_MAX_HISTORY | 单个会话最大历史消息数 | 50 |
| SESSION_EXPIRY_HOURS | 会话过期时间（小时） | 72 |

#### 安全配置
| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| ENABLE_SAFE_MODE | 启用安全模式 | true |
| MAX_AUDIO_SIZE | 最大音频文件大小 | 10485760 (10MB) |

### 安全模式

启用安全模式后（`ENABLE_SAFE_MODE=true`），系统将：

- 禁止执行系统命令
- 启用命令白名单机制
- 过滤危险关键字
- 防止路径遍历攻击

## 开发指南

### 代码规范

项目遵循 Go 官方代码规范和最佳实践：

- 使用 `go fmt` 格式化代码
- 使用 `golangci-lint` 进行代码检查
- 遵循 12-Factor App 原则

### 格式化代码

```bash
make fmt
```

### 运行测试

```bash
make test
```

### 测试覆盖

项目包含以下模块的单元测试：

- `internal/config`: 配置管理测试
- `internal/context`: 上下文管理测试（覆盖率 89.8%）
- `internal/security`: 安全验证测试
- `internal/executor`: 任务执行测试

运行测试：
```bash
# 运行所有测试
go test ./...

# 运行特定模块测试
go test ./internal/config
go test ./internal/context -v -cover
go test ./internal/security
go test ./internal/executor

# 查看测试覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 添加新的操作类型

1. 在 `internal/executor/executor.go` 中注册新的处理器：

```go
e.RegisterHandler("new_action", e.handleNewAction)
```

2. 实现处理函数：

```go
func (e *Executor) handleNewAction(ctx context.Context, params map[string]interface{}) *types.ExecutionResult {
    // 实现逻辑
}
```

3. 在 `internal/security/security.go` 中添加安全规则（如需要）。

## Web 界面使用说明

### 功能特性

- **实时录音**：按住"按住说话"按钮进行录音，松开自动发送
- **文件上传**：支持上传 WAV/MP3 音频文件
- **对话历史**：显示完整的交互记录
- **音频播放**：自动播放语音反馈
- **状态指示**：实时显示连接和处理状态
- **会话管理**：自动维护对话会话

### 浏览器要求

- Chrome 60+
- Firefox 55+
- Safari 11+
- Edge 79+

需要浏览器支持：
- MediaRecorder API（用于录音）
- Fetch API（用于网络请求）

### 使用步骤

1. 确保服务器已启动（`make run` 或 `go run cmd/server/main.go`）
2. 浏览器访问 `http://localhost:8080`
3. 允许浏览器麦克风权限（首次使用时）
4. 按住"按住说话"按钮进行录音
5. 松开按钮，等待处理
6. 查看文本响应和收听语音反馈

### 移动端支持

Web 界面完全响应式设计，支持在移动设备上使用。

## 故障排查

### 常见问题

1. **Go 版本过低导致编译失败**
   - 错误信息：`package XXX is not in GOROOT`
   - 解决方法：升级 Go 到 1.21 或更高版本
   - 验证：`go version` 应显示 >= 1.21

2. **七牛云 API 调用失败**
   - 检查 API Key 是否正确
   - 确认网络连接正常
   - 查看日志中的详细错误信息

3. **音频文件上传失败**
   - 检查文件格式是否为 WAV
   - 确认文件大小不超过 10MB
   - 检查 temp 目录权限

4. **应用程序无法打开**
   - 确认应用程序已安装
   - 在 macOS 上使用准确的应用程序名称
   - 检查系统权限设置

5. **浏览器麦克风权限被拒绝**
   - Chrome: 点击地址栏左侧的锁图标，允许麦克风权限
   - Firefox: 点击地址栏左侧的图标，管理权限
   - Safari: 系统偏好设置 → 安全性与隐私 → 隐私 → 麦克风

### 日志查看

应用程序日志会输出到标准输出，包含详细的执行流程信息。

## 部署

### Docker 部署（TODO）

```bash
docker build -t voicepilot-eino .
docker run -p 8080:8080 --env-file .env voicepilot-eino
```

### 系统服务部署

创建 systemd 服务文件 `/etc/systemd/system/voicepilot-eino.service`：

```ini
[Unit]
Description=VoicePilot-Eino Service
After=network.target

[Service]
Type=simple
User=your-user
WorkingDirectory=/path/to/voicepilot-eino
ExecStart=/path/to/voicepilot-eino/bin/voicepilot-eino
Restart=on-failure
Environment="PATH=/usr/local/bin:/usr/bin:/bin"
EnvironmentFile=/path/to/voicepilot-eino/.env

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable voicepilot-eino
sudo systemctl start voicepilot-eino
```

## 安全建议

1. **生产环境必须启用安全模式**
2. **定期更新依赖包**
3. **使用环境变量管理敏感信息**
4. **配置防火墙限制访问**
5. **定期审查操作日志**

## 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 许可证

MIT License

## 项目文档

完整的项目文档请参考 [docs 目录](./docs/)：

- 📋 [设计方案](./docs/设计方案.md) - 总体架构设计
- 💡 [产品设计](./docs/PRODUCT_DESIGN.md) - 产品功能设计
- ✅ [实现验证](./docs/IMPLEMENTATION_VERIFICATION.md) - 实现验证文档
- 🎤 [ASR 技术文档](./docs/ASR_README.md) - 语音识别技术说明
- 🔄 [上下文集成文档](./docs/CONTEXT_INTEGRATION.md) - 上下文管理集成说明
- 📘 [上下文模块 API](./internal/context/README.md) - 上下文管理模块详细文档

## 相关链接

- [七牛云官网](https://www.qiniu.com/)
- [七牛云 API 文档](https://developer.qiniu.com/)
- [GitHub 仓库](https://github.com/deca/voicepilot-eino)

## 作者

Deca

## 致谢

感谢七牛云提供的 AI 服务支持。
