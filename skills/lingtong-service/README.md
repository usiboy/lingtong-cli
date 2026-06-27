# lingtong-service 技能

`lingtong-service` 是 OpenAPI 自动生成命令层，用来覆盖没有专用命令的长尾 API。高频场景优先使用 `app`、`connector`、`scene`、`workflow`、`table` 等专用命令；专用命令缺失时再使用 `service`。

## 何时使用

- 需要调用某个没有专用命令的绫通平台 API
- 需要按模块浏览平台所有可用接口
- 已通过 `schema` 找到路径，但不想手写 `api <method> <path>`
- 需要访问 `/basicdata/*`、`/factory/*`、`/meta/*`、`/noco/*` 等长尾接口

## 快速开始

```bash
# 查看全部自动生成模块
lingtong-cli service --help

# 查看某个模块的操作
lingtong-cli service scene --help

# 调用自动生成命令
lingtong-cli service scene list --pageNum 1 --pageSize 10

# body 参数可以用 --body JSON 兜底
lingtong-cli service basicdata save --body '{"name":"test"}'
```

## 与 schema 配合

```bash
# 搜索接口
lingtong-cli schema search 场景

# 查看模块接口
lingtong-cli schema module scene

# 查看路径细节
lingtong-cli schema path /scene/list
```

`service` 与 `schema` 使用同一份 OpenAPI 规格：运行时 override -> `~/.lingtong-cli/openapi.json` -> 二进制内置规格。

## 参数规则

| OpenAPI 参数位置 | CLI 表达 |
|------------------|----------|
| query | 同名 flag，如 `--pageNum 1` |
| path | 同名必填 flag，自动替换 `{param}` |
| body 简单对象 | 尽量展开为字段 flag |
| body 复杂对象 | 使用 `--body '<json>'` |

## 命令层选择

| 需求 | 推荐命令 |
|------|----------|
| 高频业务操作 | 专用命令，如 `scene list`、`table data query` |
| OpenAPI 长尾接口 | `service <module> <operation>` |
| 完全自定义 method/path | `api <method> <path>` |
| 不确定接口位置 | `schema search <keyword>` |

## 注意事项

- `service` 命令数量会随内置 OpenAPI 规格变化，具体数量以 `lingtong-cli schema list` 为准。
- `basicdata` 独立手写命令已废弃，但 OpenAPI 模块中仍可能存在 `service basicdata ...`。
- 写入和删除类长尾接口调用前，先用 `schema path` 确认 method、body 和风险。

## 相关文档

- [SKILL.md](SKILL.md) - Agent 使用指令
- `lingtong-shared` - 配置、认证、输出契约
- `lingtong-table` - 基础资料/表格的专用命令入口
