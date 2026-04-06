---
name: lingtong-table
version: 1.0.0
description: "绫通表格管理：表格 CRUD、数据查询、记录创建。当用户需要操作绫通表格、查询数据、创建记录时触发。关键词：table、表格、table data、query、create record。"
---

# lingtong-table 技能

## 概述

本技能指导你如何通过 `lingtong-cli` 管理绫通平台的表格数据。

## 核心命令

### 列出表格

```bash
lingtong-cli table list
```

### 查询表格数据

```bash
lingtong-cli table data query --table-id <id> [--filter <json>] [--sort <json>]
```

**示例**:
```bash
# 查询所有数据
lingtong-cli table data query --table-id 123

# 带过滤条件
lingtong-cli table data query --table-id 123 --filter '{"status":"active"}'

# 带排序
lingtong-cli table data query --table-id 123 --sort '{"createdAt":"desc"}'
```

### 创建记录

```bash
lingtong-cli table data create --table-id <id> --data <json>
```

**示例**:
```bash
lingtong-cli table data create --table-id 123 --data '{"name":"测试","value":100,"status":"active"}'
```

## 数据模型

表格数据查询返回的典型数据结构：

```json
{
  "success": true,
  "data": {
    "total": 100,
    "records": [
      {
        "id": 1,
        "fields": {
          "name": "测试",
          "value": 100,
          "status": "active"
        },
        "createdAt": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

## 最佳实践

1. **使用 `--filter` 精确查询**，减少数据传输量
2. **批量创建记录时**，可考虑多次调用或使用 API 批量接口
3. **表格数据变更是实时的**，创建/更新后立即生效
