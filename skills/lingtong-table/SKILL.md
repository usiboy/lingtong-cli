---
name: lingtong-table
version: 2.9.0
description: "绫通表格管理：表格 CRUD、数据查询、记录创建、批量更新、删除、Schema 管理。当用户需要操作绫通表格、查询数据、创建记录、批量更新、删除记录、理解表格结构时触发。关键词：table、表格、table data、query、create record、batch-update、delete、schema。"
---

# lingtong-table 技能

## 概述

本技能指导你如何通过 `lingtong-cli` 管理绫通平台的表格数据。

## 前置条件

执行表格命令前，需要先配置平台 Host：
```bash
lingtong-cli config init --host https://your-lingtong-host.com
```

## 核心命令

### 列出表格

```bash
lingtong-cli table list [--app-id <id>]
```

**示例**:
```bash
# 列出所有表格
lingtong-cli table list

# 按应用筛选
lingtong-cli table list --app-id 165
```

### 创建表格

创建表格分为两步：先创建表格基本信息，再创建 Schema（可选）。

```bash
lingtong-cli table create --app-id <id> --name <name> [--source <n>] [--type <n>] [--open-high-mode <n>] [--open-connector <n>] [--columns-schema <json>]
```

**参数说明**:
- `--app-id`: 应用 ID（必填）
- `--name`: 表格名称（必填）
- `--source`: 来源，1-手工维护，2-连接器（默认 1）
- `--type`: 类型，1-常规资料，2-结构映射（默认 1）
- `--open-high-mode`: 高性能模式，0-关闭，1-开启（默认 0）
- `--open-connector`: 连接器同步，0-关闭，1-开启（默认 0）
- `--columns-schema`: Schema 字段定义 JSON 数组（可选）

**示例**:
```bash
# 创建手工表格
lingtong-cli table create --app-id 165 --name "客户资料" --source 1 --type 1

# 创建高性能模式表格
lingtong-cli table create --app-id 165 --name "订单表" --source 1 --type 1 --open-high-mode 1

# 创建带 Schema 的表格
lingtong-cli table create --app-id 165 --name "产品表" --source 1 --type 1 --columns-schema '[{"title":"名称","key":"name","type":"text"},{"title":"价格","key":"price","type":"number"},{"title":"日期","key":"date","type":"date","dateFormat":"YYYY-MM-DD"}]'
```

**响应结构**:
```json
{
  "success": true,
  "result": {
    "id": 1630,
    "name": "产品表",
    "appId": 165,
    "source": 1,
    "type": 1,
    "openHighMode": 0
  },
  "schemaResult": {
    "id": 2545,
    "basicDataId": 1630,
    "version": 1,
    "columnsSchema": [...]
  }
}
```

### 查询表格数据

```bash
lingtong-cli table data query --basic-data-id <id> \
  [--text <string>] [--filter <json>] [--view-id <id>] \
  [--ids <csv>] \
  [--start-update-time <ms>] [--end-update-time <ms>] \
  [--primary-key <json-array>] \
  [--page <n>] [--page-size <n>] \
  [--order-by <field>] [--order-asc]
```

**参数说明**：
| 参数 | 说明 |
|------|------|
| `--basic-data-id` | 表格 ID（必填） |
| `--text` | 全文搜索（MySQL 模式：LIKE 模糊匹配） |
| `--filter` | 字段级精确匹配（JSON：`{"fieldKey":"value"}`） |
| `--view-id` | 视图 ID（加载已保存的过滤/排序配置） |
| `--ids` | 按记录 ID 列表筛选，逗号分隔（如 `1001,1002`） |
| `--start-update-time` | 更新时间起始（毫秒时间戳） |
| `--end-update-time` | 更新时间截止（毫秒时间戳） |
| `--primary-key` | 按主键值筛选（JSON 数组） |
| `--page` | 页码（默认 1） |
| `--page-size` | 每页条数（默认 20） |
| `--order-by` | 排序字段（如 `created`、`updated`） |
| `--order-asc` | 升序排列（默认降序） |
| `--view-group-data` | 分组明细查询 JSON（从 `table view group-data` 响应获取） |

**过滤模式说明**：

