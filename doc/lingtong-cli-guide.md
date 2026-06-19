# lingtong-cli 应用管理命令使用指南

## 概述

`lingtong-cli app` 命令集用于管理绫通平台上的集成应用配置。支持应用列表查询、详情查询、导出、导入、验证、脚手架生成、差异比较以及场景管理。

所有命令通过 `lingtong-cli app` 子命令访问，采用 JSON 格式作为应用配置的交换格式。

## 命令总览

| 命令 | 用途 |
|------|------|
| `app list` | 列出平台应用，支持分页和名称/ID/租户过滤 |
| `app get` | 查询单个应用详情 |
| `app export` | 从平台导出应用配置为 JSON 文件 |
| `app validate` | 验证导出的 JSON 文件结构是否正确 |
| `app import` | 将 JSON 配置导入到绫通平台 |
| `app scaffold` | 生成最小化的应用配置模板 |
| `app diff` | 比较两个应用配置文件的差异 |
| `app scene list` | 列出应用中定义的所有场景 |
| `app scene add` | 向应用添加新场景 |
| `app scene remove` | 从应用移除场景 |

## 应用配置结构

导出的应用 JSON 文件包含以下顶层字段：

```json
{
  "appName": "应用名称",
  "appConnectors": ["连接器A", "连接器B"],
  "basicDatas": [{}],
  "scenes": [{}],
  "workflows": [{}],
  "workflowConnectors": [{}]
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `appName` | string | 是 | 应用名称，不能为空 |
| `appConnectors` | array | 是 | 应用使用的连接器名称列表 |
| `basicDatas` | array | 是 | 基础资料配置 |
| `scenes` | array | 是 | 集成场景列表 |
| `workflows` | array | 否 | 工作流配置 |
| `workflowConnectors` | array | 否 | 工作流连接器配置 |

## 快速开始：创建新的集成应用

### 方式一：使用模板（推荐）

对于快麦 ERP 到金蝶云星空的集成场景，使用内置模板：

```bash
lingtong-cli app scaffold --name "my-integration" --template kuaimai-kingdee --output app.json
```

模板会自动生成：
- 5 个基础资料（快麦店铺、快麦仓库、金蝶物料、金蝶仓位、店铺映射表）
- 7 个预配置场景（商品同步、销售订单同步、销退入库单同步、库存同步等）
- 连接器配置（kmerp 和 kingDeeCloudStar）

### 方式二：手动指定连接器

```bash
lingtong-cli app scaffold --name "my-app" --source "connector-a" --target "connector-b" --output app.json
```

生成一个空的应用骨架，包含基础结构但场景和基础资料为空。

### 导入到平台

```bash
# 先验证文件
lingtong-cli app validate --file app.json

# 预览导入内容
lingtong-cli app import --file app.json --dry-run

# 执行导入
lingtong-cli app import --file app.json
```

### 从平台导出已有应用

```bash
lingtong-cli app export --app-id 123 --output backup.json
```

在线平台命令（`app list/get/export/import`）需要先配置 Host：

```bash
lingtong-cli config init --host https://your-lingtong-host.com
```

未配置时会返回：`no host configured. Run lingtong-cli config init --host <url> first`。`app import --dry-run` 只读取本地文件，不要求 Host。

## 命令参考

### app list

列出绫通平台应用，不修改平台状态。

```bash
lingtong-cli app list [--page-num <n>] [--page-size <n>] [--name <name>] [--app-id <id>] [--tenant-id <id>]
```

**参数：**

| 参数 | 必填 | 说明 |
|------|------|------|
| `--page-num` | 否 | 页码，默认 1 |
| `--page-size` | 否 | 每页数量，默认 40 |
| `--name` | 否 | 应用名称过滤 |
| `--app-id` | 否 | 应用 ID 过滤 |
| `--tenant-id` | 否 | 租户 ID 过滤 |

**示例：**

```bash
lingtong-cli app list --page-num 1 --page-size 40
lingtong-cli app list --name "Sales Sync" --format json
```

---

### app get

查询单个绫通平台应用详情，不修改平台状态。

```bash
lingtong-cli app get --application-id <id>
# 或
lingtong-cli app get --id <id>
```

**参数：**

| 参数 | 必填 | 说明 |
|------|------|------|
| `--application-id` | 条件 | 应用 ID；与 `--id` 二选一 |
| `--id` | 条件 | `--application-id` 的别名 |

**示例：**

```bash
lingtong-cli app get --application-id 123
lingtong-cli app get --id 123 --format pretty
```

---

### app export

从绫通平台导出应用配置为 JSON 文件。

```bash
lingtong-cli app export --app-id <app-id> --output <file> [--dry-run]
```

**参数：**

| 参数 | 必填 | 说明 |
|------|------|------|
| `--app-id` | 是 | 应用 ID |
| `--output` | 是 | 输出文件路径 |
| `--dry-run` | 否 | 仅打印将执行的操作，不实际导出 |

**示例：**

```bash
# 导出应用到文件
lingtong-cli app export --app-id 123 --output app.json

