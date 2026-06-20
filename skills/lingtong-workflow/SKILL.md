---
name: lingtong-workflow
version: 3.0.0
description: "绫通工作流全生命周期管理：设计、创建、验证、测试、发布、监控。当用户需要编排业务流程、管理工作流DSL、执行工作流测试时使用。关键词：workflow、工作流、DSL、节点、编排、发布。"
metadata:
  requires:
    bins: ["lingtong-cli"]
  cliHelp: "lingtong-cli workflow --help"
---

# lingtong-workflow

## 何时使用

使用本 skill：

- 用户要创建、编辑、验证工作流 DSL
- 用户要发布、测试、监控工作流
- 用户要理解工作流节点类型和配置
- 用户要使用模板快速创建工作流
- 用户要管理工作流版本和回滚

不要使用本 skill：

- 只是查询连接器信息或账户，转 `lingtong-connector`
- 只是管理场景（scene），转 `lingtong-scene`
- 只是操作表格数据，转 `lingtong-table`

## 使用边界

- 工作流 DSL 编辑使用 `lingtong-cli workflow` 命令
- 通用连接器调用使用 `lingtong-cli connector invoke`，不直接编辑工作流
- 节点配置以 references/node-details.md 为准
- DSL 格式以 references/dsl-specification.md 为准
- 脚本编写以 references/script-guide.md 为准

## 快速路由

| 用户目标 | 优先命令 | 何时读 reference |
|---|---|---|
| 查看工作流列表 | `workflow list --app-id <id>` | - |
| 查看工作流详情 | `workflow info --workflow-id <id>` | - |
| 创建工作流 | `workflow create --name <name> --template <tpl>` | 读 [template-guide.md](references/template-guide.md) 了解模板 |
| 更新工作流 DSL | `workflow update --workflow-id <id> --dsl-file <file>` | 读 [dsl-specification.md](references/dsl-specification.md) |
| 验证 DSL | `workflow validate --dsl-file <file>` | 读 [dsl-specification.md](references/dsl-specification.md) |
| 测试工作流 | `workflow test run --workflow-id <id>` | 读 [testing-guide.md](references/testing-guide.md) |
| 发布工作流 | `workflow publish --workflow-id <id>` | - |
| 查看版本历史 | `workflow versions --workflow-id <id>` | - |
| 启用 Open API | `workflow api-enable --workflow-id <id>` | - |
| 调用连接器接口 | `connector invoke --connector <name> --method <m>` | 读 [connector-invoke.md](references/connector-invoke.md) |
| 理解节点配置 | - | 读 [node-details.md](references/node-details.md) |
| 编写脚本节点 | - | 读 [script-guide.md](references/script-guide.md) |

## 工作流心智模型

- 工作流由**节点（Node）**和**边（Edge）**组成有向无环图（DAG）
- 每个工作流必须有且仅有一个 `w_start` 节点
- 至少有一个 `w_end` 节点
- 数据从上游流向下游，不可逆向
- 变量引用语法：`$nodeId.variableName`
- 工作流支持 12 种节点类型，详见 [node-details.md](references/node-details.md)

## 节点类型概览

