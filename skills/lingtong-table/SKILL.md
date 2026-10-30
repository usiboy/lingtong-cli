---
name: lingtong-table
version: 2.12.0
description: "绫通表格管理：表格 CRUD、数据查询、记录创建、批量创建、单条更新、批量更新、记录计数、删除（带确认）、Schema 与公式字段管理。当用户需要操作绫通表格、查询数据、创建记录、批量导入、更新记录、统计记录数、删除记录、理解表格结构、添加公式或计算列时触发。关键词：table、表格、table data、query、create record、batch-create、update、count、batch-update、delete、schema、公式、function。"
---

# lingtong-table 技能

## 概述

本技能指导你如何通过 `lingtong-cli` 管理绫通平台的表格数据。

**公式/计算列入口**：需要金额、条件判断、日期差、文本或数组公式时，继续阅读 [lingtong-table-formula](../lingtong-table-formula/SKILL.md)，其中包含经过源码核对的字段契约、完整 Schema 更新流程及 ES/Painless 限制。

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

### 更新表格配置

```bash
lingtong-cli table update --id <id> \
  [--name <name>] \
  [--open-high-mode <0|1>] \
  [--open-connector <0|1>]
```

**参数说明**：
| 参数 | 说明 |
|------|------|
| `--id` | 表格 ID（必填） |
| `--name` | 新表格名称 |
| `--open-high-mode` | 高性能模式：0-关闭，1-开启 |
| `--open-connector` | 连接器同步：0-关闭，1-开启 |

**说明**：
- 命令会自动获取当前表格配置，仅更新指定的字段
- 开启连接器后，表格可在场景和工作流中作为连接器使用
- 高性能模式会限制表格 Schema 类型，提升查询性能

**示例**:
```bash
# 开启连接器
lingtong-cli table update --id 1568 --open-connector 1

# 开启高性能模式
lingtong-cli table update --id 1568 --open-high-mode 1

# 更新表格名称
lingtong-cli table update --id 1568 --name "新名称"

# 同时开启连接器和高性能模式
lingtong-cli table update --id 1568 --open-connector 1 --open-high-mode 1
```

### 查询表格数据

```bash
lingtong-cli table data query --basic-data-id <id> \
  [--text <string>] [--filter <json>] [--view-id <id>] \
  [--id <record-id>] [--ids <csv>] [--primary <value>] \
  [--start-update-time <ms>] [--end-update-time <ms>] \
  [--primary-key <json-array>] \
  [--page <n>] [--page-size <n>] \
  [--order-by <field>] [--order-asc] [--omit-null=false]
```

**参数说明**：
| 参数 | 说明 |
|------|------|
| `--basic-data-id` | 表格 ID（必填） |
| `--text` | 全文搜索（MySQL 模式：LIKE 模糊匹配） |
| `--filter` | 字段级精确匹配（JSON：`{"fieldKey":"value"}`） |
| `--view-id` | 视图 ID（加载已保存的过滤/排序配置） |
| `--id` | 按单条记录 ID 精确筛选 |
| `--ids` | 按记录 ID 列表筛选，逗号分隔（如 `1001,1002`） |
| `--primary` | 按单个业务主键值精确筛选 |
| `--start-update-time` | 更新时间起始（毫秒时间戳） |
| `--end-update-time` | 更新时间截止（毫秒时间戳） |
| `--primary-key` | 按主键值筛选（JSON 数组） |
| `--page` | 页码（默认 1） |
| `--page-size` | 每页条数（默认 20） |
| `--order-by` | 排序字段（如 `created`、`updated`） |
| `--order-asc` | 升序排列（默认降序） |
| `--omit-null=false` | 保留 null 字段；全字段更新前必须使用，避免读改写时丢字段 |
| `--view-group-data` | 分组明细查询 JSON（从 `table view group-data` 响应获取） |

在包含“基础资料精确条件查询直接读取数据库”改动的服务端，`--primary` 非空白、`--primary-key` 或 `--ids` 非空，或指定 `--id` 时，记录列表查询绕过 ES，按表存储类型直接查询 MySQL 或 PostgreSQL。多个条件同时传入时共同生效；空数组和空白主键不会触发该路由。其他条件仍按原有路由处理，目标环境需部署对应服务端版本。

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
lingtong-cli table data query --basic-data-id 1553 --id 1001

