---
name: lingtong-connector
version: 2.0.0
description: "绫通连接器管理：查询连接器详情、类目列表、模型元数据、账户授权认证、连接器接口调用（invoke）。当用户需要查询连接器配置、获取接口模型、了解数据结构、管理连接器账户授权、调用连接器方法时触发。关键词：connector、连接器、invoke、connector invoke、methods、cache、AppInvoker。"
---

# lingtong-connector 技能

## 概述

本技能指导你如何通过 `lingtong-cli` 查询和管理绫通平台的连接器，包括连接器账户的授权认证管理，以及**通过通用工作流调用任意连接器接口**。

## 核心命令

### 查询连接器详情

```bash
lingtong-cli connector info --connector <connector_name> [--env test|prod] [--auth-account-id <id>]
```

**示例**:
```bash
# 查询金蝶云连接器
lingtong-cli connector info --connector kingdee

# 查询生产环境
lingtong-cli connector info --connector kmerp --env prod

# 指定认证账户
lingtong-cli connector info --connector kmerp --auth-account-id 123
```

### 查询连接器类目列表

```bash
lingtong-cli connector category list --connector <connector_name>
```

**示例**:
```bash
lingtong-cli connector category list --connector kmerp
```

## 连接器账户管理

### 列出连接器账户

```bash
lingtong-cli connector account list --connector <connector_name> [--env test|prod]
```

**示例**:
```bash
# 列出快麦ERP所有环境的账户
lingtong-cli connector account list --connector kmerp

# 仅列出生产环境
lingtong-cli connector account list --connector kmerp --env prod
```

### 验证连接器账户

```bash
lingtong-cli connector account verify --connector <connector_name> --account-id <id> [--env test|prod]
```

**示例**:
```bash
# 验证快麦ERP账户连接
lingtong-cli connector account verify --connector kmerp --account-id 123

# 验证飞书账户连接
lingtong-cli connector account verify --connector feishu --account-id 456
```

### 创建连接器账户

```bash
lingtong-cli connector account create --connector <connector_name> --name <account_name> --env <env> --data '<json>'
```

**示例**:
```bash
# 创建快麦ERP账户（AppKey/AppSecret认证）
lingtong-cli connector account create \
  --connector kmerp \
  --name "快麦测试账号" \
  --env test \
  --data '{"appKey":"xxx","appSecret":"yyy"}'

# 创建飞书账户（OAuth认证）
lingtong-cli connector account create \
  --connector feishu \
  --name "飞书多维表格" \
  --env prod \
  --data '{"app_id":"xxx","app_secret":"yyy","tenant_access_token":"zzz"}'
```

## 授权检查

### 检查工作流/场景的连接器授权状态

```bash
# 检查工作流使用的连接器是否已授权
lingtong-cli connector check-auth --workflow-id <workflow_id>

# 检查场景使用的连接器是否已授权
lingtong-cli connector check-auth --scene-id <scene_id>

# 检查特定连接器的授权状态
lingtong-cli connector check-auth --connector <connector_name>
```

**示例**:
```bash
# 检查工作流947的连接器授权
lingtong-cli connector check-auth --workflow-id 947

# 检查快麦ERP授权状态
lingtong-cli connector check-auth --connector kmerp
```

**输出示例**:
```
Checking connector authorization for Workflow ID: 947...

⚠ Some connectors are missing authorization:

  ✗ kmerp: No account configured
    → Run: lingtong-cli connector account create --connector kmerp --name "My Account" --data '{...}'

  ✗ feishu: No account configured
    → Run: lingtong-cli connector account create --connector feishu --name "My Account" --data '{...}'
```

## 快捷命令

```bash
# 快捷查询连接器
lingtong-cli +connector-info --connector kmerp
```

## 数据模型

### 连接器详情返回结构

```json
{
  "success": true,
  "data": {
    "id": 123,
    "name": "kmerp",
    "displayName": "管家婆ERP",
    "version": "1.0.0",
    "description": "...",
    "models": [...],
    "categories": [...]
  }
}
```

### 连接器账户返回结构

```json
{
  "success": true,
  "result": [
    {
      "id": 1,
      "connector": "kmerp",
      "name": "快麦测试账号",
      "env": "test",
      "open": 1,
      "accountRelationId": "xxx",
      "created": "2026-04-06T10:00:00"
    }
  ]
}
```

### 账户验证返回结构

```json
{
  "success": true,
  "result": {
    "success": true,
    "errMsg": null,
    "refreshResult": [],
    "webhookCode": "abc123"
  }
}
```

## 连接器接口调用

### 工作原理

`connector invoke` 通过自动创建和管理一个**通用工作流**（Start → Script → End）来调用任意连接器接口：
- **单一工作流复用**：每个环境（test/prod）只创建一个工作流，所有调用共享
- **参数外部化**：connector/method/authAccount/body 通过 Start 节点参数传入，工作流内不硬编码任何连接器
- **自动缓存**：首次调用自动创建工作流并缓存，后续调用直接复用

### 调用连接器方法

```bash
lingtong-cli connector invoke \
  --connector <connector_name> \
  --method <method_name> \
  --auth-account <account_name> \
  [--env test|prod] \
  [--body '<json>'] \
  [--force]
```

