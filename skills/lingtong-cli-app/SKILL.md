---
name: lingtong-cli-app
version: 1.0.0
description: 绫通 CLI 应用管理技能。用于导出、导入、验证、生成和比较集成应用配置。当需要管理 lingtong 应用生命周期、创建新集成应用、导出应用配置、验证应用结构或比较应用差异时触发。关键词：app、应用、export、import、validate、scaffold、diff、scene。
metadata:
  requires:
    bins:
      - lingtong-cli
---

# 绫通 CLI 应用管理技能

管理绫通集成应用的完整生命周期，包括导出、导入、验证、脚手架生成和场景管理。

## 何时使用此技能

当以下情况出现时使用此技能：

- 需要从绫通平台导出应用配置到 JSON 文件
- 需要将应用配置导入到绫通平台
- 需要验证应用导出文件的结构和引用完整性
- 需要快速生成新的集成应用模板
- 需要比较两个应用配置文件的差异
- 需要管理应用中的场景（scene）列表
- 需要使用预置模板创建快麦-金蝶集成应用

## 前置要求

- `lingtong-cli` 已安装并可用
- 已通过 `lingtong-cli config init --host https://your-lingtong-host.com` 配置平台 Host
- 已通过 `lingtong-cli auth login` 完成认证，或将令牌放入 `LINGTONG_API_TOKEN` 后使用 `lingtong-cli auth login --from-env`

## 应用数据结构

应用导出文件使用统一的 JSON 结构：

```json
{
  "appName": "应用名称",
  "appConnectors": ["连接器A", "连接器B"],
  "basicDatas": [
    {
      "name": "基础资料名称",
      "dataSource": "connector|manual",
      "connector": "连接器名称",
      "autoSync": true,
      "syncInterval": "1h"
    }
  ],
  "scenes": [
    {
      "name": "场景名称",
      "description": "场景描述",
      "source": "源连接器",
      "target": "目标连接器",
      "trigger": "scheduled|realtime|manual"
    }
  ],
  "workflows": [],
  "workflowConnectors": []
}
```

## 命令参考

### lingtong-cli app list

列出平台应用，不修改平台状态。

```bash
lingtong-cli app list [--page-num <n>] [--page-size <n>] [--name <name>] [--app-id <id>] [--tenant-id <id>]
```

| 参数 | 必填 | 说明 |
|------|------|------|
| `--page-num` | 否 | 页码，默认 1 |
| `--page-size` | 否 | 每页数量，默认 40 |
| `--name` | 否 | 应用名称过滤 |
| `--app-id` | 否 | 应用 ID 过滤 |
| `--tenant-id` | 否 | 租户 ID 过滤 |

示例：

```bash
lingtong-cli app list --page-num 1 --page-size 40
lingtong-cli app list --name "Sales Sync" --format json
```

### lingtong-cli app get

查询单个平台应用详情，不修改平台状态。

```bash
lingtong-cli app get --application-id <id>
lingtong-cli app get --id <id>
```

| 参数 | 必填 | 说明 |
|------|------|------|
| `--application-id` | 条件 | 应用 ID；与 `--id` 二选一 |
| `--id` | 条件 | `--application-id` 的别名 |

示例：

```bash
lingtong-cli app get --application-id 123
lingtong-cli app get --id 123 --format pretty
```

### lingtong-cli app export

导出应用配置到 JSON 文件。

```bash
lingtong-cli app export --app-id <app-id> --output <file> [--dry-run]
```

| 参数 | 必填 | 说明 |
|------|------|------|
| `--app-id` | 是 | 应用 ID |
| `--output` | 是 | 输出文件路径 |
| `--dry-run` | 否 | 预览操作，不执行 |

示例：

```bash
# 导出应用到文件
lingtong-cli app export --app-id 123 --output app.json

# 预览导出操作
lingtong-cli app export --app-id 123 --output app.json --dry-run
```

### lingtong-cli app validate