| 模式 | 参数 | 行为 |
|------|------|------|
| 全文搜索 | `--text '关键词'` | MySQL 模式：LIKE 模糊匹配 data 字段 |
| 字段精确匹配 | `--filter '{"字段名":"值"}'` | PG 模式：`data->>'字段名' = '值'`<br>MySQL 模式：`JSON_EXTRACT(data, '$."字段名"') = '值'` |
| 视图过滤 | `--view-id 123` | 加载已保存的视图过滤/排序配置 |

**后端支持的过滤操作符**（视图过滤使用）：

| 操作符 | 说明 | 示例 |
|--------|------|------|
| `eq` | 等于 | `{"comparator":"eq","columnId":"1","value":"系统订单"}` |
| `neq` | 不等于 | `{"comparator":"neq","columnId":"1","value":"测试"}` |
| `gt` | 大于（数值） | `{"comparator":"gt","columnId":"17","value":"10"}` |
| `gte` | 大于等于（数值） | `{"comparator":"gte","columnId":"17","value":"10"}` |
| `lt` | 小于（数值） | `{"comparator":"lt","columnId":"17","value":"100"}` |
| `lte` | 小于等于（数值） | `{"comparator":"lte","columnId":"17","value":"100"}` |
| `ct` | 包含（文本） | `{"comparator":"ct","columnId":"1","value":"订单"}` |
| `nct` | 不包含（文本） | `{"comparator":"nct","columnId":"1","value":"测试"}` |
| `empty` | 为空 | `{"comparator":"empty","columnId":"1"}` |
| `notempty` | 不为空 | `{"comparator":"notempty","columnId":"1"}` |

**示例**:
```bash
# 基础查询（返回第 1 页，20 条）
lingtong-cli table data query --basic-data-id 1553

# 全文搜索（LIKE 模糊匹配）
lingtong-cli table data query --basic-data-id 1553 --text '系统订单'

# 字段级精确匹配（注意：MySQL 模式下 JSON_EXTRACT 可能有兼容性问题）
lingtong-cli table data query --basic-data-id 1553 --filter '{"1":"系统订单"}'

# 使用视图过滤（加载已保存的过滤/排序配置）
lingtong-cli table data query --basic-data-id 1553 --view-id 456

# 按记录 ID 筛选
lingtong-cli table data query --basic-data-id 1553 --ids 1001,1002,1003

# 按更新时间范围过滤
lingtong-cli table data query --basic-data-id 1553 \
  --start-update-time 1700000000000 --end-update-time 1700100000000

# 分页 + 排序
lingtong-cli table data query --basic-data-id 1553 \
  --page 2 --page-size 50 --order-by created --order-asc

# 组合过滤 + 分页 + 排序
lingtong-cli table data query --basic-data-id 1553 \
  --text '系统订单' --page 1 --page-size 50 --order-by created --order-asc

# 分组明细查询（先通过 table view group-data 获取分组信息）
lingtong-cli table data query --basic-data-id 1568 --view-id 621 \
  --view-group-data '{"esKey":"data1568.1","key":"测试","leafId":"#data1568.1@测试"}'
```

### 视图管理

```bash
# 创建视图
lingtong-cli table view save --schema-id <id> --name <name> [--type basic|pivot] [--group-column <fieldId>] [--group-asc]

# 列出视图
lingtong-cli table view list --schema-id <id>

# 更新视图（含分组配置）
lingtong-cli table view update --id <viewId> --schema-id <id> [--name <name>] [--group-column <fieldId>] [--group-asc]

# 删除视图
lingtong-cli table view delete --id <viewId>

# 获取分组数据
lingtong-cli table view group-data --view-id <viewId>

# 获取统计指标
lingtong-cli table view merits --schema-id <id> [--view-id <viewId>] [--text <string>]
```

**分组数据工作流**：
1. `table view group-data --view-id 621` 获取第一层分组
2. 从响应中获取 `esKey`、`key`、`leafId`
3. `table data query --view-group-data '{"esKey":"...","key":"...","leafId":"..."}'` 获取分组内明细
4. 分组最多支持三层嵌套