# 预览导出操作
lingtong-cli app export --app-id 123 --output app.json --dry-run
```

**工作原理：**
1. 调用 `GET /application/export?sign=<sign>&appId=<app-id>` API
2. 验证返回的 JSON 格式
3. 格式化后写入指定文件

---

### app validate

验证应用导出 JSON 文件的结构和内容。

```bash
lingtong-cli app validate --file <file>
```

**参数：**

| 参数 | 必填 | 说明 |
|------|------|------|
| `--file` | 是 | 要验证的 JSON 文件路径 |

**验证规则：**

| 规则 | 说明 |
|------|------|
| 应用名称 | 必须非空，且不含危险特殊字符 |
| 连接器列表 | 必须非空数组 |
| 基础资料 | 必须存在（允许空数组） |
| 场景列表 | 必须非空数组 |
| 场景名称 | 每个场景必须有名称 |
| 连接器引用 | 场景引用的连接器必须在 appConnectors 中声明 |
| 循环依赖 | 检测场景间的循环连接器依赖 |
| 字段映射 | 检查 sourceField 和 targetField 是否为空 |
| 重复场景名 | 检测同名的重复场景 |
| 特殊字符 | 名称中不允许引号、斜杠、控制字符 |
| 未知顶层键 | 只允许已知的 6 个顶层字段 |
| 空值检查 | 必填字段不能为 null |
| 类型检查 | appConnectors 必须是字符串数组 |
| 大小限制 | JSON 文件不能超过 10MB |

**示例：**

```bash
lingtong-cli app validate --file app.json
# 输出: Validation passed: my-app
```

**失败示例：**

```bash
lingtong-cli app validate --file invalid.json
# 输出:
# validation failed:
# appName: appName is required and must be a non-empty string
# scenes[0].connectorSource.name: connector 'unknown-conn' not declared in appConnectors (preflight check)
```

---

### app import

将应用配置从 JSON 文件导入绫通平台。

```bash
lingtong-cli app import --file <file> [--dry-run] [--force]
```

**参数：**

| 参数 | 必填 | 说明 |
|------|------|------|
| `--file` | 是 | 要导入的 JSON 文件路径 |
| `--dry-run` | 否 | 预览导入内容，不实际执行 |
| `--force` | 否 | 跳过预飞行验证检查 |

**示例：**

```bash
# 预览导入
lingtong-cli app import --file app.json --dry-run
# 输出:
# [dry-run] Would import application: my-app
# [dry-run] Connectors: 2
# [dry-run] Scenes: 3
# [dry-run] Workflows: 1
# [dry-run] Would POST: /application/import

# 执行导入（含预飞行验证）
lingtong-cli app import --file app.json

# 强制导入（跳过验证）
lingtong-cli app import --file app.json --force
```

**导入流程：**
1. 读取并解析 JSON 文件
2. 执行预飞行验证（除非使用 `--force`）
3. 调用 `POST /application/import` API
4. 返回创建的场景和表格列表

---

### app scaffold

生成最小化的应用配置模板。

```bash
# 标准模式
lingtong-cli app scaffold --name <app> --source <conn> --target <conn> [--output <file>] [--dry-run]