验证应用导出文件的结构和引用完整性。

```bash
lingtong-cli app validate --file <file>
```

| 参数 | 必填 | 说明 |
|------|------|------|
| `--file` | 是 | 输入 JSON 文件路径 |

示例：

```bash
lingtong-cli app validate --file app.json
```

### lingtong-cli app import

将应用配置导入到绫通平台。

```bash
lingtong-cli app import --file <file> [--dry-run] [--force]
```

| 参数 | 必填 | 说明 |
|------|------|------|
| `--file` | 是 | 输入 JSON 文件路径 |
| `--dry-run` | 否 | 预览导入内容，不执行 |
| `--force` | 否 | 跳过预检查验证 |

示例：

```bash
# 导入应用
lingtong-cli app import --file app.json

# 预览导入
lingtong-cli app import --file app.json --dry-run

# 强制导入（跳过验证）
lingtong-cli app import --file app.json --force
```

### lingtong-cli app scaffold

生成最小应用配置模板。

```bash
lingtong-cli app scaffold --name <app> --source <conn> --target <conn> [--template kuaimai-kingdee] [--output <file>] [--dry-run]
```

| 参数 | 必填 | 说明 |
|------|------|------|
| `--name` | 是 | 应用名称 |
| `--source` | 模板模式否 | 源连接器名称 |
| `--target` | 模板模式否 | 目标连接器名称 |
| `--template` | 否 | 预置模板：kuaimai-kingdee |
| `--output` | 模板模式否 | 输出文件路径 |
| `--dry-run` | 否 | 输出 JSON 到控制台 |

示例：

```bash
# 标准模式：生成空白模板
lingtong-cli app scaffold --name "my-app" --source "connector-a" --target "connector-b" --output app.json

# 预览生成内容
lingtong-cli app scaffold --name "my-app" --source "connector-a" --target "connector-b" --dry-run

# 使用预置模板
lingtong-cli app scaffold --name "kuaimai-kingdee" --template kuaimai-kingdee --output app.json
```

### lingtong-cli app diff

比较两个应用配置文件的差异。

```bash
lingtong-cli app diff --file-a <file1> --file-b <file2>
```

| 参数 | 必填 | 说明 |
|------|------|------|
| `--file-a` | 是 | 第一个应用 JSON 文件 |
| `--file-b` | 是 | 第二个应用 JSON 文件 |

示例：

```bash
lingtong-cli app diff --file-a app-v1.json --file-b app-v2.json
```

输出格式：

```
Comparing: app-v1.json -> app-v2.json
--------------------------------------------------
Added scenes (2):
  + 库存同步
  + 商品同步

Removed connectors (1):
  - old-connector

Added connectors (1):
  + new-connector
```

### lingtong-cli app scene list

列出应用中的所有场景。

```bash
lingtong-cli app scene list --file <file> [--format json|table|pretty]
```

| 参数 | 必填 | 说明 |
|------|------|------|
| `--file` | 是 | 应用 JSON 文件路径 |
| `--format` | 否 | 输出格式，默认 json |

示例：

```bash
lingtong-cli app scene list --file app.json
lingtong-cli app scene list --file app.json --format table
```

### lingtong-cli app scene add

向应用添加新场景。

```bash
lingtong-cli app scene add --file <file> --name <name> --source <conn> --target <conn> [--output <file>]
```

| 参数 | 必填 | 说明 |
|------|------|------|
| `--file` | 是 | 应用 JSON 文件路径 |
| `--name` | 是 | 场景名称 |
| `--source` | 是 | 源连接器名称 |
| `--target` | 是 | 目标连接器名称 |
| `--output` | 否 | 输出文件路径，默认覆盖输入文件 |

示例：

```bash
# 添加场景并覆盖原文件
lingtong-cli app scene add --file app.json --name "Order Sync" --source kmerp --target kingDeeCloudStar

# 添加场景并输出到新文件
lingtong-cli app scene add --file app.json --name "Order Sync" --source kmerp --target kingDeeCloudStar --output updated.json
```

