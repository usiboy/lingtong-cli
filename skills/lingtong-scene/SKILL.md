---
name: lingtong-scene
version: 1.0.0
description: "绫通场景管理：创建、查询、列出集成场景。当用户需要管理集成场景、创建新的数据同步流程、查询场景详情时触发。关键词：scene、场景、scene list、scene create、integration。"
---

# lingtong-scene 技能

## 概述

本技能指导你如何通过 `lingtong-cli` 管理绫通平台的集成场景。

## 核心命令

### 列出场景

```bash
lingtong-cli scene list [--page <n>] [--page-size <n>]
```

**示例**:
```bash
# 列出所有场景
lingtong-cli scene list

# 分页查询
lingtong-cli scene list --page 1 --page-size 50
```

### 创建场景

```bash
lingtong-cli scene create --name <name> [--description <desc>]
```

**示例**:
```bash
# 创建订单同步场景
lingtong-cli scene create --name "订单同步" --description "从金蝶云同步订单到自有ERP"
```

### 查询场景详情

```bash
lingtong-cli scene info --scene-id <id>
```

**示例**:
```bash
lingtong-cli scene info --scene-id 123
```

## 快捷命令

```bash
# 快捷列出场景
lingtong-cli +scene-list
```

## 数据模型

场景返回的典型数据结构：

```json
{
  "success": true,
  "data": {
    "id": 123,
    "name": "订单同步",
    "description": "从金蝶云同步订单到自有ERP",
    "status": "active",
    "createdAt": "2024-01-01T00:00:00Z",
    "workflows": [...]
  }
}
```

## 最佳实践

1. **先列出场景**，了解现有场景结构
2. **创建场景时使用描述性名称**，便于后续管理
3. **场景是工作流的容器**，一个场景可包含多个工作流
