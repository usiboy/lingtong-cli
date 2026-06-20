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

## 常见问题

| 问题 | 原因 | 解决方案 |
|------|------|----------|
| `context 信息不完整` | 未设置 `_env` | 脚本中添加 `context.put("_env", env)` |
| `connector,不能为空` | 参数被 `args` 包裹 | Open API 参数放顶层 |
| `args is not defined` | 脚本直接引用 `args` | 使用 `context.get("paramName")` |
| Script 拿不到参数 | 缺少 `inputVariables` | 配置 `$w_start_first.xxx` 引用 |

## 相关文档

- [脚本编写指南](script-guide.md)
- [DSL 规范](dsl-specification.md)
- [节点详解](node-details.md)