**示例**:
```bash
# 创建带分组的视图
lingtong-cli table view save --schema-id 2367 --name "按单据类型分组" --group-column 1

# 查看视图列表
lingtong-cli table view list --schema-id 2367

# 获取分组数据
lingtong-cli table view group-data --view-id 621

# 获取统计指标
lingtong-cli table view merits --schema-id 2367 --view-id 621

# 查询某个分组下的明细
lingtong-cli table data query --basic-data-id 1568 --view-id 621 \
  --view-group-data '{"esKey":"data1568.1","key":"系统订单","leafId":"#data1568.1@系统订单"}' \
  --page-size 10
```

### 数据透视表

```bash
# 查询透视表数据
lingtong-cli table pivot query --table-id <id> --view-id <id>

# 获取透视表配置
lingtong-cli table pivot config --table-id <id> --view-id <id>

# 保存透视表配置
lingtong-cli table pivot config-save --table-id <id> --view-id <id> \
  --business-id <id> --business-type <n> --config <json>
```

**参数说明**：
| 参数 | 说明 |
|------|------|
| `--table-id` | 表格 ID（必填） |
| `--view-id` | 视图 ID（必填） |
| `--business-id` | 业务 ID（必填） |
| `--business-type` | 业务类型（默认 2） |
| `--config` | 配置 JSON（必填） |

**配置结构**：
```json
{
  "dimensions": [
    {"id": "dim1", "field": "data1568.1", "alias": "单据类型", "disabled": false}
  ],
  "measures": [
    {"id": "init", "type": "count", "disabled": false}
  ],
  "advanced": {
    "openDimensionSplit": false,
    "openStatistic": true,
    "displayMode": "normal"
  },
  "other": {
    "maxRowsPerPage": 50,
    "showSummary": true,
    "otherExtractConfigVOList": [
      {"id": "init", "type": "measure", "summaryType": "count", "disabled": false}
    ]
  }
}
```

**示例**:
```bash
# 查询透视表
lingtong-cli table pivot query --table-id 1568 --view-id 621

# 获取透视表配置
lingtong-cli table pivot config --table-id 1568 --view-id 621

# 保存透视表配置（按单据类型分组，统计数量）
lingtong-cli table pivot config-save --table-id 1568 --view-id 621 \
  --business-id 1568 --business-type 2 \
  --config '{"dimensions":[{"id":"dim1","field":"data1568.1","alias":"单据类型","disabled":false}],"measures":[{"id":"init","type":"count","disabled":false}],"advanced":{"openDimensionSplit":false,"openStatistic":true,"displayMode":"normal"},"other":{"maxRowsPerPage":50,"showSummary":true,"otherExtractConfigVOList":[{"id":"init","type":"measure","summaryType":"count","disabled":false}]}}'
```

### 创建记录

```bash
lingtong-cli table data create --basic-data-id <id> --schema-id <id> --data <json> [--version <n>]
```

**参数说明**:
- `--basic-data-id`: 表格 ID（必填）
- `--schema-id`: Schema ID（必填，从 `table data query` 响应的 `schemaId` 字段获取）
- `--data`: 记录数据，JSON 格式，key 为字段 ID（字符串），value 为字段值（必填）
- `--version`: Schema 版本号（可选，不指定则跳过版本校验）

**示例**:
```bash
# 创建记录（字段 ID 为数字字符串）
lingtong-cli table data create --basic-data-id 1553 --schema-id 2189 --data '{"0":"测试","1":"2024-01-01","2":"产品A"}'

# 指定版本
lingtong-cli table data create --basic-data-id 1553 --schema-id 2189 --data '{"0":"测试"}' --version 5
```

### 批量更新记录

```bash
lingtong-cli table data batch-update --basic-data-id <id> --schema-id <id> --records <json> [--version <n>]
```

**参数说明**:
- `--basic-data-id`: 表格 ID（必填）
- `--schema-id`: Schema ID（必填）
- `--records`: 更新记录 JSON 数组（必填），每条记录包含：
  - `id`: 记录 ID（必填，从 `table data query` 响应获取）
  - `data`: 字段数据，JSON Map，key 为字段 ID（必填）
- `--version`: Schema 版本号（可选，自动填充到每条记录）

**示例**:
```bash
# 批量更新两条记录
lingtong-cli table data batch-update --basic-data-id 1633 --schema-id 2546 \
  --records '[{"id":33832272,"data":{"1":"新值1","2":10}},{"id":33832273,"data":{"1":"新值2","2":20}}]'

# 带版本号批量更新
lingtong-cli table data batch-update --basic-data-id 1633 --schema-id 2546 --version 1 \
  --records '[{"id":33832272,"data":{"1":"更新"}}]'
```

