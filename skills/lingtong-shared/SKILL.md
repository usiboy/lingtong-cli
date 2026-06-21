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

## 全局配置选项

### 省略 Null 字段 (--omit-null)

默认情况下，CLI 会在 JSON 输出中省略值为 `null` 的字段，使输出更简洁。

```bash
# 默认行为：省略 null 字段（推荐）
lingtong-cli connector account list

# 显示所有字段（包括 null）
lingtong-cli connector account list --omit-null=false
```

**配置示例**：

默认输出（omit-null=true）：
```json
{
  "id": 250,
  "name": "测试账号",
  "connector": "kmerp",
  "env": "test"
}
```

完整输出（omit-null=false）：
```json
{
  "id": 250,
  "name": "测试账号",
  "connector": "kmerp",
  "connectorTitle": null,
  "env": "test",
  "envs": null,
  "icon": null
}
```

**适用场景**：
- 默认模式适合日常使用和 AI Agent 处理
- 完整模式适合调试和查看完整数据结构

## 输出格式

支持结构化输出的命令通常提供 `--format` 参数，例如 `app`、`scene` 等命令：

```bash
--format json      # Full JSON (default, AI Agent friendly)
--format pretty    # Human-friendly formatted output
--format table     # Readable table output
```

**规则**: AI Agent 调用支持 `--format` 的命令时优先使用 `--format json`，避免解析文本输出。

## AI Agent 输出契约

### 标准信封 (--envelope)

默认关闭（为兼容现有脚本/skills/MCP 的解析，处于迁移期）。开启后所有输出被包裹成统一结构，便于 Agent 可靠解析：

```bash
lingtong-cli connector info --connector kmerp --envelope
```

成功：
```json
{
  "ok": true,
  "identity": "connector.info",
  "data": { "name": "kmerp" }
}
```

失败：
```json
{
  "ok": false,
  "error": {
    "type": "authentication",
    "subtype": "token_invalid",
    "message": "API error (401): {\"error\":\"unauthorized\"}",
    "hint": "Run 'lingtong-cli auth login' to refresh your token",
    "retryable": false
  }
}
```

**规则**：Agent 在自动化场景应加 `--envelope`，通过 `.ok` 判断成功/失败，通过 `.error.type` / `.error.subtype` 决定处理策略，`.error.retryable` 为 `true` 时可重试。

### 类型化退出码

无论是否开启 `--envelope`，进程退出码都具备语义，可直接用于脚本判断：

| 退出码 | 含义 | error.type |
|--------|------|------------|
| 0 | 成功 | — |
| 1 | API / 通用错误 | `api` |
| 2 | 参数校验失败 | `validation` |
| 3 | 认证失败（token 无效/过期/无权限） | `authentication` |
| 4 | 网络错误（超时 / DNS / 连接被拒） | `network` |
| 5 | 内部错误 | `internal` |
| 6 | 内容安全违规 | — |
| 10 | 高风险操作需 `--yes` 确认 | — |

```bash
lingtong-cli connector info --connector invalid --envelope
echo $?   # 1 (api) 或 3 (authentication)，取决于服务端响应
```

### jq 过滤 (--jq / -q)

所有命令支持全局 `--jq`，在输出前应用 jq 表达式（基于 gojq，无需安装外部 jq）。标量结果按 `jq -r` 裸输出，复杂结果按缩进 JSON 输出。

```bash
# 提取字段列表
lingtong-cli workflow list --jq '.data[].name'

# 过滤
lingtong-cli connector list --jq '.data[] | select(.status=="enabled")'

# 与 --envelope 组合时，表达式作用于整个信封（用 .data 进入数据）
lingtong-cli connector info --connector kmerp --envelope --jq '.data.name'
```

非法的 jq 表达式会以退出码 2（validation）提前报错；遍历 null 等运行期错误会给出友好提示和顶层字段名。

## Shell 自动补全 (completion)

```bash
lingtong-cli completion bash > /etc/bash_completion.d/lingtong-cli
lingtong-cli completion zsh  > ~/.zsh/completions/_lingtong-cli
lingtong-cli completion fish > ~/.config/fish/completions/lingtong-cli.fish
lingtong-cli completion powershell > lingtong-cli.ps1
```

## 健康检查 (doctor)

排查配置/认证/网络问题的自助诊断：

```bash
lingtong-cli doctor
```

检查项：CLI 版本、配置文件、host 格式、认证状态（token 脱敏显示）、API 端点可达性与延迟、相关环境变量。
