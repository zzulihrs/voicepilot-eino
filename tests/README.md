# 测试文件

本目录包含 VoicePilot-Eino 项目的测试文件。

## 测试脚本

### test_api.sh
API 端点测试脚本，测试所有 HTTP API 功能。

**使用方法：**
```bash
# 从项目根目录运行
./tests/test_api.sh
```

**前置条件：**
- 服务正在运行（`make run` 或 `./bin/voicepilot-eino`）
- 服务监听在 `localhost:8080`

## Go 测试程序

### test_qiniu_api.go
测试七牛云 API 集成功能。

**运行：**
```bash
cd tests
go run test_qiniu_api.go
```

### test_websocket_asr.go
测试 WebSocket ASR（语音识别）功能。

**运行：**
```bash
cd tests
go run test_websocket_asr.go
```

### test_dual_strategy_asr.go
测试双策略 ASR 实现（WebSocket + HTTP 降级）。

**运行：**
```bash
cd tests
go run test_dual_strategy_asr.go
```

## 测试音频文件

### simple_test.wav
测试用的示例音频文件，用于语音识别测试。