**注意事项**:
- 每条记录必须包含 `id` 和 `data`
- `basicDataId` 和 `schemaId` 会自动从命令行参数填充
- 日期字段使用毫秒时间戳存储

### 删除记录

```bash
lingtong-cli table data delete --schema-id <id> --id <id>
```

**参数说明**:
- `--schema-id`: Schema ID（必填）
- `--id`: 记录 ID（必填）

**示例**:
```bash
# 删除单条记录
lingtong-cli table data delete --schema-id 2546 --id 33832272
```

### 批量删除记录

```bash
lingtong-cli table data batch-delete --schema-id <id> --ids <ids>
```

**参数说明**:
- `--schema-id`: Schema ID（必填）
- `--ids`: 逗号分隔的记录 ID 列表（必填）

**示例**:
```bash
# 批量删除多条记录
lingtong-cli table data batch-delete --schema-id 2546 --ids "33832272,33832273,33832274"
```

**注意事项**:
- 删除操作不可逆，请谨慎使用
- 批量删除支持一次删除多条记录
- 删除后数据可通过 rollback 接口恢复（需后端支持）

### 查询表格 Schema

Schema 包含表格的字段定义（字段 ID、名称、类型等），是 AI 理解表格数据结构的关键。

```bash
lingtong-cli table schema query --basic-data-id <id> [--business-type <n>]
```

**参数说明**:
- `--basic-data-id`: 表格 ID（必填）
- `--business-type`: 业务类型，1-场景，2-基础资料（默认 2）

**示例**:
```bash
# 查询基础资料的 Schema
lingtong-cli table schema query --basic-data-id 1553

# 查询场景的 Schema
lingtong-cli table schema query --basic-data-id 123 --business-type 1
```

### 更新表格 Schema

更新表格 Schema 支持两种模式：

#### 模式 1：简单模式（推荐）

只需提供字段定义数组，版本号自动获取：

```bash
lingtong-cli table schema update --schema-id <id> --basic-data-id <id> --columns-schema <json>
```

**参数说明**:
- `--schema-id`: Schema ID（必填）
- `--basic-data-id`: 表格 ID（必填，用于自动获取当前版本号）
- `--columns-schema`: 字段定义 JSON 数组（必填）

**示例**:
```bash
# 更新字段定义（自动获取版本号）
lingtong-cli table schema update --schema-id 2546 --basic-data-id 1633 \
  --columns-schema '[{"id":1,"title":"名称","type":"text","key":"name","primaryKey":true,"show":true,"width":200,"align":"left"}]'
```

#### 模式 2：完整模式

提供完整的 Schema JSON，包含版本号：

```bash
lingtong-cli table schema update --schema-id <id> --schema <json>
```

**参数说明**:
- `--schema-id`: Schema ID（必填）
- `--schema`: 完整 Schema JSON（必填，需包含 `columnsSchema` 和 `version`）

**示例**:
```bash
# 完整模式（需指定版本号）
lingtong-cli table schema update --schema-id 2546 \
  --schema '{"columnsSchema":[{"id":1,"title":"名称","type":"text","key":"name"}],"version":2}'
```

**注意事项**:
- 后端使用乐观锁机制，`version` 必须与数据库当前版本一致，否则会报 "表格列已被其他人更新"
- 使用简单模式（`--basic-data-id`）可自动获取版本号，避免版本冲突
- 更新成功后，版本号会自动 +1
- `columnsSchema` 必须包含完整的字段定义，不支持部分更新

---

## 字段定义完整指南

### 字段类型总览

绫通表格支持以下 9 种字段类型（来源：前端 `constants.js` 的 `typeList`）：