**示例**:
```bash
# 查询仓库列表
lingtong-cli connector invoke \
  --connector kmerp \
  --method "erp.warehouse.list.query" \
  --auth-account "广州力人服饰" \
  --env test \
  --body '{}'

# 带参数查询（分页）
lingtong-cli connector invoke \
  --connector kmerp \
  --method "erp.trade.list.query" \
  --auth-account "广州力人服饰" \
  --env test \
  --body '{"pageSize":10,"pageNo":1}'

# 强制重建通用工作流（忽略缓存）
lingtong-cli connector invoke \
  --connector kmerp \
  --method "erp.warehouse.list.query" \
  --auth-account "广州力人服饰" \
  --env test --force
```

**返回示例**:
```json
{
  "result": {
    "body": "{...}",
    "list": [...],
    "receiptStatus": "SUCCESS",
    "success": true,
    "total": 12
  },
  "_metadata": {
    "workflowId": 1010,
    "appTag": "b2nUqsNASykMgi",
    "connector": "kmerp",
    "method": "erp.warehouse.list.query",
    "authAccount": "广州力人服饰",
    "env": "test",
    "universal": true
  }
}
```

### 列出可用方法

```bash
lingtong-cli connector methods \
  --connector <connector_name> \
  --app-id <app_id> \
  --auth-account-id <account_id>
```

**示例**:
```bash
# 列出 kmerp 连接器的所有可用方法
lingtong-cli connector methods --connector kmerp --app-id 165 --auth-account-id 296
```

**返回示例**:
```json
{
  "count": 80,
  "methods": [
    {
      "method": "erp.warehouse.list.query",
      "title": "仓库查询",
      "business": "warehouse",
      "catId": "基础",
      "modelId": 119
    }
  ]
}
```

### 获取接口参数元数据

```bash
lingtong-cli connector schema \
  --connector <connector_name> \
  --method <method_name> \
  --auth-account-id <account_id>
```

**示例**:
```bash
# 获取仓库查询接口的请求和响应字段定义
lingtong-cli connector schema \
  --connector kmerp \
  --method "erp.warehouse.list.query" \
  --auth-account-id 296 \
  --format pretty
```

**返回示例**:
```json
{
  "requestFieldList": [
    {
      "name": "requestBody",
      "title": "仓库查询",
      "valueType": "object",
      "fullPath": "requestBody"
    },
    {
      "name": "code",
      "valueType": "string",
      "fullPath": "requestBody.code"
    },
    {
      "name": "name",
      "valueType": "string",
      "fullPath": "requestBody.name"
    },
    {
      "name": "id",
      "valueType": "integer",
      "fullPath": "requestBody.id"
    }
  ],
  "responseFieldList": [
    {
      "name": "contactPhone",
      "title": "联系电话",
      "valueType": "string"
    },
    {
      "name": "district",
      "title": "区",
      "valueType": "string"
    }
  ],
  "_metadata": {
    "connector": "kmerp",
    "method": "erp.warehouse.list.query",
    "modelId": 119
  }
}
```

**使用场景**:
- 在调用 `connector invoke` 前，先通过 `schema` 命令了解接口需要哪些参数
- 根据 `requestFieldList` 构造正确的 `--body` JSON
- 根据 `responseFieldList` 了解返回数据的结构

### 缓存管理

```bash
# 查看缓存的通用工作流
lingtong-cli connector cache list

# 清除所有缓存
lingtong-cli connector cache clear

# 清除特定环境的缓存
lingtong-cli connector cache clear --key "test"
```

## 最佳实践

1. **先查询连接器详情**，了解可用模型和类目
2. **使用 `connector methods`** 查看可用的方法名和 modelId
3. **使用 `connector schema`** 获取接口的请求参数定义，构造正确的 `--body`
4. **使用 `connector account list`** 确认账户名称正确
5. **首次调用用 `--force`** 确保工作流正常创建
6. **相同环境复用同一工作流**，无需为每个接口单独创建
7. **定期检查授权状态**，特别是在执行工作流前

## 完整调用流程示例

```bash
# 1. 查看可用账户
lingtong-cli connector account list --connector kmerp --format pretty

# 2. 查看可用方法
lingtong-cli connector methods --connector kmerp --app-id 165 --auth-account-id 296

# 3. 查看接口参数定义
lingtong-cli connector schema --connector kmerp --method "erp.warehouse.list.query" --auth-account-id 296

# 4. 根据 schema 构造参数并调用
lingtong-cli connector invoke \
  --connector kmerp \
  --method "erp.warehouse.list.query" \
  --auth-account "广州力人服饰" \
  --env test \
  --body '{"code":"A","name":"广州"}' \
  --format pretty
```

## 常见连接器认证字段

### 快麦ERP (kmerp)
- `url`: API地址
- `appKey`: 应用Key
- `appSecret`: 应用密钥
- `session`: 访问令牌（Token）
- `refreshToken`: 刷新令牌

### 飞书 (feishu)
- `app_id`: 应用ID
- `app_secret`: 应用密钥
- `tenant_access_token`: 租户访问令牌
- `refresh_token`: 刷新令牌

### 钉钉 (dingtalk)
- `app_key`: 应用Key
- `app_secret`: 应用密钥
- `access_token`: 访问令牌