### lingtong-cli app scene remove

从应用移除场景。

```bash
lingtong-cli app scene remove --file <file> --name <name> [--output <file>]
```

| 参数 | 必填 | 说明 |
|------|------|------|
| `--file` | 是 | 应用 JSON 文件路径 |
| `--name` | 是 | 要移除的场景名称（不区分大小写） |
| `--output` | 否 | 输出文件路径，默认覆盖输入文件 |

示例：

```bash
lingtong-cli app scene remove --file app.json --name "Order Sync"
lingtong-cli app scene remove --file app.json --name "Order Sync" --output updated.json
```

## 工作流：创建新集成应用

从零开始创建并导入一个完整的集成应用：

```bash
# 1. 使用模板生成应用配置
lingtong-cli app scaffold --name "my-integration" --source "connector-a" --target "connector-b" --output app.json

# 2. 验证生成的配置
lingtong-cli app validate --file app.json

# 3. 添加具体场景
lingtong-cli app scene add --file app.json --name "Data Sync" --source "connector-a" --target "connector-b"

# 4. 再次验证
lingtong-cli app validate --file app.json

# 5. 预览导入内容
lingtong-cli app import --file app.json --dry-run

# 6. 执行导入
lingtong-cli app import --file app.json
```

## 工作流：导出并修改现有应用

导出线上应用配置，本地修改后重新导入：

```bash
# 1. 从平台导出应用
lingtong-cli app export --app-id 123 --output app-backup.json

# 2. 备份原始文件
cp app-backup.json app-backup.json.bak

# 3. 查看当前场景列表
lingtong-cli app scene list --file app-backup.json --format table

# 4. 添加新场景
lingtong-cli app scene add --file app-backup.json --name "New Sync" --source "conn-a" --target "conn-b"

# 5. 验证修改后的配置
lingtong-cli app validate --file app-backup.json

# 6. 比较修改前后的差异
lingtong-cli app diff --file-a app-backup.json.bak --file-b app-backup.json

# 7. 导入更新后的配置
lingtong-cli app import --file app-backup.json
```

## 工作流：比较两个应用

比较两个不同版本或环境的应用配置：

```bash
# 比较两个版本
lingtong-cli app diff --file-a app-v1.json --file-b app-v2.json

# 比较开发和生产环境
lingtong-cli app diff --file-a app-dev.json --file-b app-prod.json
```

差异报告会显示：

- 新增的场景（Added scenes）
- 移除的场景（Removed scenes）
- 修改的场景（Modified scenes）
- 新增的连接器（Added connectors）
- 移除的连接器（Removed connectors）
- 新增的基础资料（Added basicDatas）
- 移除的基础资料（Removed basicDatas）

## 工作流：管理场景

对应用中的场景进行增删查操作：

```bash
# 查看所有场景
lingtong-cli app scene list --file app.json --format table

# 添加场景
lingtong-cli app scene add --file app.json --name "Order Sync" --source kmerp --target kingDeeCloudStar

# 移除场景
lingtong-cli app scene remove --file app.json --name "Order Sync"

# 验证场景变更
lingtong-cli app validate --file app.json
```

## 模板：快麦-金蝶集成

`kuaimai-kingdee` 模板预置了快麦 ERP 与金蝶云星辰的完整集成方案。

### 基础资料

模板包含 5 个预配置的基础资料：

| 名称 | 数据源 | 连接器 | 自动同步 | 同步间隔 |
|------|--------|--------|----------|----------|
| 快麦店铺 | connector | kmerp | 是 | 1h |
| 快麦仓库 | connector | kmerp | 是 | 1h |
| 金蝶物料 | connector | kingDeeCloudStar | 是 | - |
| 金蝶仓位 | connector | kingDeeCloudStar | 是 | - |
| 店铺映射表 | manual | - | 否 | - |

