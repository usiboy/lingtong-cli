---
name: lingtong-model
version: 1.0.0
description: "绫通模型元数据管理：查询接口模型、领域模型、动态模型视图。当用户需要了解连接器数据结构、查询模型定义、获取字段元数据时触发。关键词：model、元数据、interface list、domain、dynamic view、schema。"
---

# lingtong-model 技能

## 概述

本技能指导你如何通过 `lingtong-cli` 查询绫通平台的模型元数据。

## 核心命令

### 查询接口模型列表

```bash
lingtong-cli model interface list --connector <name> --filter-model-type <type> [--model-type <type>] [--auth-account-id <id>]
```

**示例**:
```bash
# 查询所有接口模型
lingtong-cli model interface list --connector kmerp --filter-model-type all

# 查询特定类型
lingtong-cli model interface list --connector kmerp --filter-model-type basic_domain
```

### 查询领域模型

```bash
lingtong-cli model domain get --connector <name> --business <name> [--auth-account-id <id>]
```

**示例**:
```bash
lingtong-cli model domain get --connector kmerp --business order
```

### 查询动态模型视图

```bash
lingtong-cli model dynamic view --connector <name> --auth-account-id <id> [--model-name <name>] [--business-object-name <name>]
```

**示例**:
```bash
lingtong-cli model dynamic view --connector kmerp --auth-account-id 123
```

## 数据模型

接口模型返回的典型数据结构：

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "Order",
      "displayName": "订单",
      "modelType": "basic_domain",
      "fields": [
        {
          "name": "orderNo",
          "displayName": "订单号",
          "type": "string",
          "required": true
        }
      ]
    }
  ]
}
```

## 最佳实践

1. **先查询接口模型列表**，了解连接器支持的模型
2. **使用 `--filter-model-type` 过滤**，获取特定类型的模型
3. **动态模型视图包含实际字段定义**，适合用于字段映射
4. **模型元数据是只读的**，不可通过此接口修改
