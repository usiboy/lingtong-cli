---
name: lingtong-basicdata
version: 1.0.0
description: "绫通 CLI 基础资料记录操作 (basicdata record)：表格记录的增删改查、计数、批量删除,带破坏性操作确认。当用户需要查询/统计/新增/更新/删除基础资料表记录时触发。"
---

# lingtong-cli 基础资料记录操作 (basicdata)

`basicdata record` 是对基础资料表记录(`/basicdata/record/*`)的手写命令组,平台最高频的 API 面。强校验、统一 CRUD 形态,删除操作带确认。

## 核心概念

- `--basic-data-id`:基础资料表 ID(即"表格"ID)。
- `--schema-id`:该表的列结构 ID。
- 字段值用 **JSON 对象**传入,键为字段 ID:`'{"0":"名称","1":"数值"}'`。

## 命令一览

```bash
basicdata record list          # 分页查询记录(POST /basicdata/record/listNew)
basicdata record count         # 统计记录数(GET /basicdata/record/count)
basicdata record save          # 新增记录(POST /basicdata/record/save)
basicdata record update        # 更新单条记录(POST /basicdata/record/update)
basicdata record delete        # 删除单条(需 --yes)
basicdata record batch-delete  # 批量删除(需 --yes)
```

## 查询

```bash
# 基础分页
lingtong-cli basicdata record list --basic-data-id 123

# 字段精确过滤 + 分页 + 排序
lingtong-cli basicdata record list --basic-data-id 123 \
  --filter '{"1":"系统订单"}' --page 1 --page-size 50 --order-by created --order-asc

# 按记录 ID 过滤
lingtong-cli basicdata record list --basic-data-id 123 --ids 1001,1002

# 统计数量(需 schema-id + version)
lingtong-cli basicdata record count --schema-id 2546 --version 1
```

## 写入

```bash
# 新增
lingtong-cli basicdata record save --basic-data-id 123 --schema-id 1 --data '{"0":"v1","1":"v2"}'

# 更新单条(只传要改的字段)
lingtong-cli basicdata record update --basic-data-id 123 --schema-id 1 --id 33832272 --data '{"1":"newValue"}'
```

## 删除(破坏性,需确认)

删除命令默认**拒绝执行**,必须加 `--yes`,否则以**退出码 10**(需要确认)失败:

```bash
# 不加 --yes → 退出码 10,提示 re-run with --yes
lingtong-cli basicdata record delete --schema-id 2546 --id 33832272

# 确认删除单条
lingtong-cli basicdata record delete --schema-id 2546 --id 33832272 --yes

# 批量删除
lingtong-cli basicdata record batch-delete --schema-id 2546 --ids 1001,1002,1003 --yes
```

## 与 `table data` 的关系

`table data`(query/create/batch-update/delete/batch-delete)与本命令组都操作 `/basicdata/record/*`。
`basicdata record` 额外提供 **count(计数)**、**update(单条更新)**,并对删除统一加 **`--yes` 确认**。
高频日常查询两者皆可;需要计数、单条更新或确认保护时用 `basicdata`。

## 契约

- 所有命令共享统一输出契约:`--envelope`、类型化退出码、`--jq`、`--omit-null`(见 [[lingtong-shared]])。
- 缺必填参数 → 退出码 **2**;未确认的删除 → 退出码 **10**;参数 JSON 非法 → 退出码 **2**。
- 长尾接口(导入/导出/回滚等)可用自动生成层 [[lingtong-service]] 调用:`lingtong-cli service basicdata ...`。
