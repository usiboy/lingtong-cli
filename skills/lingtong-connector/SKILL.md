---
name: lingtong-connector
version: 1.1.0
description: "绫通连接器管理：查询连接器详情、类目列表、模型元数据、账户授权认证。当用户需要查询连接器配置、获取接口模型、了解数据结构、管理连接器账户授权时触发。关键词：connector、连接器、connector info、category list、model interface、account、authorization、verify。"
---

# lingtong-connector 技能

## 概述

本技能指导你如何通过 `lingtong-cli` 查询和管理绫通平台的连接器，包括连接器账户的授权认证管理。

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

## 最佳实践

1. **先查询连接器详情**，了解可用模型和类目
2. **使用 `--env prod`** 查询生产环境配置
3. **指定 `--auth-account-id`** 获取特定账户的扩展字段
4. **创建账户前先检查授权**，避免重复创建
5. **验证账户连接**，确保认证信息正确
6. **定期检查授权状态**，特别是在执行工作流前

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