| type | 中文名称 | 说明 | 数据存储格式 |
|------|----------|------|-------------|
| `text` | 单行文本 | 短文本，如名称、编号 | 字符串 `"产品A"` |
| `textarea` | 多行文本 | 长文本，如描述、备注 | 字符串 `"多行内容"` |
| `number` | 数字 | 整数或小数 | 数字 `100` 或字符串 `"100"` |
| `password` | 加密文本 | 敏感数据，前端显示为 `***` | 字符串（后端加密存储） |
| `date` | 日期 | 日期或日期时间 | 毫秒时间戳 `1735689600000` |
| `table` | 子表格 | 嵌套表格，包含子字段 | 嵌套对象 |
| `function` | 公式 | 计算字段，引用其他字段 | 自动计算（需高性能模式） |
| `createDate` | 创建日期 | 系统字段，记录创建时间 | 毫秒时间戳（只读） |
| `updateDate` | 最后更新日期 | 系统字段，记录更新时间 | 毫秒时间戳（只读） |

### 字段属性详解

每个字段定义包含以下属性（来源：后端 `ColumnsSchema.java`）：

| 属性 | 类型 | 必填 | CLI 默认值 | 说明 |
|------|------|------|-----------|------|
| `id` | Integer | 是 | 自动分配（从 1 开始） | 字段唯一 ID，用于数据写入时的 key |
| `title` | String | 是 | - | 字段名称，如 `"订单日期"` |
| `type` | String | 是 | - | 字段类型，见上表 |
| `key` | String | 是 | - | 字段键名，如 `"orderDate"` |
| `primaryKey` | Boolean | 否 | 第一列 `true`，其他 `false` | 是否为主键 |
| `show` | Boolean | 否 | `true` | **是否在前端显示** |
| `width` | Integer | 否 | `200` | 列宽（像素） |
| `align` | String | 否 | `"left"` | 对齐方式：`left`/`center`/`right` |
| `fixed` | String | 否 | `"left"` | 固定位置：`left`/`right`/`null` |
| `description` | String | 否 | `""` | 字段描述 |
| `dateFormat` | String | 否 | `"YYYY-MM-DD"` | 日期格式（type=date 时使用） |
| `numberFormat` | String | 否 | `"1000"` | 数字格式（type=number 时使用） |
| `function` | String | 否 | `""` | 公式表达式（type=function 时使用） |
| `functionFieldType` | String | 否 | `"normal"` | 公式返回类型：`normal`/`text`/`textarea`/`number`/`date` |
| `keyDisabled` | Boolean | 否 | `false` | key 是否可修改 |
| `typeDisabled` | Boolean | 否 | `true` | type 是否可修改 |
| `dataDisabled` | Boolean | 否 | `false` | 数据是否可编辑（系统字段为 `true`） |
| `children` | Array | 否 | - | 子表格字段定义（type=table 时使用） |

### 日期格式选项

`dateFormat` 支持以下值（来源：前端 `constants.js` 的 `dateFormatList`）：

| 格式值 | 显示效果 |
|--------|----------|
| `YYYY-MM-DD` | 2024-09-30 |
| `YYYY/MM/DD` | 2024/09/30 |
| `MM-DD` | 09-30 |
| `MM/DD` | 09/30 |
| `YYYY-MM-DD HH:mm` | 2024-09-30 14:00 |
| `YYYY/MM/DD HH:mm` | 2024/09/30 14:00 |
| `YYYY-MM-DD HH:mm:ss` | 2024-09-30 14:00:00 |
| `YYYY/MM/DD HH:mm:ss` | 2024/09/30 14:00:00 |

> ⚠️ **注意**: 日期格式必须使用大写字母 `YYYY`、`MM`、`DD`，小写会导致显示异常。

### 数字格式选项

`numberFormat` 支持以下值（来源：前端 `constants.js` 的 `numberFormatList`）：

| 格式值 | 显示效果 |
|--------|----------|
| `1000` | 整数 |
| `1,000` | 千分位 |
| `1,000.00` | 千分位（两位小数） |
| `1.0` | 保留一位小数 |
| `1.00` | 保留两位小数 |
| `1.000` | 保留三位小数 |
| `100%` | 百分比 |
| `100.00%` | 百分比（两位小数） |

### 各字段类型创建示例

#### 1. 单行文本 (text)
```json
{"title": "产品名称", "key": "productName", "type": "text"}
```

#### 2. 多行文本 (textarea)
```json
{"title": "备注", "key": "remark", "type": "textarea"}
```

#### 3. 数字 (number)
```json
{"title": "数量", "key": "quantity", "type": "number", "numberFormat": "1,000"}
```

