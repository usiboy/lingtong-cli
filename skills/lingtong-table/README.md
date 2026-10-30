# lingtong-table 技能

`lingtong-table` 是表格和基础资料的主入口，覆盖表格 CRUD、Schema、公式字段、记录 CRUD、批量操作、统计、视图、分组和基础透视表入口。

## 何时使用

- 创建、查询或更新表格
- 查询、创建、批量创建、更新、批量更新、统计、删除表格记录
- 查询或更新表格 Schema
- 添加金额、日期差、条件判断等公式/计算列
- 管理表格视图、分组和统计指标
- 创建透视表视图或进入透视表配置流程
- 从旧 `basicdata` 口径迁移到 `table data`

## 快速开始

```bash
# 查表和 Schema
lingtong-cli table list --app-id 165
lingtong-cli table schema query --basic-data-id 1568

# 查数据
lingtong-cli table data query --basic-data-id 1568 --page 1 --page-size 20

# 创建记录
lingtong-cli table data create --basic-data-id 1568 --schema-id 2367 --data '{"1":"系统订单"}'

# 批量创建（先把 Excel 行映射成 JSON 数组）
lingtong-cli table data batch-create --basic-data-id 1568 --schema-id 2367 --records-file ./records.json
```

## 表格管理

| 目标 | 命令 |
|------|------|
| 列出表格 | `lingtong-cli table list [--app-id <id>]` |
| 创建表格 | `lingtong-cli table create --app-id <id> --name <name> --source 1 --type 1` |
| 带 Schema 创建 | `lingtong-cli table create --app-id <id> --name <name> --columns-schema '<json-array>'` |
| 更新表格配置 | `lingtong-cli table update --id <id> --open-connector 1` |
| 查询 Schema | `lingtong-cli table schema query --basic-data-id <id>` |
| 更新 Schema | `lingtong-cli table schema update --schema-id <id> --basic-data-id <id> --columns-schema '<json-array>'` |

## 数据操作

公式字段的创建、更新与核验见 [lingtong-table-formula](../lingtong-table-formula/SKILL.md)。使用 `type: "function"`、`function: "ROUND([单价]*[数量],2)"`、`functionFieldType: "number"`，按字段标题引用；通过完整 Schema 更新配置，不在记录数据中写计算结果。ES 查询还需确认高性能模式、租户查询开关和数据同步。

| 目标 | 命令 |
|------|------|
| 查询记录 | `lingtong-cli table data query --basic-data-id <id>` |
| 按记录 ID 查询 | `lingtong-cli table data query --basic-data-id <id> --id 1001` 或 `--ids 1001,1002` |
| 按业务主键查询 | `lingtong-cli table data query --basic-data-id <id> --primary '订单-001'` 或 `--primary-key '["订单-001","订单-002"]'` |
| 字段精确过滤 | `lingtong-cli table data query --basic-data-id <id> --filter '{"1":"系统订单"}'` |
| 全文搜索 | `lingtong-cli table data query --basic-data-id <id> --text 系统订单` |
| 创建记录 | `lingtong-cli table data create --basic-data-id <id> --schema-id <id> --data '<json>'` |
| 批量创建/按主键覆盖 | `lingtong-cli table data batch-create --basic-data-id <id> --schema-id <id> --records '<json-array>'` 或 `--records-file <json-file>` |
| 更新记录 | `lingtong-cli table data update --basic-data-id <id> --schema-id <id> --id <record-id> --data '<json>'` |
| 批量更新 | `lingtong-cli table data batch-update --basic-data-id <id> --schema-id <id> --records '<json-array>'` |
| 统计数量 | `lingtong-cli table data count --schema-id <id> --version <n>` |
| 删除记录 | `lingtong-cli table data delete --schema-id <id> --id <record-id> --yes` |
| 批量删除 | `lingtong-cli table data batch-delete --schema-id <id> --ids 1,2,3 --yes` |

