---
name: lingtong-script
version: 1.0.0
description: "绫通脚本编写技能。覆盖工作流脚本节点、场景字段函数、连接器字段函数三种执行场景。提供完整的内置对象参考（InfInvoker 表格函数、AppInvoker 连接器调用、log 日志）、ES5.1 语法规范和常见模式库。当用户需要编写绫通脚本、操作表格数据、调用连接器接口、进行数据统计分析时触发。关键词：script、脚本、InfInvoker、AppInvoker、表格函数、字段函数、ES5.1。"
metadata:
  requires:
    bins: ["lingtong-cli"]
  cliHelp: "lingtong-cli workflow --help"
---

# lingtong-script

## 何时使用

使用本 skill：

- 用户要编写工作流中的脚本节点代码
- 用户要编写场景字段函数中的脚本
- 用户要编写连接器字段函数中的脚本
- 用户要使用 InfInvoker 操作绫通表格数据（CRUD / ES 聚合）
- 用户要使用 AppInvoker 调用连接器接口
- 用户需要了解脚本语法限制和内置对象

不要使用本 skill：

- 只是管理工作流结构（创建/发布/更新），转 `lingtong-workflow`
- 只是查询连接器信息或账户，转 `lingtong-connector`
- 只是操作表格数据（通过 CLI），转 `lingtong-table`

## 使用边界

- 脚本编写基于 **ECMAScript 5.1**（Nashorn 引擎），不支持 ES6+ 语法
- 内置对象在不同场景中的可用性不同，参见下方矩阵
- InfInvoker 的 esAgg 方法要求表格开启高性能模式

## 脚本执行场景

绫通脚本在三种场景中执行，每种场景可用的内置对象不同：

| 场景 | 说明 | 触发方式 |
|------|------|----------|
| **工作流脚本节点** | 工作流中的 w_script 节点 | 工作流执行到脚本节点时 |
| **场景字段函数** | 场景配置中的字段映射函数 | 数据同步时字段映射阶段 |
| **连接器字段函数** | 连接器节点中的参数预处理脚本 | 连接器调用前参数构建阶段 |

### 内置对象可用性矩阵

| 内置对象 | 工作流脚本 | 场景字段函数 | 连接器字段函数 | 说明 |
|---------|:---:|:---:|:---:|------|
| `InfInvoker` | ✅ | ✅ | ✅ | 绫通表格 CRUD + ES 聚合 |
| `AppInvoker` | ✅ | ✅ | ❌ | 连接器接口调用 |
| `PushStateInvoker` | ✅ | ✅ | ❌ | 推送状态查询 |
| `FlowDataInvoker` | ✅ | ✅ | ❌ | 工作流数据查询 |
| `log` | ✅ | ✅ | ✅ | 日志输出 |
| `console` | ✅ | ✅ | ❌ | 控制台输出 |

## 快速路由

| 用户目标 | 命令/文档 | 参考文档 |
|----------|-----------|----------|
| 了解脚本语法限制 | — | [script-basics.md](references/script-basics.md) |
| 使用 context 对象 | — | [script-basics.md#context-api](references/script-basics.md#context-api) |
| 查询表格数据 | `InfInvoker.search()` | [inf-invoker-guide.md](references/inf-invoker-guide.md) |
| 按 ID 查询表格 | `InfInvoker.queryById()` | [inf-invoker-guide.md](references/inf-invoker-guide.md) |
| 插入/更新/删除表格 | `InfInvoker.insert()` 等 | [inf-invoker-guide.md](references/inf-invoker-guide.md) |
| ES 聚合统计分析 | `InfInvoker.esAgg()` | [inf-invoker-guide.md#esaggregations](references/inf-invoker-guide.md#esaggregations) |
| 调用连接器接口 | `AppInvoker.invoke()` | [script-basics.md#appinvoker](references/script-basics.md#appinvoker) |
| 表格操作完整示例 | — | [examples/table-operations.md](references/examples/table-operations.md) |

## 常见恢复

| 错误信息 | 原因 | 解决方案 |
|----------|------|----------|
| `context 不能为空` | 脚本缺少 context 参数 | 确保函数签名包含 context |
| `context 信息不完整` | context 缺少 _tenantId | 工作流/场景自动注入，检查节点配置 |
| `基础资料ID不能为空` | basicDataId 为 null | 传入正确的表格 ID |
| `该表格不包含字段：xxx` | 字段名错误 | 使用字段 key（如 `"1"`）而非中文名称 |
| `Unexpected token` | 使用了 ES6+ 语法 | 改用 ES5.1 语法（var 代替 let/const） |
| `esAgg 失败` | 表格未开启高性能 | 在表格设置中开启高性能模式 |
| `record.data` 返回 undefined | 使用 JS 语法访问 Java 对象 | 改用 `record.getData()` |
| `data['1']` 返回 undefined | 使用 JS 语法访问 Java Map | 改用 `data.get('1')` |
| 工作流返回空结果 | End 节点使用了 inputVariables | End 节点必须使用 `outputVariables` |

## 保留 Reference

| 文档 | 路径 | 说明 |
|------|------|------|
| 脚本基础 | [references/script-basics.md](references/script-basics.md) | ES5.1 语法、context API、内置对象 |
| InfInvoker 指南 | [references/inf-invoker-guide.md](references/inf-invoker-guide.md) | 表格函数完整参考 |
| 表格操作示例 | [references/examples/table-operations.md](references/examples/table-operations.md) | 基于出入库记录表的完整示例 |
