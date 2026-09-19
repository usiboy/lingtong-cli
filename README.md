# lingtong-cli

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.23-blue.svg)](https://go.dev/)

绫通 (Lingtong) iPaaS 平台官方 CLI 工具。它让人类用户和 AI Agent 可以在终端中安全、稳定地操作绫通平台，覆盖应用、连接器、连接器工厂、场景、工作流、表格、透视表、模型元数据和全量 OpenAPI 端点。

当前文档：中文

## 能力总览

| 领域 | 能力 |
|------|------|
| 应用管理 | 应用列表/详情、导出、导入、验证、脚手架、配置差异、应用内场景维护 |
| 配置与认证 | Host 初始化、多环境 Profile、多身份 Auth、Token Keychain 存储、租户信息缓存 |
| 连接器 | 连接器信息、类目、账户授权、账户验证、方法列表、字段 Schema、通用方法调用、调用缓存 |
| 连接器工厂 | 自定义连接器定义、HTTP 接口、脚本的查询、保存、运行、删除 |
| 场景 | 场景列表、详情、草稿创建、更新、复制、开关、发布、版本、触发条件、字段映射调试 |
| 工作流 | 创建、更新、删除、验证、测试、发布、回滚、执行、日志、Open API、模板、文档、依赖分析 |
| 表格 | 表格 CRUD、Schema 管理、数据查询/创建/更新/批量更新/统计/删除、视图、分组、统计指标 |
| 透视表 | 透视表视图、配置读取/保存、维度/度量聚合查询 |
| 模型元数据 | 接口模型、领域模型、动态模型视图 |
| API 自省 | OpenAPI Schema 浏览、模块搜索、自动生成 `service` 命令、原始 `api` 调用 |
| Agent Skills | 11 个内置 Skills，可安装到 Claude Code、OpenCode、Qoder、Cursor、Trae、Codex |

## 为什么选 lingtong-cli？

- 为 AI Agent 原生设计：默认 JSON 输出，支持 `--envelope`、`--jq`、Skills 安装和安全提示。
- 命令粒度完整：快捷命令、专用业务命令、OpenAPI 自动命令、原始 API 调用按需选择。
- 与平台同步演进：近期已覆盖多租户身份、Profile、Schema 自省、service 自动命令、透视表、表格视图、连接器通用调用等能力。
- 安全可控：Token 存储在 OS Keychain，删除/批量删除等破坏性操作要求 `--yes`。
- 可自检可更新：提供 `doctor`、`update`、`schema`、`skills status` 等运维命令。

## 安装与快速开始

### 环境要求

- Go `1.23`+（源码构建需要）
- Git（源码安装需要）
- Node.js/npm（npm 安装需要）

### 安装 CLI

```bash
# 方式一：通过 npx 下载并安装到用户目录
npx @lingtong-cli/cli@latest install
lingtong-cli --version

# 方式二：npm 全局安装
npm install -g @lingtong-cli/cli

# 方式三：源码安装
cd lingtong-cli
make install

```

`npx ... install` 默认安装到 macOS/Linux 的 `~/.local/bin`，Windows 的
`%LOCALAPPDATA%\Lingtong\bin`。可通过 `LINGTONG_CLI_INSTALL_DIR` 指定目录；
如果 GitHub Release 不可达，可通过 `LINGTONG_CLI_DOWNLOAD_BASE` 指定 HTTPS 制品镜像。

也可以把以下指令发送给 Cursor、Claude Code、OpenCode 等 Agent：

```text
帮我安装绫通 CLI，并按照 doc/LINGTONG_CLI_INSTALL_GUIDE_AGENT.md 完成版本和环境自检。
```

### 初始化、登录、验证

```bash
# 1. 配置平台 Host
lingtong-cli config init --host https://your-lingtong-host.com

# 2. 登录。推荐通过环境变量避免 Token 暴露在 shell history 中
export LINGTONG_API_TOKEN=apk-xxx
lingtong-cli auth login --from-env --label default

# 3. 自检
lingtong-cli auth status
lingtong-cli doctor

# 4. 调用一个只读命令
lingtong-cli connector info --connector kmerp --format json
```

### 安装 Agent Skills

```bash
# 自动探测当前项目和用户主目录中的 AI 编辑器
lingtong-cli skills install

# 仅安装到指定编辑器
lingtong-cli skills install --editor claude,opencode

# 预览，不写文件
lingtong-cli skills install --dry-run

# 查看状态
lingtong-cli skills status
```

Skills 已通过 `go:embed` 内置到二进制中，无需源码树，也无需 `npx`。

## 输出与全局约定

### 输出格式

多数业务命令支持：

```bash
--format json      # 默认，适合 AI Agent
--format pretty    # 人类可读
--format table     # 表格输出
```

根命令全局支持：

| 标志 | 说明 |
|------|------|
| `--profile <name>` | 使用指定配置 Profile，如 `dev`、`staging`、`prod` |
| `--auth <label>` | 使用指定认证身份，如 `company-a` |
| `--envelope` | 输出 `{ok,data,error}` 机器可读信封 |
| `--jq, -q <expr>` | 用 jq 表达式过滤 JSON 输出 |
| `--omit-null` | JSON 输出省略 null 字段，默认开启 |

示例：

```bash
lingtong-cli scene list --envelope --jq '.data.result.total'
lingtong-cli table data query --basic-data-id 123 --format json --jq '.result.data[0]'
```

### 破坏性操作

删除、批量删除、移除 Profile、删除工厂等操作默认拒绝执行，必须显式传入 `--yes`。

```bash
lingtong-cli table data delete --schema-id 2546 --id 33832272 --yes
lingtong-cli scene delete --scene-id 123 --yes
lingtong-cli factory delete --id 12 --yes
```

## 命令体系

CLI 提供四层调用方式：

| 层级 | 适用场景 | 示例 |
|------|----------|------|
| 快捷命令 | 高频只读/执行操作，适合快速验证 | `lingtong-cli +scene-list` |
| 专用业务命令 | 优先使用，参数和输出经过优化 | `lingtong-cli workflow execute --workflow-id 123` |
| OpenAPI 自动命令 | 长尾 API，不想手写 path 时使用 | `lingtong-cli service scene list --pageNum 1 --pageSize 10` |
| 原始 API | 完全自定义 method/path/body | `lingtong-cli api GET /gw/ai/connector/info?connector=kmerp` |

快捷命令是根级命令，不挂在业务子命令下：

```bash
lingtong-cli +connector-info --connector kmerp
lingtong-cli +scene-list --page 1 --page-size 20
lingtong-cli +workflow-execute --workflow-id 123
```

## 命令参考

### 配置与认证

| 命令 | 功能 | 示例 |
|------|------|------|
| `config init` | 初始化 Host | `lingtong-cli config init --host https://app.ltpass.com` |
| `config show` | 查看当前配置 | `lingtong-cli config show` |
| `config delete` | 删除配置 | `lingtong-cli config delete` |
| `config profile list` | 列出 Profile | `lingtong-cli config profile list` |
| `config profile current` | 查看当前 Profile | `lingtong-cli config profile current` |
| `config profile add` | 添加 Profile | `lingtong-cli config profile add --name dev --host https://dev.example.com --token apk-xxx` |
| `config profile use` | 切换 Profile | `lingtong-cli config profile use prod` |
| `config profile remove` | 删除 Profile | `lingtong-cli config profile remove --name dev --yes` |
| `auth login` | 登录并存储 Token | `lingtong-cli auth login --from-env --label company-a` |
| `auth status` | 查看认证状态 | `lingtong-cli auth status` |
| `auth list` | 列出身份 | `lingtong-cli auth list` |
| `auth use` | 切换身份 | `lingtong-cli auth use company-a` |
| `auth remove` | 删除身份 | `lingtong-cli auth remove --label company-a` |
| `auth logout` | 登出当前或全部身份 | `lingtong-cli auth logout --all` |

Token 读取顺序以代码为准：`LINGTONG_TOKEN` 可覆盖运行时 Token，`auth login --from-env` 读取 `LINGTONG_API_TOKEN`。

### 应用管理

| 命令 | 功能 | 示例 |
|------|------|------|
| `app list` | 列出应用 | `lingtong-cli app list --page-num 1 --page-size 40` |
| `app get` | 查询应用详情 | `lingtong-cli app get --application-id 123` |
| `app export` | 导出应用 JSON | `lingtong-cli app export --app-id 123 --output app.json` |
| `app validate` | 验证应用 JSON | `lingtong-cli app validate --file app.json` |
| `app import` | 导入应用 JSON | `lingtong-cli app import --file app.json --dry-run` |
| `app scaffold` | 生成应用脚手架 | `lingtong-cli app scaffold --name demo --source kmerp --target kingdee --output app.json` |
| `app scaffold` | 使用模板生成应用 | `lingtong-cli app scaffold --name demo --template kuaimai-kingdee --output app.json` |
| `app diff` | 比较应用 JSON | `lingtong-cli app diff --file-a app-dev.json --file-b app-prod.json` |
| `app scene list` | 列出应用文件内场景 | `lingtong-cli app scene list --file app.json` |
| `app scene add` | 添加应用文件内场景 | `lingtong-cli app scene add --file app.json --name 商品同步 --source kmerp --target kingDeeCloudStar` |
| `app scene remove` | 删除应用文件内场景 | `lingtong-cli app scene remove --file app.json --name 商品同步` |

内置应用模板：

| 模板 | 场景数 | 描述 |
|------|--------|------|
| `kuaimai-kingdee` | 7 | 快麦 ERP 与金蝶云星辰集成，覆盖商品、销售订单、库存、退货等同步场景 |

### 连接器

| 命令 | 功能 | 示例 |
|------|------|------|
| `connector info` | 查询连接器详情 | `lingtong-cli connector info --connector kmerp` |
| `connector list` | 列出连接器账户 | `lingtong-cli connector list --app-id 165` |
| `connector category list` | 查询连接器类目 | `lingtong-cli connector category list --connector kmerp` |
| `connector account list` | 列出授权账户 | `lingtong-cli connector account list --connector kmerp --env prod` |
| `connector account create` | 创建授权账户 | `lingtong-cli connector account create --connector kmerp --name 测试 --data '{"appKey":"x"}'` |
| `connector account verify` | 验证授权账户 | `lingtong-cli connector account verify --connector kmerp --account-id 123` |
| `connector check-auth` | 检查工作流/场景/连接器授权 | `lingtong-cli connector check-auth --workflow-id 947` |
| `connector methods` | 列出可调用方法 | `lingtong-cli connector methods --connector kmerp --app-id 165 --auth-account-id 296` |
| `connector schema` | 查询方法请求/响应字段 | `lingtong-cli connector schema --connector kmerp --method erp.warehouse.list.query --auth-account-id 296` |
| `connector invoke` | 通过通用工作流调用连接器方法 | `lingtong-cli connector invoke --connector kmerp --method erp.warehouse.list.query --auth-account 广州力人服饰 --env test --body '{}'` |
| `connector cache list` | 查看 invoke 缓存工作流 | `lingtong-cli connector cache list` |
| `connector cache clear` | 清理 invoke 缓存 | `lingtong-cli connector cache clear --key test` |

`connector invoke` 会按环境创建或复用一个由 CLI 管理的通用工作流，并将 `connector`、`method`、`authAccount`、`body` 作为运行参数传入。若平台中的自动工作流被手动修改，可使用 `--force` 重建。

### 连接器工厂

| 命令 | 功能 | 示例 |
|------|------|------|
| `factory list` | 列出工厂 | `lingtong-cli factory list --page 1 --page-size 20` |
| `factory get` | 查询工厂详情 | `lingtong-cli factory get --id 12` |
| `factory save` | 创建或更新工厂 | `lingtong-cli factory save --data '{"name":"my-connector","connector":"mycon"}'` |
| `factory delete` | 删除工厂 | `lingtong-cli factory delete --id 12 --yes` |
| `factory http list` | 列出 HTTP 接口 | `lingtong-cli factory http list --factory-id 12` |
| `factory http get` | 查询 HTTP 接口详情 | `lingtong-cli factory http get --id 88` |
| `factory http run` | 测试运行 HTTP 接口 | `lingtong-cli factory http run --data '{"id":88,"params":{"page":1}}'` |
| `factory http delete` | 删除 HTTP 接口 | `lingtong-cli factory http delete --id 88 --yes` |
| `factory script list` | 列出脚本 | `lingtong-cli factory script list --factory-id 12` |
| `factory script get` | 查询脚本详情 | `lingtong-cli factory script get --id 55` |
| `factory script delete` | 删除脚本 | `lingtong-cli factory script delete --id 55 --yes` |

### 场景

| 命令 | 功能 | 示例 |
|------|------|------|
| `scene list` | 列出场景 | `lingtong-cli scene list --page 1 --page-size 20 --app-id 165` |
| `scene info` | 查询场景详情 | `lingtong-cli scene info --scene-id 123` |
| `scene create` | 创建最小草稿场景 | `lingtong-cli scene create --name 订单同步 --description 同步订单` |
| `scene update` | 更新场景 JSON | `lingtong-cli scene update --data '{"id":123,"name":"renamed"}'` |
| `scene copy` | 复制场景 | `lingtong-cli scene copy --scene-id 123` |
| `scene open` | 开关场景 | `lingtong-cli scene open --scene-id 123` |
| `scene publish` | 发布场景版本 | `lingtong-cli scene publish --scene-id 123` |
| `scene version list` | 查看版本历史 | `lingtong-cli scene version list --scene-id 123` |
| `scene trigger get` | 查询触发条件 | `lingtong-cli scene trigger get --scene-id 123` |
| `scene trigger save` | 保存触发条件 | `lingtong-cli scene trigger save --data '{"sceneId":123,"conditions":[]}'` |
| `scene field-mapping list` | 查询字段映射值 | `lingtong-cli scene field-mapping list --scene-id 123` |
| `scene field-mapping execute` | 调试执行字段映射 | `lingtong-cli scene field-mapping execute --data '{"sceneId":123,"record":{}}'` |
| `scene delete` | 删除场景 | `lingtong-cli scene delete --scene-id 123 --yes` |

`scene create` 只创建 name/description 级别的草稿场景；完整场景配置通常需要结合连接器、模型、字段映射和工作流能力。

### 工作流

| 命令 | 功能 | 示例 |
|------|------|------|
| `workflow list` | 列出工作流 | `lingtong-cli workflow list --app-id 165` |
| `workflow info` | 查询详情 | `lingtong-cli workflow info --workflow-id 123` |
| `workflow create` | 创建工作流 | `lingtong-cli workflow create --name MyWF --dsl-file wf.json` |
| `workflow update` | 更新工作流 | `lingtong-cli workflow update --workflow-id 100 --dsl-file wf.json --forced` |
| `workflow delete` | 删除工作流 | `lingtong-cli workflow delete --workflow-id 100 --confirm` |
| `workflow validate` | 验证 DSL | `lingtong-cli workflow validate --dsl-file wf.json --strict` |
| `workflow test run` | 测试运行 | `lingtong-cli workflow test run --workflow-id 100 --params '{"k":"v"}'` |
| `workflow publish` | 发布版本 | `lingtong-cli workflow publish --workflow-id 100 --version v1.0.0` |
| `workflow versions` | 列出版本 | `lingtong-cli workflow versions --workflow-id 100` |
| `workflow version rollback` | 回滚版本 | `lingtong-cli workflow version rollback --workflow-id 100 --version v1.0.0 --dry-run` |
| `workflow api-enable` | 开启 Open API | `lingtong-cli workflow api-enable --workflow-id 100` |
| `workflow api-disable` | 关闭 Open API | `lingtong-cli workflow api-disable --workflow-id 100` |
| `workflow api-test` | 测试 Open API | `lingtong-cli workflow api-test --app-tag abc123 --params '{"k":"v"}'` |
| `workflow execute` | 执行工作流 | `lingtong-cli workflow execute --workflow-id 123 --wait --params '{"k":"v"}'` |
| `workflow logs` | 查询日志 | `lingtong-cli workflow logs --receipt-id abc123` |
| `workflow template list` | 列出模板 | `lingtong-cli workflow template list` |
| `workflow template show` | 查看模板 | `lingtong-cli workflow template show --name order_sync` |
| `workflow template use` | 生成模板 DSL | `lingtong-cli workflow template use --name approval --output wf.json` |
| `workflow doc generate` | 生成工作流文档 | `lingtong-cli workflow doc generate --workflow-id 100 --output-file docs.md` |
| `workflow dependency list` | 分析依赖 | `lingtong-cli workflow dependency list --dsl-file wf.json` |

内置工作流模板：

| 模板 | 节点数 | 描述 |
|------|--------|------|
| `simple` | 2 | 简单开始到结束 |
| `connector` | 3 | 包含连接器节点 |
| `order_sync` | 5 | ERP 订单同步 |
| `approval` | 4 | 条件分支审批 |
| `data_pipeline` | 5 | 数据转换管道 |
| `api_wrapper` | 4 | 将连接器封装为 API |

### 表格与数据

| 命令 | 功能 | 示例 |
|------|------|------|
| `table list` | 列出表格 | `lingtong-cli table list --app-id 165` |
| `table create` | 创建表格 | `lingtong-cli table create --app-id 165 --name 客户资料 --source 1 --type 1` |
| `table update` | 更新表格配置 | `lingtong-cli table update --id 1568 --open-connector 1` |
| `table schema query` | 查询 Schema | `lingtong-cli table schema query --basic-data-id 123` |
| `table schema update` | 更新 Schema | `lingtong-cli table schema update --schema-id 2189 --basic-data-id 1553 --columns-schema '[{"id":1,"title":"Name","type":"text","key":"name","show":true}]'` |
| `table data query` | 查询数据 | `lingtong-cli table data query --basic-data-id 123 --filter '{"1":"系统订单"}'` |
| `table data create` | 创建记录 | `lingtong-cli table data create --basic-data-id 123 --schema-id 1 --data '{"0":"val"}'` |
| `table data update` | 更新单条记录 | `lingtong-cli table data update --basic-data-id 123 --schema-id 1 --id 33832272 --data '{"1":"new"}'` |
| `table data batch-update` | 批量更新 | `lingtong-cli table data batch-update --basic-data-id 123 --schema-id 1 --records '[{"id":1,"data":{"1":"v"}}]'` |
| `table data count` | 统计记录数 | `lingtong-cli table data count --schema-id 2546 --version 1` |
| `table data delete` | 删除单条记录 | `lingtong-cli table data delete --schema-id 2546 --id 33832272 --yes` |
| `table data batch-delete` | 批量删除 | `lingtong-cli table data batch-delete --schema-id 2546 --ids 1,2,3 --yes` |

`basicdata` 独立命令已废弃，常规基础资料/表格数据操作统一使用 `table data`。OpenAPI 中仍存在 `/basicdata/*` 模块，长尾接口可使用 `lingtong-cli service basicdata ...`。

### 表格视图与透视表

| 命令 | 功能 | 示例 |
|------|------|------|
| `table view save` | 创建视图 | `lingtong-cli table view save --schema-id 2367 --name 按类型分组 --group-column 1` |
| `table view list` | 列出视图 | `lingtong-cli table view list --schema-id 2367` |
| `table view update` | 更新视图 | `lingtong-cli table view update --id 621 --schema-id 2367 --name 新名称` |
| `table view delete` | 删除视图 | `lingtong-cli table view delete --id 621` |
| `table view group-data` | 获取分组数据 | `lingtong-cli table view group-data --view-id 621` |
| `table view merits` | 获取统计指标 | `lingtong-cli table view merits --schema-id 2367 --view-id 621` |
| `table pivot query` | 查询透视表数据 | `lingtong-cli table pivot query --table-id 1568 --view-id 621` |
| `table pivot config` | 读取透视表配置 | `lingtong-cli table pivot config --table-id 1568 --view-id 621` |
| `table pivot config-save` | 保存透视表配置 | `lingtong-cli table pivot config-save --table-id 1568 --view-id 621 --business-id 1568 --business-type 2 --config '{"dimensions":[],"measures":[]}'` |

### 模型元数据

| 命令 | 功能 | 示例 |
|------|------|------|
| `model interface list` | 查询接口模型 | `lingtong-cli model interface list --connector kmerp --filter-model-type query` |
| `model domain get` | 查询领域模型 | `lingtong-cli model domain get --connector kmerp --business order` |
| `model dynamic view` | 查询动态模型视图 | `lingtong-cli model dynamic view --connector kmerp --auth-account-id 296 --model-name SalesOrder` |

### API 自省、自动命令与原始 API

| 命令 | 功能 | 示例 |
|------|------|------|
| `schema list` | 列出 OpenAPI 路径 | `lingtong-cli schema list --format table` |
| `schema path` | 查看单个路径 Schema | `lingtong-cli schema path /scene/list` |
| `schema module` | 查看模块接口 | `lingtong-cli schema module scene` |
| `schema search` | 搜索接口 | `lingtong-cli schema search 场景` |
| `service <module> <operation>` | 调用自动生成命令 | `lingtong-cli service scene list --pageNum 1 --pageSize 10` |
| `api <method> <path>` | 调用任意 API | `lingtong-cli api GET /gw/ai/connector/info?connector=kmerp` |

`service` 命令从内置 OpenAPI 规格生成，当前覆盖约 487 个 operation；具体数量以 `schema list` 输出为准。

### 工具与运维命令

| 命令 | 功能 | 示例 |
|------|------|------|
| `doctor` | 检查版本、配置、认证、连通性 | `lingtong-cli doctor` |
| `update` | 检查版本并给出更新命令 | `lingtong-cli update --check` |
| `completion` | 生成 shell completion | `lingtong-cli completion zsh > ~/.zsh/completions/_lingtong-cli` |
| `skills list` | 列出内置 Skills | `lingtong-cli skills list` |
| `skills install` | 安装 Skills 到 AI 编辑器 | `lingtong-cli skills install --scope project` |
| `skills status` | 查看安装状态 | `lingtong-cli skills status` |
| `skills uninstall` | 卸载 Skills | `lingtong-cli skills uninstall --editor cursor` |

## Agent Skills

当前二进制内置 11 个有效 Skills。可通过 `lingtong-cli skills list` 查看实时清单。

| Skill | 说明 |
|-------|------|
| `lingtong-shared` | 配置、认证、Profile、多身份、输出契约、schema、update、skills install、安全规则 |
| `lingtong-cli-app` | 应用导出/导入/验证/脚手架/diff/应用内场景管理 |
| `lingtong-connector` | 连接器信息、类目、账户授权、methods/schema/invoke/cache |
| `lingtong-factory` | 连接器工厂、HTTP 接口、脚本管理 |
| `lingtong-scene` | 场景查询、草稿创建、生命周期操作、触发条件、字段映射调试 |
| `lingtong-workflow` | 工作流 DSL、节点编排、验证、测试、发布、执行、Open API |
| `lingtong-table` | 表格 CRUD、Schema、记录 CRUD、视图、分组、统计 |
| `lingtong-pivot-table` | 透视表查询、配置、维度/度量聚合 |
| `lingtong-model` | 接口模型、领域模型、动态模型视图 |
| `lingtong-service` | OpenAPI 自动命令层，覆盖长尾 API |
| `lingtong-script` | ES5.1/Nashorn 脚本、InfInvoker、AppInvoker、context API |

`skills/lingtong-basicdata/` 是历史遗留目录，不再作为有效 Skill 内置。基础资料操作已合并到 `lingtong-table`。

## 安装 Skills 到 AI 编辑器

支持的编辑器：

| 编辑器 | Skills 目录 | Agent 指引文件 |
|--------|-------------|----------------|
| Claude Code | `~/.claude/skills` / `.claude/skills` | `CLAUDE.md` |
| OpenCode | `~/.config/opencode/skills` / `.opencode/skills` | `AGENTS.md` |
| Qoder | `~/.qoder/skills` / `.qoder/skills` | `AGENTS.md` |
| Cursor | `~/.cursor/skills` / `.cursor/skills` | `AGENTS.md` |
| Trae | `~/.trae/skills` / `.trae/skills` | `AGENTS.md` |
| Codex | 无独立 Skills 目录 | `AGENTS.md`、`~/.codex/AGENTS.md` |

安装行为：

- 自动探测：未指定 `--scope` 时，项目内存在对应编辑器目录则安装到项目级，用户主目录存在对应编辑器目录则安装到全局。
- 幂等更新：`AGENTS.md` / `CLAUDE.md` 中的绫通内容写入受管标记块，重复安装只替换该块。
- 可卸载：`skills uninstall` 会删除内置 skill 目录并清除受管块。

也可使用封装脚本：

```bash
./scripts/install-skills.sh --editor opencode --scope project
```

## 安全与风险提示

本工具可供 AI Agent 调用以自动化操作绫通平台。认证后，Agent 将以你的用户身份执行操作，可能造成敏感数据泄露、误删除、误发布、越权变更等风险。

请遵守以下规则：

- 不要输出 Token 明文，不要把 Token 写入仓库文件。
- 优先使用 `--dry-run` 预览导入、导出、回滚、脚手架等操作。
- 删除、批量删除、删除工厂/脚本/场景前必须确认业务影响，并显式传 `--yes`。
- AI Agent 自动化调用建议使用 `--envelope` 获取结构化错误，并按 `.ok` / `.error.type` 判定结果。
- 不要随意关闭默认安全保护；放宽限制后的后果由操作者自行承担。

## 开发

```bash
# 构建
make build

# 静态检查
make vet

# 单元测试
make unit-test

# 完整测试
make test

# 生成命令文档
make gen-docs

# 检查命令文档是否最新
make check-docs

# 安装到系统
make install

# 安装内置 Skills
make install-skills ARGS="--scope project"

# 清理
make clean
```

## 文档维护口径

- 主 README 以 `./lingtong-cli --help`、各子命令 `--help`、`lingtong-cli skills list` 和近期 git 提交为准。
- Skill 目录的 README 负责导航和快速入口，`SKILL.md` 负责 Agent 执行指令，`references/` 负责深文档。
- 新增命令后应同步更新：主 README、对应 `skills/<name>/README.md`、对应 `SKILL.md`，必要时更新 reference。
- `basicdata` 口径统一为：独立命令废弃，常规操作走 `table data`，长尾 OpenAPI 走 `service basicdata`。

## 许可证

本项目基于 MIT 许可证开源。