# 模板模式
lingtong-cli app scaffold --name <app> --template <template> [--output <file>] [--dry-run]
```

**参数：**

| 参数 | 必填 | 说明 |
|------|------|------|
| `--name` | 是 | 应用名称 |
| `--source` | 条件 | 源连接器名称（标准模式必填） |
| `--target` | 条件 | 目标连接器名称（标准模式必填） |
| `--template` | 否 | 使用预配置模板（覆盖 source/target） |
| `--output` | 条件 | 输出文件路径（不使用 --dry-run 时必填） |
| `--dry-run` | 否 | 输出 JSON 到终端而非文件 |

**可用模板：**

| 模板名 | 连接器 | 场景数 | 基础资料数 |
|--------|--------|--------|-----------|
| `kuaimai-kingdee` | kmerp, kingDeeCloudStar | 7 | 5 |

**kuaimai-kingdee 模板详情：**

基础资料：
- 快麦店铺（连接器: kmerp，自动同步: 1h 间隔）
- 快麦仓库（连接器: kmerp，自动同步: 1h 间隔）
- 金蝶物料（连接器: kingDeeCloudStar，自动同步）
- 金蝶仓位（连接器: kingDeeCloudStar，自动同步）
- 店铺映射表（手动管理）

场景：

| 场景名 | 方向 | 触发方式 | 描述 |
|--------|------|----------|------|
| 商品同步 | 金蝶 . 快麦 | 定时 | 定时查询金蝶商品列表同步到快麦 |
| 销售订单同步 | 快麦 . 金蝶 | 定时 | 快麦货位进出记录转金蝶销售出库单 |
| 销退入库单同步 | 快麦 . 金蝶 | 实时 | 快麦销退上架转金蝶销退入库单 |
| 库存同步 | 金蝶 . 快麦 | 定时 | 定时查询金蝶库存修改快麦实际库存 |
| 质量管理发货通知单 | 金蝶 . 快麦 | 定时 | 抓取金蝶已审核发货通知单转快麦手工单 |
| 线下销售出库单下推 | 金蝶 . 金蝶 | 手动 | 金蝶发货通知单下推销售出库单 |
| 线下销售出库单操作 | 快麦 . 金蝶 | 手动 | 下推返回数据替换货位映射后保存审核 |

**示例：**

```bash
# 使用模板生成
lingtong-cli app scaffold --name "km-kd-integration" --template kuaimai-kingdee --output app.json

# 预览模板内容
lingtong-cli app scaffold --name "my-app" --template kuaimai-kingdee --dry-run

# 标准模式生成空骨架
lingtong-cli app scaffold --name "my-app" --source "salesforce" --target "sap" --output app.json
```

---

### app diff

比较两个应用配置文件的差异。

```bash
lingtong-cli app diff --file-a <file1> --file-b <file2>
```

**参数：**

| 参数 | 必填 | 说明 |
|------|------|------|
| `--file-a` | 是 | 第一个 JSON 文件（基准） |
| `--file-b` | 是 | 第二个 JSON 文件（对比） |

**比较内容：**
- 新增/删除/修改的场景（按名称比较）
- 新增/删除的连接器
- 新增/删除的基础资料（按数量比较）

**输出格式：**

```
Comparing: app-v1.json -> app-v2.json
--------------------------------------------------
Added scenes (2):
  + 库存同步
  + 商品同步

Removed scenes (1):
  - 旧场景

Added connectors (1):
  + kingDeeCloudStar

Removed connectors (1):
  - old-connector
```

**示例：**

```bash
# 比较两个版本
lingtong-cli app diff --file-a app-v1.json --file-b app-v2.json

# 无差异时
lingtong-cli app diff --file-a a.json --file-b b.json
# 输出: No differences found.
```

---

### app scene list

列出应用中定义的所有场景。

```bash
lingtong-cli app scene list --file <file> [--format json|table|pretty]
```

**参数：**

| 参数 | 必填 | 说明 |
|------|------|------|
| `--file` | 是 | 应用 JSON 文件路径 |
| `--format` | 否 | 输出格式，默认 json |

**场景结构：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | string | 场景名称 |
| `description` | string | 场景描述 |
| `source` | string | 源连接器 |
| `target` | string | 目标连接器 |
| `trigger` | string | 触发方式（scheduled/realtime/manual） |

**示例：**

```bash
# JSON 格式（默认）
lingtong-cli app scene list --file app.json

# 表格格式
lingtong-cli app scene list --file app.json --format table
```

---

### app scene add

向应用添加新场景。

```bash
lingtong-cli app scene add --file <file> --name <name> --source <conn> --target <conn> [--output <file>]
```

**参数：**

| 参数 | 必填 | 说明 |
|------|------|------|
| `--file` | 是 | 应用 JSON 文件路径 |
| `--name` | 是 | 场景名称 |
| `--source` | 是 | 源连接器名称 |
| `--target` | 是 | 目标连接器名称 |
| `--output` | 否 | 输出文件路径，默认覆盖原文件 |

**示例：**

```bash
# 添加场景并覆盖原文件
lingtong-cli app scene add --file app.json --name "Order Sync" --source kmerp --target kingDeeCloudStar

