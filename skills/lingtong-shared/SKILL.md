---
name: lingtong-shared
version: 1.0.0
description: "绫通 CLI 共享基础：配置初始化 (config init)、认证登录 (auth login)、Token 管理、安全规则。当用户需要第一次配置、登录、遇到权限问题或首次使用 lingtong-cli 时触发。"
---

# lingtong-cli 共享规则

## 配置初始化

首次使用需运行 `lingtong-cli config init` 完成主机配置。

```bash
lingtong-cli config init
lingtong-cli config init --host https://your-lingtong-host.com
```

## 认证

### 登录流程

```bash
# 交互式登录
lingtong-cli auth login

# 非交互式登录：从参数传入 Token
lingtong-cli auth login --token <api-token>

# 非交互式登录：从 LINGTONG_API_TOKEN 读取
lingtong-cli auth login --from-env

# 检查认证状态
lingtong-cli auth status

# 登出
lingtong-cli auth logout
```

### Token 管理

- Token 存储在 OS Keychain (macOS Keychain / Windows Credential Manager / Linux Secret Service)
- 不会明文存储在配置文件或终端输出中
- 每次请求自动携带 Authorization: Bearer <token> Header

## 安全规则

- **禁止输出 Token** 到终端明文
- **写入/删除操作前必须确认用户意图**
- 用 `--dry-run` 预览危险请求
- Token 过期时自动提示重新登录

## 输出格式

支持结构化输出的命令通常提供 `--format` 参数，例如 `app`、`scene` 等命令：

```bash
--format json      # Full JSON (default, AI Agent friendly)
--format pretty    # Human-friendly formatted output
--format table     # Readable table output
```

**规则**: AI Agent 调用支持 `--format` 的命令时优先使用 `--format json`，避免解析文本输出。
