# lingtong-basicdata 迁移说明

`lingtong-basicdata` 已不再作为有效 Agent Skill 内置。基础资料能力已经合并到 `lingtong-table`，请不要再为新任务加载或引用独立的 `lingtong-basicdata` 技能。

## 当前状态

- `skills/lingtong-basicdata/` 是历史遗留目录。
- 该目录没有 `SKILL.md`，因此 `lingtong-cli skills list` 不会把它列为内置 skill。
- 根命令中 `basicdata` 已标记废弃，并提示使用 `table data`。

## 迁移映射

| 旧口径 | 新口径 |
|--------|--------|
| 基础资料列表 | `lingtong-cli table list` |
| 基础资料记录查询 | `lingtong-cli table data query --basic-data-id <id>` |
| 创建记录 | `lingtong-cli table data create --basic-data-id <id> --schema-id <id> --data '<json>'` |
| 更新记录 | `lingtong-cli table data update --basic-data-id <id> --schema-id <id> --id <record-id> --data '<json>'` |
| 批量更新 | `lingtong-cli table data batch-update --basic-data-id <id> --schema-id <id> --records '<json-array>'` |
| 记录统计 | `lingtong-cli table data count --schema-id <id> --version <n>` |
| 删除记录 | `lingtong-cli table data delete --schema-id <id> --id <record-id> --yes` |
| 长尾 `/basicdata/*` API | `lingtong-cli service basicdata ...` |

## 推荐入口

- [../lingtong-table/README.md](../lingtong-table/README.md) - 表格与基础资料管理
- [../lingtong-service/README.md](../lingtong-service/README.md) - 长尾 OpenAPI 调用

## 维护说明

如后续决定彻底清理历史目录，可删除该目录；删除前请确认没有安装脚本、文档或外部引用依赖这个路径。
