# 通用连接器调用模式

本文档说明如何通过 CLI 命令调用连接器接口，无需手动编辑工作流。

## 概述

`lingtong-cli connector invoke` 命令自动创建和管理一个通用工作流，实现任意连接器接口的调用。

## 工作原理

```
CLI 命令 → 通用工作流 (Start → Script → End) → Open API → 连接器接口
```

- **单一工作流复用**: 每个环境（test/prod）只创建一个工作流
- **参数外部化**: 所有参数通过 Start 节点传入
- **自动缓存**: 首次调用自动创建，后续复用

## 命令格式

```bash
lingtong-cli connector invoke \
  --connector <connector_name> \
  --method <method_name> \
  --auth-account <account_name> \
  [--env test|prod] \
  [--body '<json>'] \
  [--force]
```

## 参数说明

| 参数 | 必需 | 说明 |
|------|------|------|
| `--connector` | 是 | 连接器名称（如 `kmerp`） |
| `--method` | 是 | 方法名（如 `erp.warehouse.list.query`） |
| `--auth-account` | 是 | 认证账户名称 |
| `--env` | 否 | 环境，默认 `prod` |
| `--body` | 否 | 请求参数 JSON |
| `--force` | 否 | 强制重建工作流 |

## 使用示例

### 基础调用

```bash
# 查询仓库列表
lingtong-cli connector invoke \
  --connector kmerp \
  --method "erp.warehouse.list.query" \
  --auth-account "广州力人服饰" \
  --env test \
  --body '{}' \
  --format pretty
```

### 带参数调用

```bash
# 分页查询
lingtong-cli connector invoke \
  --connector kmerp \
  --method "erp.shop.list.query" \
  --auth-account "广州力人服饰" \
  --env test \
  --body '{"pageNo":1,"pageSize":10}' \
  --format pretty
```

### 查看接口参数

```bash
# 查看接口定义
lingtong-cli connector schema \
  --connector kmerp \
  --method "erp.warehouse.list.query" \
  --auth-account-id 296 \
  --format pretty
```

## 工作流结构

通用工作流采用 Start → Script → End 三节点结构：

### Start 节点

定义 5 个输入参数：

```json
{
  "id": "w_start_first",
  "type": "w_start",
  "data": {
    "outputVariables": [
      {"variable": "connector", "variableAttr": {"dataType": "text", "required": true}},
      {"variable": "method", "variableAttr": {"dataType": "text", "required": true}},
      {"variable": "authAccount", "variableAttr": {"dataType": "text", "required": true}},
      {"variable": "body", "variableAttr": {"dataType": "json"}},
      {"variable": "env", "variableAttr": {"dataType": "text", "required": true}}
    ]
  }
}
```

### Script 节点

通过 `inputVariables` 引用上游参数：

```json
{
  "id": "w_script_invoke",
  "type": "w_script",
  "data": {
    "inputVariables": [
      {"variable": "connector", "variableAttr": {"value": "$w_start_first.connector"}},
      {"variable": "method", "variableAttr": {"value": "$w_start_first.method"}},
      {"variable": "authAccount", "variableAttr": {"value": "$w_start_first.authAccount"}},
      {"variable": "body", "variableAttr": {"value": "$w_start_first.body"}},
      {"variable": "env", "variableAttr": {"value": "$w_start_first.env"}}
    ],
    "scriptConfig": {
      "script": "var connector = context.get('connector');\nvar method = context.get('method');\nvar authAccount = context.get('authAccount');\nvar body = context.get('body') || {};\nvar env = context.get('env');\ncontext.put('_env', env);\nvar result = AppInvoker.invoke(context, method, authAccount, body);\nreturn result;"
    }
  }
}
```

### 关键配置

1. **inputVariables**: 必须声明，引用 `$w_start_first.xxx`
2. **context.put("_env", env)**: 调用 AppInvoker 前必须设置环境
3. **参数直接传递**: Open API 请求体顶层传递参数，不用 `args` 包裹

## 完整调用流程

