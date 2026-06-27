# lingtong-cli-app 技能

`lingtong-cli-app` 用于管理绫通集成应用配置，覆盖平台应用查询、本地 JSON 导出/导入、结构验证、脚手架生成、差异比较和应用内场景维护。

## 何时使用

- 需要导出线上应用配置做备份或迁移
- 需要将应用 JSON 导入到平台
- 需要校验应用 JSON 的结构、连接器引用和场景引用
- 需要从零生成应用脚手架
- 需要使用 `kuaimai-kingdee` 内置模板
- 需要比较两个应用配置文件差异
- 需要维护应用 JSON 中的场景列表

## 快速开始

```bash
# 生成模板
lingtong-cli app scaffold --name demo --source kmerp --target kingDeeCloudStar --output app.json

# 使用内置模板
lingtong-cli app scaffold --name kuaimai-demo --template kuaimai-kingdee --output app.json

# 校验并预览导入
lingtong-cli app validate --file app.json
lingtong-cli app import --file app.json --dry-run
```

## 命令清单

| 目标 | 命令 |
|------|------|
| 列出平台应用 | `lingtong-cli app list --page-num 1 --page-size 40` |
| 查询应用详情 | `lingtong-cli app get --application-id <id>` |
| 导出应用 | `lingtong-cli app export --app-id <id> --output app.json` |
| 验证应用文件 | `lingtong-cli app validate --file app.json` |
| 导入应用 | `lingtong-cli app import --file app.json --dry-run` |
| 生成脚手架 | `lingtong-cli app scaffold --name <name> --source <conn> --target <conn> --output app.json` |
| 使用模板 | `lingtong-cli app scaffold --name <name> --template kuaimai-kingdee --output app.json` |
| 比较差异 | `lingtong-cli app diff --file-a app-dev.json --file-b app-prod.json` |
| 列出文件内场景 | `lingtong-cli app scene list --file app.json` |
| 添加文件内场景 | `lingtong-cli app scene add --file app.json --name <name> --source <conn> --target <conn>` |
| 删除文件内场景 | `lingtong-cli app scene remove --file app.json --name <name>` |

## 应用 JSON 结构

```json
{
  "appName": "应用名称",
  "appConnectors": ["kmerp", "kingDeeCloudStar"],
  "basicDatas": [],
  "scenes": [],
  "workflows": [],
  "workflowConnectors": []
}
```

`validate` 会检查顶层字段、应用名称、连接器引用、场景名称、重复场景、循环依赖、字段映射和 JSON 大小限制。

## 内置模板

| 模板 | 描述 |
|------|------|
| `kuaimai-kingdee` | 快麦 ERP 与金蝶云星辰集成，包含 5 个基础资料和 7 个预配置场景 |

## 常见流程

### 导出、修改、回导

```bash
lingtong-cli app export --app-id 123 --output app-prod.json
lingtong-cli app scene list --file app-prod.json --format table
lingtong-cli app scene add --file app-prod.json --name 库存同步 --source kingDeeCloudStar --target kmerp
lingtong-cli app validate --file app-prod.json
lingtong-cli app import --file app-prod.json --dry-run
```

### 比较环境差异

```bash
lingtong-cli app diff --file-a app-dev.json --file-b app-prod.json
```

## 注意事项

- `app import --dry-run` 只读取本地文件，不要求 Host；实际导入需要 Host 和认证。
- 写入平台前必须先执行 `app validate`。
- 本技能管理的是“应用配置包”；平台运行态场景生命周期请使用 `lingtong-scene`。

## 相关文档

- [SKILL.md](SKILL.md) - 完整 Agent 指令
- `lingtong-scene` - 平台场景管理
- `lingtong-table` - 表格/基础资料
- `lingtong-workflow` - 工作流编排