| 类型 | 名称 | 用途 | 详细配置 |
|------|------|------|----------|
| `w_start` | 开始节点 | 定义输入参数 | [node-details.md](references/node-details.md#w_start) |
| `w_end` | 结束节点 | 定义输出结果 | [node-details.md](references/node-details.md#w_end) |
| `w_connector` | 连接器节点 | 调用外部 API | [node-details.md](references/node-details.md#w_connector) |
| `w_script` | 脚本节点 | 执行 JavaScript | [node-details.md](references/node-details.md#w_script) |
| `w_if` | 条件分支 | 二元条件判断 | [node-details.md](references/node-details.md#w_if) |
| `w_switch` | 选择分支 | 多路条件匹配 | [node-details.md](references/node-details.md#w_switch) |
| `w_cycle` | 循环节点 | 重复执行 | [node-details.md](references/node-details.md#w_cycle) |
| `w_modePipe` | 数据管道 | 批量数据同步 | [node-details.md](references/node-details.md#w_modepipe) |
| `w_dataSplit` | 数据拆分 | 数组遍历处理 | [node-details.md](references/node-details.md#w_datasplit) |
| `w_joinPipeline` | 聚合节点 | 合并分支结果 | [node-details.md](references/node-details.md#w_joinpipeline) |
| `w_dataPush` | 数据推送 | 带重试推送 | [node-details.md](references/node-details.md#w_datapush) |
| `w_pushRecord` | 推送记录 | 查询推送历史 | [node-details.md](references/node-details.md#w_pushrecord) |

## DSL 编排规则

### 变量引用语法
```
$nodeId.variableName              # 从节点获取变量
$w_start_first.orderId            # 从开始节点获取 orderId
$w_connector_1k65t.response.list  # 从连接器响应提取 list
```

### 边连接规则
- `w_start`: 一条出边
- `w_connector`: 可有两出边 (true/false)
- `w_if`: 多出边 (true/false)
- `w_dataSplit`: 一条出边
- `w_end`: 不能有出边

### 完整 DSL 结构
```json
{
  "environment": "formal",
  "name": "工作流名称",
  "viewport": {"x": 52, "y": 11, "zoom": 1},
  "nodes": [...],
  "edges": [...]
}
```

详细规范见 [dsl-specification.md](references/dsl-specification.md)

## 命令路由

### 基础管理
```bash
# 列出工作流
lingtong-cli workflow list --app-id 165

# 查询详情
lingtong-cli workflow info --workflow-id 955

# 执行日志
lingtong-cli workflow logs --receipt-id abc123
```

### 创建与更新
```bash
# 使用模板创建
lingtong-cli workflow create --name "审批流程" --template approval

# 使用 DSL 文件创建
lingtong-cli workflow create --name "自定义" --dsl-file workflow.json

# 更新 DSL
lingtong-cli workflow update --workflow-id 955 --dsl-file updated.json
```

### 验证与测试
```bash
# 验证 DSL
lingtong-cli workflow validate --dsl-file workflow.json

# 测试执行
lingtong-cli workflow test run --workflow-id 955 --params '{"key":"value"}'
```

### 发布与版本
```bash
# 发布
lingtong-cli workflow publish --workflow-id 955 --version "v1.0.0"

# 版本历史
lingtong-cli workflow versions --workflow-id 955

# 回滚
lingtong-cli workflow version rollback --workflow-id 955 --version "v1.0.0"
```

### Open API
```bash
# 启用
lingtong-cli workflow api-enable --workflow-id 956

# 禁用
lingtong-cli workflow api-disable --workflow-id 956

# 测试
lingtong-cli workflow api-test --app-tag <tag> --params '{"key":"value"}'
```

## 常见恢复

| 错误 / 现象 | 恢复动作 |
|---|---|
| 验证失败: 缺少开始节点 | 添加 `w_start` 节点 |
| 验证失败: 节点类型无效 | 检查 12 种合法节点类型 |
| 变量引用错误 | 检查 `$nodeId.variable` 格式 |
| API 返回"未发布" | 先执行 `workflow publish` |
| 脚本执行失败: `context 信息不完整` | 确保设置 `context.put("_env", env)` |
| 连接器调用失败: `授权信息不存在` | 检查 authAccount 名称是否正确 |
| 回滚失败 | 使用 `versions` 查看可用版本 |

## 示例

- [简单工作流](references/examples/simple-workflow.md) - 开始→结束基础流程
- [连接器 API 化](references/examples/connector-api.md) - 将连接器接口暴露为 API
- [数据管道](references/examples/data-pipeline.md) - 批量数据同步流程
- [通用连接器调用](references/connector-invoke.md) - CLI 自动化调用模式

## 保留 Reference

- [node-details.md](references/node-details.md) - 12 种节点详细配置
- [dsl-specification.md](references/dsl-specification.md) - DSL 完整规范
- [script-guide.md](references/script-guide.md) - 脚本编写指南（ES5.1）
- [connector-invoke.md](references/connector-invoke.md) - 通用连接器调用模式
- [template-guide.md](references/template-guide.md) - 模板使用指南
- [testing-guide.md](references/testing-guide.md) - 测试指南
- [error-codes.md](references/error-codes.md) - 错误码说明