# 按业务主键筛选
lingtong-cli table data query --basic-data-id 1553 --primary '订单-001'
lingtong-cli table data query --basic-data-id 1553 --primary-key '["订单-001","订单-002"]'

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

### 更新单条记录

```bash
lingtong-cli table data update --basic-data-id <id> --schema-id <id> --id <record-id> --data <json> [--version <n>]
```

**参数说明**:
- `--basic-data-id`: 表格 ID（必填）
- `--schema-id`: Schema ID（必填）
- `--id`: 记录 ID（必填）
- `--data`: 完整记录数据，JSON 对象，key 为字段 ID；必须包含业务主键（必填）
- `--version`: Schema 版本号（可选，不指定则跳过版本校验）

> **重要**：当前更新接口会替换整条记录的 `data`。未提供的字段可能被清空，不能直接用稀疏 JSON 做局部更新。应先查询原记录（带 `--omit-null=false`），保留完整 `data` 和 `updated`，修改目标字段后再提交。

**示例**:
```bash
# 读取完整记录后再更新，data 中包含主键和其余原字段
lingtong-cli table data update --basic-data-id 123 --schema-id 1 --id 33832272 \
  --data '{"1":"PK-001","2":"新值","3":10}'

# 带版本号更新
lingtong-cli table data update --basic-data-id 123 --schema-id 1 --id 33832272 --data '{"1":"新值"}' --version 5
```

### 批量创建或按主键更新记录

灰度版 `gray/1.1.50` 新增 `/basicdata/record/batchSave`；CLI 用一次请求提交 JSON 数组：

```bash
lingtong-cli table data batch-create --basic-data-id <id> --schema-id <id> \
  (--records '<json-array>' | --records-file <json-file>) [--version <n>]
```

- 每条记录需提供非空 `data` 对象，字段 ID 为数字字符串；`basicDataId`、`schemaId` 自动填充。
- 可省略 `version`；指定 `--version` 时仅为未设置版本的记录填充。响应原样返回服务端的记录列表。
- Excel 数据需先转换为 JSON 数组文件（示例结构：`[{"data":{"1":"产品A","2":10}},{"data":{"1":"产品B","2":20}}]`），`--records-file` 读取该文件；本命令不直接读取 `.xlsx`。
- **主键匹配已有记录会执行全字段更新**，未提供的字段可能被清空；同一批次重复主键会失败。提交前检查主键是否重复，并先用少量数据验证字段映射。单次命令提交一批，无自动拆批。
- CLI 不提供自动拆批、重试、进度条或断点。大数据量导入应由调用脚本单线程顺序提交：先小批量试写，再根据服务端限制调整批次；仅在批次成功后保存断点。
- 使用确定性业务主键保证重跑幂等。Schema 查询、记录查询和批量写入必须使用相同的 `--auth`、`--profile` 或 `LINGTONG_TOKEN`，避免旧 legacy Token 写入错误组织。
- 遇到 HTTP 500、信号量耗尽或数据库连接池超时时应退避重试，不要通过增加并发规避。服务端返回 `success:false` 时 CLI 会以错误退出。

**大批量导入执行契约**：

1. 只选择一种认证身份：固定 `--auth`、固定 `--profile` 中的身份，或 `LINGTONG_TOKEN`。不要同时依赖多个可能冲突的 Token 来源；写入前用同一身份查询并核对目标 Schema、主键和组织。
2. 先用约 100 条做 canary；成功后可尝试 500～1000 条。没有通用最大批次，受宽表字段数和请求体大小影响；失败时减半，不要增加并发。
3. 每批使用不可变 JSON 文件和确定性主键。瞬时错误最多重试 5～8 次，指数退避并封顶约 15 秒；参数、Schema、主键重复等永久错误立即停止。
4. 超时后批次状态视为“不确定”，使用相同完整数据和确定性主键重放；不要生成新主键。
5. 仅在 CLI 成功退出且响应 `success:true` 后原子写入断点。CLI 会在服务端返回记录数组时核对数量；若服务端不返回数组，调用脚本仍需在批后或最终对账。断点至少保存 Host/认证标签、表格 ID、Schema ID/版本、输入文件哈希、最后成功来源位置和累计成功数。
6. 进度至少显示 `已确认来源行/总行数`、累计/近期行速、ETA、重试次数和最后成功断点；不要把扫描速度误称为数据库写入速度。
7. 任务重启前重新核对 Host、身份、Schema 和输入哈希；同一断点文件只能有一个写入进程。
8. 不假设批次具有事务原子性。任务结束后核对总数，并按主键抽样或分段查询；失败或超时批次先查询目标主键，再决定重放。

