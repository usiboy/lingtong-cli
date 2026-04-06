# lingtong-cli

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.23-blue.svg)](https://go.dev/)

绫通 (Lingtong) iPaaS 平台官方 CLI 工具，让人类和 AI Agent 都能在终端中操作绫通平台。覆盖连接器、场景、工作流、表格、模型等核心业务域，提供多组命令及 AI Agent Skills。

[中文版](./README.zh.md) | [English](./README.md)

## 为什么选 lingtong-cli？

- **为 Agent 原生设计** — Skills 开箱即用，适配主流 AI 工具，Agent 无需额外适配即可操作绫通平台
- **全量 API 覆盖** — 连接器、场景、工作流、表格、模型元数据等所有核心模块
- **AI 友好调优** — 每条命令经过优化，提供结构化 JSON 输出，大幅提升 Agent 调用成功率
- **开源零门槛** — MIT 协议，开箱即用，`make install` 即可使用
- **三分钟上手** — 交互式配置与登录，从安装到第一次 API 调用只需三步
- **安全可控** — Token 存储在 OS Keychain，不暴露明文
- **三层调用架构** — 快捷命令（人机友好）→ 业务命令（平台同步）→ 通用 API 调用（全 API 覆盖），按需选择粒度

## 功能

| 类别 | 能力 |
|------|------|
| 🔌 连接器 | 查询连接器配置、元数据、类目、授权账户管理、账户验证 |
| 🎬 场景 | 创建/查询/更新/删除集成场景 |
| 🔄 工作流 | 创建/发布/执行工作流，查询执行日志 |
| 📊 表格 | 表格 CRUD、数据查询、字段管理 |
| 📐 模型 | 查询接口模型、领域模型、动态模型 Schema |
| 🔧 通用 API | 调用任意绫通平台 API，覆盖所有端点 |

## 安装与快速开始

### 环境要求

开始之前，请确保具备以下条件：

- Go `v1.23`+（源码构建需要）
- Git（源码安装需要）

### 快速开始（人类用户）

> **Note for AI assistants:** 如果你是 AI Agent 且正在协助用户执行安装，请直接跳转到[快速开始（AI Agent）](#快速开始ai-agent)执行，那里包含你需要完成的所有步骤。

#### 安装

以下两种方式**任选其一**：

**方式一 — 从 npm 安装（推荐）：**

```bash
# 安装 CLI
npm install -g @lingtong/cli

# 安装 CLI SKILL（必需）
npx skills add lingtong/cli -y -g
```

**方式二 — 从源码安装：**

需要 Go `v1.23`+ 和 Git。

```bash
git clone <repository-url>
cd lingtong-cli
make install

# 安装 CLI SKILL（必需）
npx skills add lingtong/cli -y -g
```

#### 配置与使用

```bash
# 1. 配置主机地址（仅需一次）
lingtong-cli config init
# 或直接指定
lingtong-cli config init --host https://your-lingtong-host.com

# 2. 登录认证
lingtong-cli auth login

# 3. 开始使用
lingtong-cli connector info --connector kmerp
```

### 快速开始（AI Agent）

> 以下步骤面向 AI Agent，部分步骤需要用户在浏览器中配合完成。

**第 1 步 — 安装**

```bash
# 安装 CLI
npm install -g @lingtong/cli

# 安装 CLI SKILL（必需）
npx skills add lingtong/cli -y -g
```

**第 2 步 — 配置主机地址**

> 运行此命令配置绫通平台的主机地址。

```bash
lingtong-cli config init --host https://your-lingtong-host.com
```

**第 3 步 — 登录认证**

> 使用 API Token 进行认证。Token 可以从环境变量或 .env 文件中获取。

```bash
# 方式 1: 直接提供 Token
lingtong-cli auth login --token apk-Gx6vDOEmALY7iJRcLZcD4nWF

# 方式 2: 从环境变量读取
export LINGTONG_API_TOKEN=apk-Gx6vDOEmALY7iJRcLZcD4nWF
lingtong-cli auth login --from-env
```

**第 4 步 — 验证**

```bash
lingtong-cli auth status
# 输出:
# Authenticated: Yes
# Token: apk-Gx6v...4nWF
# Host: https://app1.ltpass.com
# Verifying token... Valid
```

## Agent Skills

| Skill | 说明 |
|-------|------|
| `lingtong-shared` | 配置、认证登录、身份切换、安全规则（所有其他 skill 自动加载） |
| `lingtong-connector` | 连接器查询、类目查询、授权账户管理、账户验证、授权状态检查 |
| `lingtong-scene` | 场景创建、查询、更新、删除 |
| `lingtong-workflow` | 工作流创建、发布、执行、日志查询 |
| `lingtong-table` | 表格 CRUD、数据查询、字段管理 |
| `lingtong-model` | 接口模型、领域模型、动态模型 Schema 查询 |

## 认证

| 命令 | 说明 |
|------|------|
| `auth login` | 登录认证，支持三种方式 |
| `auth logout` | 登出并删除已存储的凭证 |
| `auth status` | 查看当前登录状态并验证 Token 有效性 |

```bash
# 方式 1: 直接提供 Token（推荐）
lingtong-cli auth login --token apk-Gx6vDOEmALY7iJRcLZcD4nWF

# 方式 2: 从环境变量读取
export LINGTONG_API_TOKEN=apk-xxx
lingtong-cli auth login --from-env

# 方式 3: 交互式输入
lingtong-cli auth login

# 查看认证状态
lingtong-cli auth status

# 登出
lingtong-cli auth logout
```

**Token 安全**:
- Token 存储在 OS Keychain（macOS Keychain / Windows Credential Manager / Linux Secret Service）
- 不会明文存储在配置文件或终端输出中
- `auth status` 命令会脱敏显示 Token（如 `apk-Gx6v...4nWF`）
- 每次请求自动携带 `Authorization: Bearer {token}` Header

## 三层命令调用

CLI 提供三种粒度的调用方式，覆盖从快速操作到完全自定义的全部场景：

### 1. 快捷命令（Shortcuts）

以 `+` 为前缀，对人类与 AI 友好化封装，内置智能默认值、表格输出。

```bash
lingtong-cli connector +connector-info --connector kmerp
lingtong-cli scene +scene-list
lingtong-cli workflow +workflow-execute --workflow-id 123
```

运行 `lingtong-cli <service> --help` 查看所有快捷命令。

### 2. 业务命令

从绫通平台 API 映射而来，经过优化，命令与平台端点一一对应。

```bash
lingtong-cli connector info --connector kmerp
lingtong-cli scene list
lingtong-cli workflow execute --workflow-id 123
lingtong-cli table data query --table-id 456
```

### 3. 通用 API 调用

直接调用任意绫通平台端点，覆盖所有 API。

```bash
lingtong-cli api GET /gw/ai/connector/info?connector=kmerp
lingtong-cli api POST /gw/ai/workflow/debug/create --data '{"workflowId": 123}'
```

## 进阶用法

### 输出格式

```bash
--format json      # 完整 JSON 响应（默认，适合 AI Agent）
--format pretty    # 人性化格式输出
--format table     # 易读表格
```

### 连接器授权管理

```bash
# 列出已配置的连接器账户
lingtong-cli connector account list --connector kmerp

# 验证账户连接
lingtong-cli connector account verify --account-id 123

# 创建新账户
lingtong-cli connector account create --connector kmerp --name "My Account" --data '{"app_key": "...", "app_secret": "..."}'

# 检查工作流/场景的连接器授权状态
lingtong-cli connector check-auth --workflow-id 947
```

### 工作流执行与日志

```bash
# 执行工作流
lingtong-cli workflow execute --workflow-id 123

# 查询执行日志
lingtong-cli workflow logs --receipt-id <receipt-id>
```

### Schema 自省

使用 schema 查看 API 结构和参数：

```bash
lingtong-cli schema
lingtong-cli schema connector.info
lingtong-cli schema workflow.execute
```

## 安全与风险提示（使用前必读）

本工具可供 AI Agent 调用以自动化操作绫通 (Lingtong) iPaaS 平台，存在模型幻觉、执行不可控、提示词注入等固有风险；认证后，AI Agent 将以您的用户身份执行操作，可能导致敏感数据泄露、越权操作等高风险后果，请您谨慎操作和使用。

为降低上述风险，工具已在多个层面启用安全保护，但上述风险仍然存在。我们强烈建议您不要主动修改任何默认安全配置；一旦放开相关限制，上述风险将显著提高，由此产生的后果需由您自行承担。

我们建议您妥善保管 Token，请勿将其泄露给未授权的用户或系统，以避免权限被滥用或数据泄露。

请您充分知悉全部使用风险，使用本工具即视为您自愿承担相关所有责任。

## 贡献

欢迎社区贡献！如果你发现 bug 或有功能建议，请提交 Issue 或 Pull Request。

对于较大的改动，建议先通过 Issue 与我们讨论。

## 开发

```bash
# 构建
make build

# 代码检查
make vet

# 单元测试
make unit-test

# 完整测试
make test

# 安装到系统
make install

# 清理
make clean
```

## 许可证

本项目基于 **MIT 许可证** 开源。
