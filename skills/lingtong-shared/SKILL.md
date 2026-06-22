---
name: lingtong-shared
version: 1.2.0
description: "绫通 CLI 共享基础：配置初始化 (config init)、认证登录 (auth login)、Token 管理、安全规则、多环境 Profile、API 文档浏览 (schema)、版本更新 (update)、系统通知、Skills 安装到 AI 编辑器 (skills install)。当用户需要第一次配置、登录、遇到权限问题、切换 dev/staging/prod 环境、浏览 API、检查更新、把 Skills 装到 Claude Code/OpenCode/Qoder/Cursor/Trae/Codex 或首次使用 lingtong-cli 时触发。关键词：config、profile、schema、update、notice、skills install、多环境、编辑器。"
---

# lingtong-cli 共享规则

## 安装 Skills 到 AI 编辑器

Skills 已内置进 `lingtong-cli` 二进制，用 `skills install` 即可写入各 AI 编辑器读取的目录
（无需源码或 npx），让 Agent 识别到 CLI、命令行与 Skills：

```bash
lingtong-cli skills install                       # 自动探测并安装（推荐）
lingtong-cli skills install --editor claude,opencode
lingtong-cli skills install --scope global         # 或 project
lingtong-cli skills install --dry-run              # 预览
lingtong-cli skills status                          # 查看已安装位置
lingtong-cli skills uninstall --editor cursor       # 卸载
```

支持编辑器：Claude Code、OpenCode、Qoder、Cursor、Trae、Codex。安装是**幂等**的——
`AGENTS.md` / `CLAUDE.md` 中的内容写在受管标记块内，重复安装只替换该块，不破坏既有内容。

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

检查项：CLI 版本、配置文件、host 格式、认证状态（token 脱敏显示）、API 端点可达性与延迟、相关环境变量,以及 **Service Spec 来源**(env 覆盖 / `~/.lingtong-cli/openapi.json` / 内嵌 + operation 数)。

## 多环境 Profile (config profile)

用 Profile 在 dev / staging / prod 之间切换。**每个 Profile 携带独立的 host、brand 和 Token**(Token 存各自的 Keychain 账户,切换 Profile 即切换凭据)。

```bash
# 新增 Profile(可选 --token 存入该环境专属凭据)
lingtong-cli config profile add --name prod --host https://api.example.com --token apk-xxxx
lingtong-cli config profile add --name dev  --host https://dev.example.com

# 列出 / 查看当前
lingtong-cli config profile list
lingtong-cli config profile current

# 持久切换当前 Profile
lingtong-cli config profile use prod

# 单条命令临时指定 Profile(不改持久状态)
lingtong-cli scene list --profile dev

# 删除(破坏性,需 --yes;会一并清理该 Profile 的 Token;不能删当前激活的 Profile)
lingtong-cli config profile remove --name dev --yes
```

**解析优先级**:`--profile` 标志 > 持久的 `currentProfile` > 顶层配置。`LINGTONG_TOKEN` 环境变量始终覆盖 Profile Token。

## API 文档浏览 (schema)

离线浏览内嵌 OpenAPI(与 `service` / `doctor` 共用同一份 spec,可被 `LINGTONG_OPENAPI` 或 `~/.lingtong-cli/openapi.json` 覆盖):

```bash
lingtong-cli schema list                 # 按模块分组列出所有路径
lingtong-cli schema path /scene/list      # 某路径的方法/参数/类型(路径无前导斜杠也可)
lingtong-cli schema module scene          # 某模块下全部操作
lingtong-cli schema search "场景"         # 按 summary/description/path 关键词搜索
```

## 版本更新 (update)

```bash
lingtong-cli update            # 检查并显示更新指引(npm / go install / 二进制)
lingtong-cli update --check    # 仅检查是否有新版本
lingtong-cli update --version v1.3.0   # 指定目标版本的更新指引
```

检查端点可用 `LINGTONG_UPDATE_URL` 覆盖。网络失败时优雅降级为手动检查指引。

### 系统通知 (_notice)

开启 `--envelope` 时,信封可能带 `_notice` 字段(如有新版本可用)。通知抓取是**非阻塞**的(后台进行,就绪才注入,绝不拖慢命令),并按 24h 在 `~/.lingtong-cli/notice-cache.json` 缓存,避免每次联网。

```json
{
  "ok": true,
  "data": { "...": "..." },
  "_notice": { "update": { "version": "v1.3.0", "url": "https://.../tag/v1.3.0" } }
}
```