```bash
lingtong-cli table data batch-create --basic-data-id 1633 --schema-id 2546 \
  --records '[{"data":{"1":"产品A","2":10}},{"data":{"1":"产品B","2":20}}]'

lingtong-cli table data batch-create --basic-data-id 1633 --schema-id 2546 \
  --version 5 --records-file ./records.json
```

### 批量更新记录

```bash
lingtong-cli table data batch-update --basic-data-id <id> --schema-id <id> \
  (--records <json> | --records-file <json-file>) [--version <n>]
```

**参数说明**:
- `--basic-data-id`: 表格 ID（必填）
- `--schema-id`: Schema ID（必填）
- `--records` / `--records-file`: 二选一，提供更新记录 JSON 数组。每条记录包含：
  - `id`: 记录 ID（必填，从 `table data query` 响应获取）
  - `data`: 完整字段数据，JSON Map，必须包含业务主键（必填）
  - `updated`: 查询结果中的更新时间，建议保留以进行并发修改校验
- `--version`: Schema 版本号（可选，自动填充到每条记录）

> **重要**：`batch-update` 同样是全字段替换。更新前使用 `table data query --omit-null=false` 读取完整记录，在原 `data` 上修改字段；省略字段可能导致数据丢失。
> `batch-update` 不自动拆批。大量宽表记录应拆成多个 JSON 文件，先小批试写，每批成功后再推进断点。超时批次使用相同 `id`、`updated` 和完整 `data` 重新核对或重放。

**示例**:
```bash
# 批量更新两条记录
lingtong-cli table data batch-update --basic-data-id 1633 --schema-id 2546 \
  --records '[{"id":33832272,"updated":1773108469514,"data":{"1":"PK-1","2":"新值1","3":10}},{"id":33832273,"updated":1773108469514,"data":{"1":"PK-2","2":"新值2","3":20}}]'

# 带版本号批量更新
lingtong-cli table data batch-update --basic-data-id 1633 --schema-id 2546 --version 1 \
  --records-file ./updates.json
```

**注意事项**:
- 每条记录必须包含 `id` 和完整 `data`，其中包含业务主键字段
- `basicDataId` 和 `schemaId` 会自动从命令行参数填充
- 日期字段使用毫秒时间戳存储
- 宽表或大量记录优先使用 `--records-file`，避免命令行参数长度限制

### 统计记录数

```bash
lingtong-cli table data count --schema-id <id> --version <n> [--basic-data-id <id>] [--filter <json>]
```

**参数说明**:
- `--schema-id`: Schema ID（必填）
- `--version`: 数据版本号（必填）
- `--basic-data-id`: 表格 ID（可选）
- `--filter`: 字段级精确匹配 JSON（可选）

**示例**:
```bash
# 统计记录总数
lingtong-cli table data count --schema-id 2546 --version 1

# 按条件统计
lingtong-cli table data count --schema-id 2546 --version 1 --filter '{"1":"系统订单"}'
```

### 删除记录

删除操作是破坏性的，需要 `--yes` 确认。

```bash
lingtong-cli table data delete --schema-id <id> --id <id> --yes
```

**参数说明**:
- `--schema-id`: Schema ID（必填）
- `--id`: 记录 ID（必填）
- `--yes`: 确认破坏性操作（必填）

**示例**:
```bash
# 删除单条记录（需要 --yes）
lingtong-cli table data delete --schema-id 2546 --id 33832272 --yes

# 不加 --yes 将返回退出码 10（需要确认）
lingtong-cli table data delete --schema-id 2546 --id 33832272
```

### 批量删除记录

批量删除同样需要 `--yes` 确认。

```bash
lingtong-cli table data batch-delete --schema-id <id> --ids <ids> --yes
```