### 场景

模板包含 7 个预配置场景：

| 场景名称 | 源 | 目标 | 触发方式 | 描述 |
|----------|-----|------|----------|------|
| 商品同步 | kingDeeCloudStar | kmerp | scheduled | 定时查询金蝶商品列表同步到快麦 |
| 销售订单同步 | kmerp | kingDeeCloudStar | scheduled | 快麦货位进出记录转金蝶销售出库单 |
| 销退入库单同步 | kmerp | kingDeeCloudStar | realtime | 快麦销退上架转金蝶销退入库单 |
| 库存同步 | kingDeeCloudStar | kmerp | scheduled | 定时查询金蝶库存修改快麦实际库存 |
| 质量管理发货通知单 | kingDeeCloudStar | kmerp | scheduled | 定时抓取金蝶已审核发货通知单转快麦系统手工单 |
| 线下销售出库单下推 | kingDeeCloudStar | kingDeeCloudStar | manual | 金蝶发货通知单下推销售出库单 |
| 线下销售出库单操作 | kmerp | kingDeeCloudStar | manual | 下推返回数据替换货位映射后保存审核 |

### 使用模板

```bash
lingtong-cli app scaffold --name "my-kuaimai-kingdee" --template kuaimai-kingdee --output app.json
```

## 验证规则

`validate` 命令检查以下内容：

| 检查项 | 规则 | 错误信息 |
|--------|------|----------|
| appName | 必须为非空字符串 | appName is required and must be a non-empty string |
| appConnectors | 必须为非空数组 | appConnectors is required and must be a non-empty array |
| basicDatas | 必须存在 | basicDatas is required |
| scenes | 必须为非空数组 | scenes is required and must be a non-empty array |
| scene.name | 每个场景必须有名称 | scene name is required |
| scene.connectorSource | 源连接器必须在 appConnectors 中 | connector 'X' not declared in appConnectors (preflight check) |
| scene.connectorTarget | 目标连接器必须在 appConnectors 中 | connector 'X' not declared in appConnectors (preflight check) |

## 故障排除

### 导出失败：应用不存在

```
failed to export application: application not found
```

确保应用名称正确。使用 `lingtong-cli app list` 查看可用应用列表。

### 验证失败：缺少必填字段

```
validation failed:
appName: appName is required and must be a non-empty string
appConnectors: appConnectors is required and must be a non-empty array
```

使用 `scaffold` 命令生成有效的模板文件，确保所有必填字段都存在。

### 导入失败：验证错误

```
pre-flight validation failed:
scenes[0].connectorSource.name: connector 'unknown' not declared in appConnectors (preflight check)
```

场景引用的连接器不在应用的连接器列表中。确保 `appConnectors` 包含所有场景使用的连接器。

### 场景未找到

```
scene "Order Sync" not found in application
```

`app scene remove` 按名称大小写不敏感匹配；如仍找不到，请使用 `app scene list --file <file>` 确认场景名称是否存在。

### 认证失败

```
failed to export application: unauthorized
```

运行 `lingtong-cli auth login` 重新认证，或检查 Keychain 中保存的 Token。非交互登录可先设置 `LINGTONG_API_TOKEN`，再运行 `lingtong-cli auth login --from-env`。

### 未配置 Host

```
no host configured. Run `lingtong-cli config init --host <url>` first
```

运行 `lingtong-cli config init --host https://your-lingtong-host.com` 后重试。此错误会在 `app list/get/export/import` 等需要访问平台的命令中出现。

### 文件读写失败

```
failed to read file app.json: no such file or directory
```

检查文件路径是否正确，使用绝对路径或确认当前工作目录。

## 相关技能

- 加载 `lingtong-api` 技能了解底层 API 调用
- 加载 `lingtong-factory` 技能了解连接器配置
- 加载 `lingtong-workflow` 技能了解工作流编排
- 加载 `lingtong-table-core` 技能了解表格数据管理
