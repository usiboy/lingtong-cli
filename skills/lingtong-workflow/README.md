# lingtong-workflow 技能

`lingtong-workflow` 覆盖绫通工作流全生命周期：DSL 设计、节点编排、创建、更新、验证、测试、发布、回滚、执行、日志、Open API、模板、文档生成和依赖分析。

## 何时使用

- 需要设计或修改工作流 DSL
- 需要创建、更新、删除、发布或回滚工作流
- 需要验证 DSL、测试执行、查询执行日志
- 需要把工作流开放为 API
- 需要分析工作流依赖或生成工作流文档
- 需要编写工作流脚本节点、连接器调用节点或数据管道

## 目录结构

```
lingtong-workflow/
├── SKILL.md
├── README.md
└── references/
    ├── node-details.md
    ├── dsl-specification.md
    ├── script-guide.md
    ├── connector-invoke.md
    ├── template-guide.md
    ├── testing-guide.md
    ├── error-codes.md
    └── examples/
        ├── simple-workflow.md
        ├── connector-api.md
        ├── data-pipeline.md
        └── outstock-sync.md
```

## 快速开始

```bash
# 使用模板创建
lingtong-cli workflow create --name "订单同步" --template order_sync

# 使用 DSL 文件创建或更新
lingtong-cli workflow create --name "自定义流程" --dsl-file workflow.json
lingtong-cli workflow update --workflow-id 100 --dsl-file workflow.json --forced

# 验证和测试
lingtong-cli workflow validate --dsl-file workflow.json --strict
lingtong-cli workflow test run --workflow-id 100 --params '{"key":"value"}' --verbose

# 发布与执行
lingtong-cli workflow publish --workflow-id 100 --version v1.0.0
lingtong-cli workflow execute --workflow-id 100 --wait --params '{"key":"value"}'
```

## 命令清单

| 目标 | 命令 |
|------|------|
| 列出工作流 | `lingtong-cli workflow list --app-id <id>` |
| 查询详情 | `lingtong-cli workflow info --workflow-id <id>` |
| 创建工作流 | `lingtong-cli workflow create --name <name> --dsl-file workflow.json` |
| 更新工作流 | `lingtong-cli workflow update --workflow-id <id> --dsl-file workflow.json --forced` |
| 删除工作流 | `lingtong-cli workflow delete --workflow-id <id> --confirm` |
| 验证 DSL | `lingtong-cli workflow validate --dsl-file workflow.json --strict` |
| 测试运行 | `lingtong-cli workflow test run --workflow-id <id> --params '<json>'` |
| 发布版本 | `lingtong-cli workflow publish --workflow-id <id> --version v1.0.0` |
| 列出版本 | `lingtong-cli workflow versions --workflow-id <id>` |
| 回滚版本 | `lingtong-cli workflow version rollback --workflow-id <id> --version v1.0.0 --dry-run` |
| 启用 Open API | `lingtong-cli workflow api-enable --workflow-id <id>` |
| 禁用 Open API | `lingtong-cli workflow api-disable --workflow-id <id>` |
| 测试 Open API | `lingtong-cli workflow api-test --app-tag <tag> --params '<json>'` |
| 执行工作流 | `lingtong-cli workflow execute --workflow-id <id> --wait` |
| 查询日志 | `lingtong-cli workflow logs --receipt-id <id>` |
| 模板列表 | `lingtong-cli workflow template list` |
| 使用模板 | `lingtong-cli workflow template use --name approval --output wf.json` |
| 生成文档 | `lingtong-cli workflow doc generate --workflow-id <id> --output-file docs.md` |
| 分析依赖 | `lingtong-cli workflow dependency list --dsl-file workflow.json` |

## 内置模板

| 模板 | 说明 |
|------|------|
| `simple` | 开始到结束的最小工作流 |
| `connector` | 包含连接器节点 |
| `order_sync` | ERP 订单同步示例 |
| `approval` | 条件分支审批流程 |
| `data_pipeline` | 数据转换管道 |
| `api_wrapper` | 将连接器封装为 Open API |

## DSL 维护重点

- `w_start` 定义入口变量。
- `w_script` 脚本节点必须包含 `assertConfig`，推荐 `{"assertType":"throwException"}`。
- `w_end` 通过 `outputVariables` 输出结果。
- 节点间变量引用使用 `$nodeId.variable`。
- 脚本运行环境为 ES5.1/Nashorn，不能使用现代 JS 语法。

## Reference 导航

| 文档 | 说明 |
|------|------|
| [SKILL.md](SKILL.md) | Agent 执行指令与快速路由 |
| [references/node-details.md](references/node-details.md) | 12 类节点配置 |
| [references/dsl-specification.md](references/dsl-specification.md) | DSL 结构与变量规则 |
| [references/script-guide.md](references/script-guide.md) | 工作流脚本节点指南 |
| [references/connector-invoke.md](references/connector-invoke.md) | AppInvoker/连接器调用模式 |
| [references/template-guide.md](references/template-guide.md) | 模板使用说明 |
| [references/testing-guide.md](references/testing-guide.md) | 测试与调试指南 |
| [references/error-codes.md](references/error-codes.md) | 常见错误码 |
| [references/examples/simple-workflow.md](references/examples/simple-workflow.md) | 最小工作流示例 |
| [references/examples/connector-api.md](references/examples/connector-api.md) | 连接器 API 化示例 |
| [references/examples/data-pipeline.md](references/examples/data-pipeline.md) | 数据管道示例 |
| [references/examples/outstock-sync.md](references/examples/outstock-sync.md) | 出库同步场景示例 |

## 相关技能

| 技能 | 说明 |
|------|------|
| `lingtong-script` | 脚本节点、InfInvoker、AppInvoker |
| `lingtong-connector` | 连接器方法、Schema 与 invoke |
| `lingtong-table` | 表格查询、统计与记录操作 |
| `lingtong-scene` | 场景生命周期与字段映射 |

## 维护说明

- 节点配置以 `references/node-details.md` 为准。
- DSL 结构以 `references/dsl-specification.md` 为准。
- 新增工作流示例时必须同步本 README 的目录树和 Reference 表。
