# lingtong-script 技能

`lingtong-script` 用于编写绫通平台中的 JavaScript 脚本，覆盖工作流脚本节点、场景字段函数、连接器字段函数，以及脚本中对表格和连接器的调用。

## 何时使用

- 编写工作流 `w_script` 节点脚本
- 编写场景字段映射函数、预处理函数或过滤逻辑
- 编写连接器请求前/响应后处理脚本
- 在脚本中使用 `InfInvoker` 操作表格数据
- 在脚本中使用 `AppInvoker` 调用连接器接口
- 排查 Nashorn / ES5.1 语法兼容问题

## 目录结构

```
lingtong-script/
├── SKILL.md
├── README.md
└── references/
    ├── inf-invoker-guide.md
    ├── script-basics.md
    └── examples/
        └── table-operations.md
```

## 执行环境

- JavaScript 标准：ECMAScript 5.1
- 引擎：Nashorn
- 不要使用 `let`、`const`、箭头函数、可选链、模板字符串、`Array.prototype.includes` 等现代语法。
- 推荐使用 `var`、普通函数、显式空值判断和 `JSON.stringify` / `JSON.parse`。

## 内置对象速查

| 对象 | 用途 |
|------|------|
| `context` | 读取/写入工作流或脚本上下文变量 |
| `log` | 输出调试日志 |
| `InfInvoker` | 表格/基础资料 CRUD 和统计查询 |
| `AppInvoker` | 调用连接器接口 |

## 快速示例

### context

```javascript
var input = context.get("input");
context.put("output", input);
return input;
```

### InfInvoker

```javascript
var records = InfInvoker.query(context, {
  basicDataId: 1568,
  pageNum: 1,
  pageSize: 20
});
return records;
```

### AppInvoker

```javascript
var body = { pageSize: 10 };
var result = AppInvoker.invoke(context, "erp.warehouse.list.query", "广州力人服饰", body);
return result;
```

## 文档导航

| 文档 | 说明 |
|------|------|
| [SKILL.md](SKILL.md) | Agent 执行指令 |
| [references/script-basics.md](references/script-basics.md) | ES5.1 语法、context、日志、常见坑 |
| [references/inf-invoker-guide.md](references/inf-invoker-guide.md) | InfInvoker 表格函数完整指南 |
| [references/examples/table-operations.md](references/examples/table-operations.md) | 表格操作脚本示例 |
| [../lingtong-workflow/references/script-guide.md](../lingtong-workflow/references/script-guide.md) | 工作流脚本节点补充指南 |
| [../lingtong-workflow/references/connector-invoke.md](../lingtong-workflow/references/connector-invoke.md) | 连接器调用模式 |

## 与其他技能的分工

| 需求 | 推荐技能 |
|------|----------|
| 编写脚本本身 | `lingtong-script` |
| 设计工作流 DSL 和节点 | `lingtong-workflow` |
| 查询表格字段和记录 | `lingtong-table` |
| 查询连接器方法和字段 | `lingtong-connector` / `lingtong-model` |
| 管理连接器工厂脚本 | `lingtong-factory` |

## 常见坑

- 字段 ID 常是字符串数字，如 `data["1"]`。
- 动态模型字段与授权账号有关，换账号后要重新查 Schema。
- `AppInvoker.invoke` 的账号参数通常是授权账号名称，不是账号 ID。
- 工作流脚本节点的错误处理依赖 `assertConfig`，DSL 中不要漏写。
- 表格大批量操作先小样本验证，再扩大范围。

## 维护说明

- 本 README 负责脚本能力导航。
- 详细 API 以 reference 文档和后端实现为准。
- 新增脚本对象、函数或限制时，同步更新 `SKILL.md` 与对应 reference。