```bash
# 1. 查看可用账户
lingtong-cli connector account list --connector kmerp

# 2. 查看可用方法
lingtong-cli connector methods --connector kmerp --app-id 165 --auth-account-id 296

# 3. 查看接口参数定义
lingtong-cli connector schema --connector kmerp --method "erp.warehouse.list.query" --auth-account-id 296

# 4. 调用接口
lingtong-cli connector invoke \
  --connector kmerp \
  --method "erp.warehouse.list.query" \
  --auth-account "广州力人服饰" \
  --env test \
  --body '{"pageSize":10}' \
  --format pretty
```

## 缓存管理

```bash
# 查看缓存
lingtong-cli connector cache list

# 清除缓存
lingtong-cli connector cache clear

# 强制重建
lingtong-cli connector invoke ... --force
```

## AppInvoker 请求/响应结构（易错点，已实测）

以快麦 ERP trade 类接口（`erp.trade.outstock.simple.query`、`erp.trade.list.query` 等）为例：

1. **请求参数传【扁平】，不要手动包 `requestBody`**。`AppInvoker.invoke` 的 body 运行时会自动包裹进
   `requestBody`；若自己再包一层会变成 `requestBody.requestBody`（双重包裹）→ ERP 收到空参数 →
   返回全 null（`TradeOutStockListQueryResponse(total=null,...)`）。在工作流脚本节点与 `connector invoke`
   两种路径下都应传扁平参数：
   ```js
   // ✅ 正确：扁平
   var body = {timeType:'created', startTime: s, endTime: e, pageSize: 100, pageNo: 1};
   var result = AppInvoker.invoke(context, 'erp.trade.outstock.simple.query', '广州力人服饰', body);
   // ❌ 错误：{requestBody:{...}} -> 双重包裹 -> 查不到数据
   ```
   用 `connector schema --connector <c> --method <m> --auth-account-id <id>` 看字段名（`requestFieldList`
   的 `fullPath` 形如 `requestBody.startTime`，确认要传哪些参数；但调用时传扁平即可，无需自己加 `requestBody` 前缀）。
2. **pageSize 不能太小**。kmerp 出库/订单接口对过小的 pageSize 会返回 null（实测 `pageSize=5` 失败、
   `20/50/100/200` 正常）。建议默认 100，并对 `<20` 的值兜底。
3. **成功响应直接是 `responseData` 对象**：字段在顶层（`{list, total, success, ...}`），读
   `result.list` / `result.total`（Nashorn 会按 JavaBean 规则把 `.list` 映射到 `getList()`），
   **不是** `result.result.list`。
4. **空/失败响应会抛异常**：无数据时连接器节点抛 `节点执行失败:XxxResponse(total=null,...)`，会中断整个工作流。
   查询脚本应 `try/catch`，仅对「空结果」(`total=null`) 降级为空数组、其它异常向上抛出：
   ```js
   var list = [], total = 0;
   try {
     var r = AppInvoker.invoke(context, method, account, body);
     if (r != null) { if (r.list != null) list = r.list; if (r.total != null) total = r.total; }
   } catch (e) { if (('' + e).indexOf('total=null') < 0) { throw e; } }
   return {list: list, total: total};
   ```

> 排错顺序：先用 `connector invoke` 单点验证接口能否返回数据（与工作流隔离），再排查工作流编排。
> 单节点级调试见 [testing-guide.md](testing-guide.md) 的「单节点调试 /workflow/debug/node/do」。

## 常见问题

| 问题 | 原因 | 解决方案 |
|------|------|----------|
| `context 信息不完整` | 未设置 `_env` | 脚本中添加 `context.put("_env", env)` |
| `connector,不能为空` | 参数被 `args` 包裹 | Open API 参数放顶层 |
| `args is not defined` | 脚本直接引用 `args` | 使用 `context.get("paramName")` |
| Script 拿不到参数 | 缺少 `inputVariables` | 配置 `$w_start_first.xxx` 引用 |
| 响应全 null（`total=null,list=null`） | 手动包了 `requestBody` 导致双重包裹，或 pageSize 过小 | 传扁平参数；pageSize 用 ≥20（建议 100） |
| 读不到响应数据 | 误用 `result.result.list` | 用 `result.list` / `result.total`（响应即 responseData） |
| 空数据导致整流程失败 | 连接器对空响应抛异常 | 查询脚本 `try/catch`，仅对 `total=null` 降级 |

## 相关文档

- [脚本编写指南](script-guide.md)
- [DSL 规范](dsl-specification.md)
- [节点详解](node-details.md)