# 添加场景并输出到新文件
lingtong-cli app scene add --file app.json --name "Order Sync" --source kmerp --target kingDeeCloudStar --output updated.json
```

---

### app scene remove

从应用移除场景。

```bash
lingtong-cli app scene remove --file <file> --name <name> [--output <file>]
```

**参数：**

| 参数 | 必填 | 说明 |
|------|------|------|
| `--file` | 是 | 应用 JSON 文件路径 |
| `--name` | 是 | 要移除的场景名称（不区分大小写） |
| `--output` | 否 | 输出文件路径，默认覆盖原文件 |

**示例：**

```bash
# 移除场景
lingtong-cli app scene remove --file app.json --name "Order Sync"

# 移除并输出到新文件
lingtong-cli app scene remove --file app.json --name "Order Sync" --output updated.json
```

## 典型工作流

### 场景一：从零创建集成应用

```bash
# 1. 使用模板生成应用骨架
lingtong-cli app scaffold --name "erp-integration" --template kuaimai-kingdee --output app.json

# 2. 验证生成的配置
lingtong-cli app validate --file app.json

# 3. 预览导入内容
lingtong-cli app import --file app.json --dry-run

# 4. 执行导入
lingtong-cli app import --file app.json
```

### 场景二：备份和恢复应用

```bash
# 1. 导出应用配置作为备份
lingtong-cli app export --app-id 123 --output backup-20260418.json

# 2. 验证备份文件
lingtong-cli app validate --file backup-20260418.json

# 3. 需要时恢复
lingtong-cli app import --file backup-20260418.json --force
```

### 场景三：版本对比和审计

```bash
# 1. 导出当前版本
lingtong-cli app export --app-id 123 --output current.json

# 2. 与历史版本对比
lingtong-cli app diff --file-a backup-20260401.json --file-b current.json

# 3. 列出所有场景确认
lingtong-cli app scene list --file current.json --format table
```

### 场景四：增量添加场景

```bash
# 1. 导出应用
lingtong-cli app export --app-id 123 --output app.json

# 2. 添加新场景
lingtong-cli app scene add --file app.json --name "New Sync" --source src --target tgt

# 3. 验证修改后的文件
lingtong-cli app validate --file app.json

# 4. 导入更新
lingtong-cli app import --file app.json
```

## 故障排查

### 常见问题

| 错误 | 原因 | 解决方法 |
|------|------|----------|
| `--app-id is required` | 导出时未提供应用 ID | 添加 `--app-id` 参数 |
| `--output is required` | 未指定输出文件且未使用 `--dry-run` | 添加 `--output` 或 `--dry-run` |
| `invalid JSON` | 文件格式不正确 | 使用 JSON 校验工具检查文件 |
| `validation failed` | 配置结构不符合要求 | 运行 `app validate` 查看详细错误 |
| `scene not found` | 场景名称不匹配 | 使用 `scene list` 确认场景名称 |
| `no host configured` | 尚未配置平台 Host | 运行 `lingtong-cli config init --host https://your-lingtong-host.com` |
| `circular connector dependency` | 场景间形成循环依赖 | 检查场景的 source/target 是否形成环路 |
| `duplicate scene name` | 存在同名场景 | 修改场景名称使其唯一 |
| `payload size exceeds maximum` | JSON 文件超过 10MB | 拆分应用或减少配置数据 |

### 验证失败排查步骤

1. 运行 `lingtong-cli app validate --file <file>` 获取详细错误列表
2. 检查 `appName` 是否非空且不含特殊字符
3. 确认 `appConnectors` 是非空数组
4. 检查每个场景的 `name` 字段
5. 确保场景引用的连接器在 `appConnectors` 中声明
6. 检查是否存在循环依赖（如 A . B . C . A）
7. 确认场景名称唯一
8. 检查 `fieldMappings` 中的字段映射是否完整

### dry-run 调试技巧

所有修改类命令都支持 `--dry-run` 参数。在不确定命令会做什么时，先加 `--dry-run` 预览：

```bash
# 预览导出操作
lingtong-cli app export --app-id 123 --output app.json --dry-run

# 预览导入内容
lingtong-cli app import --file app.json --dry-run

# 预览模板输出
lingtong-cli app scaffold --name "app" --template kuaimai-kingdee --dry-run
```

## 输出格式

`app` 命令支持 `--format` 参数控制输出格式：

| 格式 | 说明 | 适用场景 |
|------|------|----------|
| `json` | 紧凑 JSON 输出 | 脚本处理和管道操作 |
| `pretty` | 格式化的 JSON 输出 | 人工阅读 |
| `table` | 表格格式输出 | 终端查看 |

```bash
lingtong-cli app scene list --file app.json --format table
```