**参数说明**:
- `--schema-id`: Schema ID（必填）
- `--ids`: 逗号分隔的记录 ID 列表（必填）
- `--yes`: 确认破坏性操作（必填）

**示例**:
```bash
# 批量删除多条记录（需要 --yes）
lingtong-cli table data batch-delete --schema-id 2546 --ids "33832272,33832273,33832274" --yes
```

**注意事项**:
- 删除操作不可逆，请谨慎使用
- 删除和批量删除都需要 `--yes` 确认，防止误操作
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

**两种模式都完整替换 `columnsSchema`**。先读取并保留原有完整字段数组，只在原数组上合并改动。下方单列 JSON 仅展示参数结构，不能直接用于已有多列表格；自动取版本不等于自动合并列，涉及并发修改时优先携带读取时的 `version`。

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

下表的“CLI 默认值”由 `table create --columns-schema` 补齐；`table schema update` 不补齐新列默认属性，新增公式列等字段时显式设置 `show`、`primaryKey` 和显示配置。

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
> 面向 ES 查询的公式先开启高性能模式 `--open-high-mode 1`，并确认租户 ES 查询与同步已启用。新表先建立普通源字段及映射，再通过完整 Schema 更新追加公式列。

```json
{
  "id": 4,
  "title": "总金额",
  "key": "totalAmount",
  "type": "function",
  "function": "ROUND([单价]*[数量],2)",
  "functionFieldType": "number",
  "numberFormat": "1.00",
  "show": true,
  "primaryKey": false,
  "dataDisabled": true
}
```

公式按字段 **title** 引用，记录 `data` 才使用字段 ID。上例是新增列片段，更新时必须保留原有完整 `columnsSchema` 和读取时的 `version`；不能只提交公式列。`--basic-data-id` 自动取版本不会合并列。字段属性是小写 `type: "function"` 与 `function`，不是 `FORMULA`/`formula`。

函数目录、日期单位、空值和公式验证方式见 [公式与函数](../lingtong-table-formula/references/functions.md)。当前没有独立 `table formula` 命令，使用现有 `table schema query/update` 配置并查询样本核验。

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
4. **版本校验**：创建记录可按需省略；生产更新建议传入当前 Schema version，Schema 变化时立即失败而不是误写
5. **字段 ID 是数字字符串**：`--data` 中的 key 使用字符串格式的数字（如 `"0"`, `"1"`）
6. **创建时字段默认可见**：`table create` 自动设置 `show: true`；`table schema update` 新增列时需显式提供
7. **公式字段需高性能模式**：使用 `type: "function"` 时必须 `--open-high-mode 1`
8. **系统字段只读**：`createDate`/`updateDate` 设置 `dataDisabled: true`
9. **更新是全字段替换**：单条和批量更新前先用 `--omit-null=false` 查询完整记录，在原 `data` 上修改；保留业务主键和 `updated`
10. **大批导入单线程执行**：使用 `batch-create --records-file` 顺序提交，先试写再调批次；记录进度、速度和成功批次断点，不并发压测表格数据库
11. **认证必须一致**：选择一个身份来源并贯穿 Schema、查询、写入；写前核对已知 Schema/主键，避免环境 Token、`--auth` 与 legacy Token 冲突
12. **分页必须稳定推进**：CLI 查询优先使用 `--order-by id --order-asc` 与递增 `--page`，检测页码和记录 ID 重复；自定义脚本直调 `listNew` 时请求字段是 `pageNum`。页码分页期间不得删除/插入影响排序范围的记录；可变数据集应先冻结范围、记录最大 ID/截止更新时间，或先快照目标 ID，再处理该固定集合

## 扩展操作

筛选配置、Excel 导入导出、字典与连接器同步见 [表格扩展操作](references/advanced.md)。

## 相关技能

- [lingtong-table-formula](../lingtong-table-formula/SKILL.md) - 公式字段、常用函数、ES Runtime Field/Painless
- [lingtong-pivot-table](../lingtong-pivot-table/SKILL.md) - 绫通表格数据透视表
- [lingtong-kmmdtable](../lingtong-kmmdtable/SKILL.md) - 快麦多维表格资源、文件夹、仪表盘、自定义页面与权限