记录数据的 key 通常是字段 ID 字符串，例如 `{"1":"值"}`。先用 `table schema query` 确认字段 ID 和字段类型。
部署了对应服务端改动后，非空白 `--primary`、非空 `--primary-key`/`--ids` 或指定 `--id` 会使记录列表查询绕过 ES，直接读取表对应的 MySQL 或 PostgreSQL；多个条件共同生效。
批量创建调用 `batchSave`，同批次主键重复会失败；与已有主键相同的数据会全字段更新，未提供的字段可能被清空。Excel 文件需先转换为 `[{"data":{"1":"值"}}]` 格式的 JSON 文件。

更新和批量更新同样会替换完整 `data`，不是安全的局部字段更新。先用 `table data query --omit-null=false` 读取完整记录，保留业务主键与 `updated`，在原 `data` 上修改后再提交。宽表或大量更新使用 `batch-update --records-file`。

大数据量导入使用单线程顺序批次：约 100 条 canary 成功后再尝试 500～1000 条；失败时减半，不增加并发。使用确定性主键和不可变批次文件，瞬时故障重试 5～8 次；仅在 `success:true` 后原子保存 Host、认证标签、表/Schema、输入哈希和最后成功位置。CLI 不自动拆批，也不提供进度或断点，调用脚本需自行实现。

认证只选一个来源并贯穿 Schema、查询和写入。全字段读改写使用稳定分页（优先按 `id` 升序），检测重复页/记录；直接调用 `listNew` 时分页字段为 `pageNum`。可变数据集先固定最大 ID、截止更新时间或目标 ID 快照，避免页码分页期间插入/删除导致漏数。批量任务完成后必须按总数和主键抽样对账，不假设批次天然具有事务原子性。

## 视图与分组

| 目标 | 命令 |
|------|------|
| 创建视图 | `lingtong-cli table view save --schema-id <id> --name <name>` |
| 创建分组视图 | `lingtong-cli table view save --schema-id <id> --name <name> --group-column <fieldId>` |
| 创建透视表视图 | `lingtong-cli table view save --schema-id <id> --name <name> --type pivot` |
| 列出视图 | `lingtong-cli table view list --schema-id <id>` |
| 更新视图 | `lingtong-cli table view update --id <view-id> --schema-id <id> --name <name>` |
| 删除视图 | `lingtong-cli table view delete --id <view-id>` |
| 查询分组数据 | `lingtong-cli table view group-data --view-id <view-id>` |
| 查询统计指标 | `lingtong-cli table view merits --schema-id <id> --view-id <view-id>` |

分组数据可用于钻取明细：从 `table view group-data` 响应中取 `esKey`、`key`、`leafId`，传给 `table data query --view-group-data '<json>'`。

## 透视表入口

透视表详细配置由 `lingtong-pivot-table` 负责：

```bash
lingtong-cli table pivot config --table-id 1568 --view-id 621
lingtong-cli table pivot config-save --table-id 1568 --view-id 621 --business-id 1568 --business-type 2 --config '<json>'
lingtong-cli table pivot query --table-id 1568 --view-id 621
```

## basicdata 迁移口径

- 独立 `basicdata` 命令已废弃。
- 常规基础资料/表格数据操作全部使用 `table data`。
- 仍未封装的 `/basicdata/*` 长尾接口使用 `lingtong-cli service basicdata ...`。

## 安全规则

- 删除和批量删除必须显式传 `--yes`。
- 大批量更新前先用小范围记录测试。
- 不要用并发写入规避超时或连接池耗尽；应降低批次并退避重试。
- 高性能模式、连接器同步开关会影响表格行为，修改前确认环境和业务影响。

## 相关文档

- [SKILL.md](SKILL.md) - 完整命令和模式说明
- [lingtong-table-formula](../lingtong-table-formula/SKILL.md) - 公式字段与函数、Schema 配置及 ES/Painless 排查
- `lingtong-pivot-table` - 透视表配置
- `lingtong-script` - 表格脚本和 InfInvoker
- `lingtong-service` - 长尾 basicdata API
