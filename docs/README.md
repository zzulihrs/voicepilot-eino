# VoicePilot-Eino 项目文档

本目录包含 VoicePilot-Eino 项目的所有设计文档、实现文档和技术说明。

## 文档索引

### 📋 设计文档

| 文档 | 说明 | 更新日期 |
|------|------|---------|
| [设计方案.md](./设计方案.md) | 项目总体设计方案，包括架构设计和工作流设计 | 2024-10-22 |
| [PRODUCT_DESIGN.md](./PRODUCT_DESIGN.md) | 产品设计文档，详细的功能规划和用户体验设计 | 2024-10-22 |

### ✅ 实现验证

| 文档 | 说明 | 版本 | 更新日期 |
|------|------|------|---------|
| [IMPLEMENTATION_VERIFICATION.md](./IMPLEMENTATION_VERIFICATION.md) | 实现验证文档，证明符合"不使用 Agent"的要求 | v1.1 | 2024-10-23 |

### 🔧 技术文档

| 文档 | 说明 | 更新日期 |
|------|------|---------|
| [ASR_README.md](./ASR_README.md) | ASR（语音识别）模块技术说明 | 2024-10-22 |
| [CONTEXT_INTEGRATION.md](./CONTEXT_INTEGRATION.md) | 上下文管理模块集成说明 | 2024-10-23 |

## 文档层级

```
docs/
├── README.md                          # 本文档（文档索引）
├── 设计方案.md                         # 总体设计方案
├── PRODUCT_DESIGN.md                  # 产品设计文档
├── IMPLEMENTATION_VERIFICATION.md     # 实现验证文档
├── ASR_README.md                      # ASR 技术文档
└── CONTEXT_INTEGRATION.md             # 上下文集成文档
```

## 快速导航

### 新手入门
1. 先阅读 [设计方案.md](./设计方案.md) 了解系统架构
2. 然后查看 [PRODUCT_DESIGN.md](./PRODUCT_DESIGN.md) 了解功能设计
3. 最后参考 [IMPLEMENTATION_VERIFICATION.md](./IMPLEMENTATION_VERIFICATION.md) 验证实现

### 技术实现
- **语音识别**：查看 [ASR_README.md](./ASR_README.md)
- **上下文管理**：查看 [CONTEXT_INTEGRATION.md](./CONTEXT_INTEGRATION.md)
- **核心工作流**：查看 `/internal/workflow/` 目录

### 代码文档
- **上下文管理模块**：[/internal/context/README.md](../internal/context/README.md)

## 文档维护

### 文档版本控制

所有文档都通过 Git 进行版本控制，重要文档应标注版本号和更新日期。

### 更新规范

1. **重大更新**：更新版本号（如 v1.0 → v1.1）
2. **小幅修改**：更新日期
3. **新增文档**：添加到本索引文件

## 相关链接

- [项目 README](../README.md) - 项目主页和快速开始
- [代码仓库](https://github.com/deca/voicepilot-eino)
- [问题追踪](https://github.com/deca/voicepilot-eino/issues)