#### 4. 加密文本 (password)
```json
{"title": "密钥", "key": "secretKey", "type": "password"}
```

#### 5. 日期 (date)
```json
{"title": "订单日期", "key": "orderDate", "type": "date", "dateFormat": "YYYY-MM-DD"}
```

#### 6. 子表格 (table)
```json
{
  "title": "订单明细",
  "key": "orderItems",
  "type": "table",
  "children": [
    {"id": 1, "title": "商品名", "key": "itemName", "type": "text"},
    {"id": 2, "title": "数量", "key": "qty", "type": "number"}
  ]
}
```

#### 7. 公式字段 (function)
> 需要开启高性能模式 `--open-high-mode 1`

```json
{
  "title": "总金额",
  "key": "totalAmount",
  "type": "function",
  "function": "[单价]*[数量]",
  "functionFieldType": "number"
}
```

#### 8. 系统字段 - 创建日期 (createDate)
```json
{
  "title": "创建时间",
  "key": "lt_sys_created",
  "type": "createDate",
  "dateFormat": "YYYY-MM-DD HH:mm:ss",
  "dataDisabled": true
}
```

#### 9. 系统字段 - 最后更新日期 (updateDate)
```json
{
  "title": "更新时间",
  "key": "lt_sys_updated",
  "type": "updateDate",
  "dateFormat": "YYYY-MM-DD HH:mm:ss",
  "dataDisabled": true
}
```

### 完整创建示例

创建一个包含多种字段类型的订单表格：

```bash
lingtong-cli table create --app-id 165 --name "订单表" --source 1 --type 1 --open-high-mode 1 \
  --columns-schema '[
    {"title":"订单编号","key":"orderNo","type":"text"},
    {"title":"客户名称","key":"customerName","type":"text"},
    {"title":"订单金额","key":"amount","type":"number","numberFormat":"1,000.00"},
    {"title":"下单日期","key":"orderDate","type":"date","dateFormat":"YYYY-MM-DD"},
    {"title":"备注","key":"remark","type":"textarea"},
    {"title":"创建时间","key":"lt_sys_created","type":"createDate","dateFormat":"YYYY-MM-DD HH:mm:ss","dataDisabled":true},
    {"title":"更新时间","key":"lt_sys_updated","type":"updateDate","dateFormat":"YYYY-MM-DD HH:mm:ss","dataDisabled":true}
  ]'
```

---

## 数据模型

### 表格数据查询响应

```json
{
  "code": 10000,
  "success": true,
  "result": {
    "data": [
      {
        "id": 21512655,
        "basicDataId": 1553,
        "schemaId": 2189,
        "primaryValue": "1",
        "data": {
          "0": "1",
          "1": "1735689600000",
          "2": "产品 A",
          "3": "15100"
        },
        "dataFrom": {
          "0": "1",
          "1": "1735689600000",
          "2": "产品 A"
        },
        "createUser": "刘明剑",
        "created": 1773107179000,
        "updated": 1773108469514,
        "version": null
      }
    ]
  }
}
```

**关键说明**:
- `data` 字段是 `Map<Integer, Object>`，key 是字段 ID（数字），value 是字段值
- 字段 ID 与 Schema 中 `columnsSchema[].id` 对应
- `dataFrom` 是相同数据的字符串表示
- 日期字段存储为毫秒时间戳

## 最佳实践

1. **先查 Schema 再查数据**：使用 `table schema query` 了解字段结构，再查询数据
2. **使用 `--basic-data-id` 而非 `--table-id`**：参数名已更新
3. **创建记录时必须指定 `--schema-id`**：从 `table data query` 响应中获取
4. **版本校验**：不指定 `--version` 时跳过校验（推荐），指定时需与 Schema 的 version 匹配
5. **字段 ID 是数字字符串**：`--data` 中的 key 使用字符串格式的数字（如 `"0"`, `"1"`）
6. **字段默认可见**：CLI 会自动设置 `show: true`，无需手动指定
7. **公式字段需高性能模式**：使用 `type: "function"` 时必须 `--open-high-mode 1`
8. **系统字段只读**：`createDate`/`updateDate` 设置 `dataDisabled: true`

## 相关技能

- [lingtong-pivot-table](../lingtong-pivot-table/SKILL.md) - 绫通表格数据透视表
