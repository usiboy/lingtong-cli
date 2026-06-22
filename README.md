# lingtong-cli

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.23-blue.svg)](https://go.dev/)

绫通 (Lingtong) iPaaS 平台官方 CLI 工具，让人类和 AI Agent 都能在终端中操作绫通平台。覆盖连接器、场景、工作流、表格、模型等核心业务域，提供多组命令及 AI Agent Skills。

当前文档：中文

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
| 📦 应用管理 | 应用列表/详情查询、导出/导入/验证/脚手架生成/比较、场景管理 |
| 🔌 连接器 | 查询连接器配置、元数据、类目、授权账户管理、账户验证 |
| 🎬 场景 | 列出、创建、查询集成场景 |
| 🔄 工作流 | 创建/发布/执行工作流，查询执行日志 |
| 📊 表格 | 表格 CRUD、数据查询、字段管理 |
| 📐 模型 | 查询接口模型、领域模型、动态模型 Schema |
| 🔧 通用 API | 调用任意绫通平台 API，覆盖所有端点 |

## 命令参考表

### 应用管理 (App)

| 命令 | 功能 | 必填参数 | 示例 |
|------|------|---------|------|
| `app list` | 列出平台应用 | 无 | `lingtong-cli app list --page-num 1 --page-size 40` |
| `app get` | 查询应用详情 | --application-id 或 --id | `lingtong-cli app get --application-id 123` |
| `app export` | 导出应用配置为 JSON | --app-id, --output | `lingtong-cli app export --app-id 123 --output app.json` |
| `app validate` | 验证应用 JSON 结构 | --file | `lingtong-cli app validate --file app.json` |
| `app import` | 从 JSON 导入应用到平台 | --file | `lingtong-cli app import --file app.json` |
| `app scaffold` | 生成最小应用脚手架 | --name, --source, --target | `lingtong-cli app scaffold --name test --source kmerp --target kingdee` |
| `app scaffold` | 使用内置模板生成应用 | --name, --template | `lingtong-cli app scaffold --name test --template kuaimai-kingdee --output app.json` |
| `app diff` | 比较两个应用配置差异 | --file-a, --file-b | `lingtong-cli app diff --file-a app1.json --file-b app2.json` |
| `app scene list` | 列出应用中的场景 | --file | `lingtong-cli app scene list --file app.json` |
| `app scene add` | 向应用添加场景 | --file, --name, --source, --target | `lingtong-cli app scene add --file app.json --name "同步" --source kmerp --target kingdee` |
| `app scene remove` | 从应用删除场景 | --file, --name | `lingtong-cli app scene remove --file app.json --name "同步"` |

#### 内置应用模板

| 模板名称 | 场景数 | 描述 |
|---------|-------|------|
| `kuaimai-kingdee` | 7 | 快麦 ERP ↔ 金蝶云星辰集成（商品/订单/库存/退货同步） |

### 认证与配置 (Auth & Config)

| 命令 | 功能 | 必填参数 | 示例 |
|------|------|---------|------|
| `config init` | 初始化 CLI 配置 | 无 | `lingtong-cli config init --host https://app.ltpass.com` |
| `auth login` | 登录认证 | --token 或 --from-env | `lingtong-cli auth login --token <api-token>` |
| `auth logout` | 登出并清除凭证 | 无 | `lingtong-cli auth logout` |
| `auth status` | 查看登录状态 | 无 | `lingtong-cli auth status` |

> `app list/get/export/import` 和 `scene list/create/info` 需要先配置 Host。未配置时会提示 `no host configured. Run lingtong-cli config init --host <url> first`。

### 连接器 (Connector)

| 命令 | 功能 | 必填参数 | 示例 |
|------|------|---------|------|
| `connector info` | 查询连接器详情 | --connector | `lingtong-cli connector info --connector kmerp` |
| `connector list` | 列出所有连接器账户 | 无 | `lingtong-cli connector list --app-id 165` |
| `connector category list` | 查询连接器类目 | --connector | `lingtong-cli connector category list --connector kmerp` |
| `connector account list` | 列出连接器账户 | --connector | `lingtong-cli connector account list --connector kmerp --env prod` |
| `connector account verify` | 验证账户连接 | --connector, --account-id | `lingtong-cli connector account verify --connector kmerp --account-id 123` |
| `connector account create` | 创建连接器账户 | --connector, --name, --data | `lingtong-cli connector account create --connector kmerp --name "测试" --data '{"appKey":"x"}'` |
| `connector check-auth` | 检查授权状态 | --workflow-id/--scene-id/--connector | `lingtong-cli connector check-auth --workflow-id 947` |

### 场景 (Scene)

| 命令 | 功能 | 必填参数 | 示例 |
|------|------|---------|------|
| `scene list` | 列出所有场景 | 无 | `lingtong-cli scene list --page 1 --page-size 20 --app-id 165` |
| `scene create` | 创建场景 | --name | `lingtong-cli scene create --name "Order Sync" --description "同步订单"` |
| `scene info` | 查询场景详情 | --scene-id | `lingtong-cli scene info --scene-id 123` |

### 工作流 (Workflow)

#### 基础命令

| 命令 | 功能 | 必填参数 | 示例 |
|------|------|---------|------|
| `workflow list` | 列出工作流 | --app-id | `lingtong-cli workflow list --app-id 165` |
| `workflow info` | 查询工作流详情 | --workflow-id | `lingtong-cli workflow info --workflow-id 123` |
| `workflow create` | 创建工作流 | --name | `lingtong-cli workflow create --name "My WF" --dsl-file wf.json` |
| `workflow update` | 更新工作流 | --workflow-id | `lingtong-cli workflow update --workflow-id 100 --dsl-file wf.json` |
| `workflow delete` | 删除工作流 | --workflow-id | `lingtong-cli workflow delete --workflow-id 100 --confirm` |

#### 执行与日志

| 命令 | 功能 | 必填参数 | 示例 |
|------|------|---------|------|
| `workflow execute` | 执行工作流 | --workflow-id | `lingtong-cli workflow execute --workflow-id 123 --wait` |
| `workflow logs` | 查询执行日志 | --receipt-id | `lingtong-cli workflow logs --receipt-id abc123` |

#### 发布与版本

| 命令 | 功能 | 必填参数 | 示例 |
|------|------|---------|------|
| `workflow publish` | 发布工作流 | --workflow-id | `lingtong-cli workflow publish --workflow-id 100 --version "v1.0.0"` |
| `workflow versions` | 列出版本 | --workflow-id | `lingtong-cli workflow versions --workflow-id 100` |
| `workflow version rollback` | 版本回滚 | --workflow-id, --version | `lingtong-cli workflow version rollback --workflow-id 100 --version "v1.0.0" --dry-run` |

#### API 访问

| 命令 | 功能 | 必填参数 | 示例 |
|------|------|---------|------|
| `workflow api-enable` | 启用 API 访问 | --workflow-id | `lingtong-cli workflow api-enable --workflow-id 100` |
| `workflow api-disable` | 禁用 API 访问 | --workflow-id | `lingtong-cli workflow api-disable --workflow-id 100` |
| `workflow api-test` | 测试开放 API | --app-tag, --params | `lingtong-cli workflow api-test --app-tag abc123 --params '{"k":"v"}'` |

#### 模板管理

| 命令 | 功能 | 必填参数 | 示例 |
|------|------|---------|------|
| `workflow template list` | 列出模板 | 无 | `lingtong-cli workflow template list` |
| `workflow template show` | 查看模板详情 | --name | `lingtong-cli workflow template show --name order_sync` |
| `workflow template use` | 使用模板 | --name, --output | `lingtong-cli workflow template use --name approval --output wf.json` |

#### 验证与测试

| 命令 | 功能 | 必填参数 | 示例 |
|------|------|---------|------|
| `workflow validate` | 验证 DSL | --workflow-id 或 --dsl-file | `lingtong-cli workflow validate --dsl-file wf.json --strict` |
| `workflow test run` | 运行测试 | --workflow-id | `lingtong-cli workflow test run --workflow-id 100 --verbose` |

#### 文档与依赖

| 命令 | 功能 | 必填参数 | 示例 |
|------|------|---------|------|
| `workflow doc generate` | 生成文档 | --workflow-id 或 --dsl-file | `lingtong-cli workflow doc generate --workflow-id 100 --output-file docs.md` |
| `workflow dependency list` | 分析依赖 | --workflow-id 或 --dsl-file | `lingtong-cli workflow dependency list --dsl-file wf.json` |

### 表格 (Table)

| 命令 | 功能 | 必填参数 | 示例 |
|------|------|---------|------|
| `table list` | 列出所有表格 | 无 | `lingtong-cli table list` |
| `table data query` | 查询表格数据 | --table-id | `lingtong-cli table data query --table-id 456 --filter '{"status":"active"}'` |
| `table data create` | 创建表格记录 | --table-id, --data | `lingtong-cli table data create --table-id 456 --data '{"name":"test"}'` |

### 模型 (Model)

| 命令 | 功能 | 必填参数 | 示例 |
|------|------|---------|------|
| `model interface list` | 查询接口模型 | --connector, --filter-model-type | `lingtong-cli model interface list --connector kmerp --filter-model-type query` |
| `model domain get` | 获取领域模型 | --connector, --business | `lingtong-cli model domain get --connector kmerp --business order` |
| `model dynamic view` | 查询动态视图 | --connector, --auth-account-id | `lingtong-cli model dynamic view --connector kmerp --auth-account-id 296` |

### 通用 API

| 命令 | 功能 | 必填参数 | 示例 |
|------|------|---------|------|
| `api <method> <path>` | 调用任意 API | method, path | `lingtong-cli api GET /gw/ai/connector/info?connector=kmerp` |

### 内置工作流模板

| 模板名称 | 节点数 | 描述 |
|---------|-------|------|
| `simple` | 2 | 简单的开始到结束工作流 |
| `connector` | 3 | 包含连接器节点的工作流 |
| `order_sync` | 5 | 从 ERP 系统同步订单 |
| `approval` | 4 | 带条件分支的审批流程 |
| `data_pipeline` | 5 | 带数据转换的管道 |
| `api_wrapper` | 4 | 将连接器封装为 API |

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

# 安装 Skills 到你的 AI 编辑器（必需，自动探测并安装）
lingtong-cli skills install
```

**方式二 — 从源码安装：**

需要 Go `v1.23`+ 和 Git。

```bash
git clone <repository-url>
cd lingtong-cli
make install

# 安装 Skills 到你的 AI 编辑器（必需）
lingtong-cli skills install
```

> Skills 已内置进二进制，`lingtong-cli skills install` 无需源码或 npx。
> 它会自动探测并适配 Claude Code、OpenCode、Qoder、Cursor、Trae、Codex。
> 详见下文 [安装 Skills 到 AI 编辑器](#安装-skills-到-ai-编辑器)。
> （仍兼容旧方式 `npx skills add lingtong/cli -y -g`。）

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

# 安装 Skills 到 AI 编辑器（必需，自动探测）
lingtong-cli skills install
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
lingtong-cli auth login --token <api-token>

# 方式 2: 从环境变量读取
export LINGTONG_API_TOKEN=<api-token>
lingtong-cli auth login --from-env
```

**第 4 步 — 验证**

```bash
lingtong-cli auth status
# 输出:
# Authenticated: Yes
# Token: <masked-token>
# Host: https://app1.ltpass.com
# Verifying token... Valid
```

## Agent Skills

| Skill | 说明 |
|-------|------|
| `lingtong-shared` | 配置、认证登录、身份切换、安全规则（所有其他 skill 自动加载） |
| `lingtong-cli-app` | 应用管理：导出/导入/验证/脚手架生成/比较、场景管理（含 kuaimai-kingdee 模板） |
| `lingtong-connector` | 连接器查询、类目查询、授权账户管理、账户验证、授权状态检查 |
| `lingtong-scene` | 场景列出、创建、查询 |
| `lingtong-workflow` | 工作流创建、发布、执行、日志查询 |
| `lingtong-table` | 表格 CRUD、数据查询、字段管理 |
| `lingtong-model` | 接口模型、领域模型、动态模型 Schema 查询 |

## 安装 Skills 到 AI 编辑器

Skills 已通过 `go:embed` 内置进 `lingtong-cli` 二进制，`lingtong-cli skills install`
会把它们写入各 AI 编辑器读取的目录，让这些 Agent 自动识别 **CLI、命令行与 Skills**。
无需源码树，也无需 `npx`。

**支持的编辑器：**

| 编辑器 | Skills 目录 | Agent 指引文件 |
|--------|-------------|----------------|
| Claude Code | `~/.claude/skills`（全局）/ `.claude/skills`（项目） | `CLAUDE.md` |
| OpenCode | `~/.config/opencode/skills` / `.opencode/skills` | `AGENTS.md` |
| Qoder | `~/.qoder/skills` / `.qoder/skills` | `AGENTS.md` |
| Cursor | `~/.cursor/skills` / `.cursor/skills` | `AGENTS.md` |
| Trae | `~/.trae/skills` / `.trae/skills` | `AGENTS.md` |
| Codex | （无独立 Skills 目录） | `AGENTS.md`、`~/.codex/AGENTS.md` |

```bash
# 自动探测当前项目与用户主目录里已存在的编辑器，逐一安装
lingtong-cli skills install

# 仅安装到指定编辑器
lingtong-cli skills install --editor claude,opencode

# 强制范围：global（用户主目录，全项目共享）| project（当前仓库）
lingtong-cli skills install --scope global
lingtong-cli skills install --scope project

# 预览将要发生的改动，不写文件
lingtong-cli skills install --dry-run

# 查看已安装位置 / 列出内置 Skills / 卸载
lingtong-cli skills status
lingtong-cli skills list
lingtong-cli skills uninstall --editor cursor
```

**行为说明：**

- **自动探测**：不带 `--scope` 时，项目里存在 `.opencode/`、`.cursor/` 等目录就装到项目级；
  `~/.claude`、`~/.codex` 等存在则装到全局。显式 `--editor X` 但未探测到时回退到全局。
- **幂等**：`AGENTS.md` / `CLAUDE.md` 中的绫通内容包裹在受管标记块内，重复安装只替换该块，
  绝不破坏你已有的内容。
- **可卸载**：`skills uninstall` 移除已安装的 skill 目录并清除受管块（若文件仅含受管块则删除文件）。

> 也可用薄封装脚本：`./scripts/install-skills.sh [同上参数]`（适合 CI / 一键安装）。

## 认证

| 命令 | 说明 |
|------|------|
| `auth login` | 登录认证，支持三种方式 |
| `auth logout` | 登出并删除已存储的凭证 |
| `auth status` | 查看当前登录状态并验证 Token 有效性 |

```bash
# 方式 1: 直接提供 Token（推荐）
lingtong-cli auth login --token <api-token>

# 方式 2: 从环境变量读取
export LINGTONG_API_TOKEN=<api-token>
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
- `auth status` 命令会脱敏显示 Token（如 `<masked-token>`）
- 每次请求自动携带 `Authorization: Bearer <token>` Header

## 三层命令调用

CLI 提供三种粒度的调用方式，覆盖从快速操作到完全自定义的全部场景：

### 1. 快捷命令（Shortcuts）

以 `+` 为前缀，对人类与 AI 友好化封装，内置智能默认值、表格输出。

```bash
lingtong-cli connector +connector-info --connector kmerp
lingtong-cli +scene-list
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
# 列出所有连接器账户
lingtong-cli connector list
lingtong-cli connector list --app-id 165

# 列出特定连接器的账户
lingtong-cli connector account list --connector kmerp
lingtong-cli connector account list --connector kmerp --env prod

# 验证账户连接
lingtong-cli connector account verify --connector kmerp --account-id 123

# 创建新账户
lingtong-cli connector account create --connector kmerp --name "快麦测试账号" --env test --data '{"appKey":"...","appSecret":"..."}'

# 检查工作流/场景的连接器授权状态
lingtong-cli connector check-auth --workflow-id 947
lingtong-cli connector check-auth --scene-id 123
lingtong-cli connector check-auth --connector kmerp
```

### 工作流全生命周期管理

#### 基础操作

```bash
# 查询工作流详情
lingtong-cli workflow info --workflow-id 123

# 列出应用下的所有工作流
lingtong-cli workflow list --app-id 165 --page 1 --size 20
```

#### 创建与更新

```bash
# 使用 DSL 文件创建工作流
lingtong-cli workflow create --name "My Workflow" --dsl-file workflow.json --description "描述"

# 使用 DSL 字符串创建
lingtong-cli workflow create --name "Simple WF" --dsl-string '{"nodes":[...],"edges":[...]}'

# 使用内置模板创建 (simple, connector, order_sync, approval, data_pipeline, api_wrapper)
lingtong-cli workflow create --name "Order Sync" --template order_sync

# 更新工作流
lingtong-cli workflow update --workflow-id 100 --dsl-file updated.json
lingtong-cli workflow update --workflow-id 100 --name "New Name"
lingtong-cli workflow update --workflow-id 100 --dsl-file workflow.json --forced  # 强制更新

# 删除工作流
lingtong-cli workflow delete --workflow-id 100
lingtong-cli workflow delete --workflow-id 100 --confirm  # 跳过确认
```

#### 模板管理

```bash
# 列出所有可用模板
lingtong-cli workflow template list

# 查看模板详情
lingtong-cli workflow template show --name order_sync

# 使用模板生成 DSL 文件
lingtong-cli workflow template use --name approval --output approval-flow.json
```

#### 验证与测试

```bash
# 验证工作流 DSL
lingtong-cli workflow validate --workflow-id 100
lingtong-cli workflow validate --dsl-file workflow.json
lingtong-cli workflow validate --dsl-file workflow.json --strict  # 包含警告

# 运行工作流测试
lingtong-cli workflow test run --workflow-id 100
lingtong-cli workflow test run --workflow-id 100 --params '{"key":"value"}'
lingtong-cli workflow test run --workflow-id 100 --test-data-file test-data.json --verbose
```

#### 发布与版本管理

```bash
# 发布工作流
lingtong-cli workflow publish --workflow-id 100
lingtong-cli workflow publish --workflow-id 100 --version "v1.0.0" --memo "Initial release"

# 查看已发布版本
lingtong-cli workflow versions --workflow-id 100 --page 1 --size 20

# 回滚到指定版本
lingtong-cli workflow version rollback --workflow-id 100 --version "v1.0.0"
lingtong-cli workflow version rollback --workflow-id 100 --version "v1.0.0" --dry-run  # 预览
lingtong-cli workflow version rollback --workflow-id 100 --version "v1.0.0" --confirm  # 跳过确认
```

#### API 访问控制

```bash
# 启用开放 API 访问
lingtong-cli workflow api-enable --workflow-id 100

# 禁用开放 API 访问
lingtong-cli workflow api-disable --workflow-id 100

# 测试已发布的工作流 API
lingtong-cli workflow api-test --app-tag abc123 --params '{"key":"value"}'
lingtong-cli workflow api-test --app-tag abc123 --token "custom-token" --params '{"key":"value"}'
```

#### 执行与日志

```bash
# 执行工作流
lingtong-cli workflow execute --workflow-id 123
lingtong-cli workflow execute --workflow-id 123 --wait  # 等待完成
lingtong-cli workflow execute --workflow-id 123 --params '{"key":"value"}'

# 查询执行日志
lingtong-cli workflow logs --receipt-id abc123
lingtong-cli workflow logs --receipt-id abc123 --pid 456  # 查询子上下文
```

#### 文档与依赖分析

```bash
# 生成工作流文档
lingtong-cli workflow doc generate --workflow-id 100
lingtong-cli workflow doc generate --dsl-file workflow.json --output-file docs.md
lingtong-cli workflow doc generate --workflow-id 100 --include-api --include-examples

# 分析工作流依赖
lingtong-cli workflow dependency list --workflow-id 100
lingtong-cli workflow dependency list --dsl-file workflow.json
```

### 场景管理

```bash
# 列出所有场景
lingtong-cli scene list --page 1 --page-size 20
lingtong-cli scene list --app-id 165 --page 1 --page-size 20

# 创建场景
lingtong-cli scene create --name "Order Sync" --description "Sync orders from ERP"

# 查询场景详情
lingtong-cli scene info --scene-id 123
```

### 表格管理

```bash
# 列出所有表格
lingtong-cli table list

# 查询表格数据
lingtong-cli table data query --table-id 456 --filter '{"status":"active"}' --sort '{"field":"created_at"}'

# 创建表格记录
lingtong-cli table data create --table-id 456 --data '{"name":"test","value":100}'
```

### 模型元数据查询

```bash
# 查询连接器接口模型列表
lingtong-cli model interface list --connector kmerp --filter-model-type query
lingtong-cli model interface list --connector kmerp --filter-model-type all --auth-account-id 296

# 获取领域模型
lingtong-cli model domain get --connector kmerp --business order

# 查询动态数据视图
lingtong-cli model dynamic view --connector kmerp --auth-account-id 296
lingtong-cli model dynamic view --connector kmerp --auth-account-id 296 --model-name SalesOrder
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
