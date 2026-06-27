# lingtong-table 技能

`lingtong-table` 是表格和基础资料的主入口，覆盖表格 CRUD、Schema、记录 CRUD、批量操作、统计、视图、分组和基础透视表入口。

## 何时使用

- 创建、查询或更新表格
- 查询、创建、更新、批量更新、统计、删除表格记录
- 查询或更新表格 Schema
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

| 目标 | 命令 |
|------|------|
| 查询记录 | `lingtong-cli table data query --basic-data-id <id>` |
| 字段精确过滤 | `lingtong-cli table data query --basic-data-id <id> --filter '{"1":"系统订单"}'` |
| 全文搜索 | `lingtong-cli table data query --basic-data-id <id> --text 系统订单` |
| 创建记录 | `lingtong-cli table data create --basic-data-id <id> --schema-id <id> --data '<json>'` |
| 更新记录 | `lingtong-cli table data update --basic-data-id <id> --schema-id <id> --id <record-id> --data '<json>'` |
| 批量更新 | `lingtong-cli table data batch-update --basic-data-id <id> --schema-id <id> --records '<json-array>'` |
| 统计数量 | `lingtong-cli table data count --schema-id <id> --version <n>` |
| 删除记录 | `lingtong-cli table data delete --schema-id <id> --id <record-id> --yes` |
| 批量删除 | `lingtong-cli table data batch-delete --schema-id <id> --ids 1,2,3 --yes` |

记录数据的 key 通常是字段 ID 字符串，例如 `{"1":"值"}`。先用 `table schema query` 确认字段 ID 和字段类型。

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
- 高性能模式、连接器同步开关会影响表格行为，修改前确认环境和业务影响。

## 相关文档

- [SKILL.md](SKILL.md) - 完整命令和模式说明
- `lingtong-pivot-table` - 透视表配置
- `lingtong-script` - 表格脚本和 InfInvoker
- `lingtong-service` - 长尾 basicdata API
