# lingtong-model 技能

`lingtong-model` 用于查询连接器模型元数据，包括接口模型、领域模型和动态模型视图。它是字段映射、连接器调用、工作流编排前的结构发现工具。

## 何时使用

- 不清楚连接器有哪些业务对象或接口模型
- 需要获取某个业务对象的领域模型字段
- 需要根据授权账号查看动态模型结构
- 为 `connector invoke` 构造请求 body 前需要确认字段
- 为场景字段映射或工作流脚本准备字段元数据

## 快速开始

```bash
# 查询接口模型
lingtong-cli model interface list --connector kmerp --filter-model-type query

# 查询领域模型
lingtong-cli model domain get --connector kmerp --business order

# 查询动态模型视图
lingtong-cli model dynamic view --connector kmerp --auth-account-id 296 --model-name SalesOrder
```

## 命令清单

| 目标 | 命令 |
|------|------|
| 接口模型列表 | `lingtong-cli model interface list --connector <name> --filter-model-type <type>` |
| 领域模型 | `lingtong-cli model domain get --connector <name> --business <business>` |
| 动态模型视图 | `lingtong-cli model dynamic view --connector <name> --auth-account-id <id>` |

## 参数说明

| 参数 | 适用命令 | 说明 |
|------|----------|------|
| `--connector` | 全部 | 连接器标识，如 `kmerp` |
| `--filter-model-type` | `interface list` | 过滤模型类型，如 `query`、`all` |
| `--model-type` | `interface list` | 可选模型类型 |
| `--business` | `domain get` | 业务对象标识，如 `order` |
| `--auth-account-id` | `interface list`、`domain get`、`dynamic view` | 授权账号 ID，动态模型通常必填 |
| `--model-name` | `dynamic view` | 可选模型名 |
| `--business-object-name` | `dynamic view` | 可选业务对象名 |

## 使用流程

```bash
# 1. 找账号
lingtong-cli connector account list --connector kmerp --env test

# 2. 找方法/接口模型
lingtong-cli model interface list --connector kmerp --filter-model-type all --auth-account-id 296

# 3. 查看动态字段结构
lingtong-cli model dynamic view --connector kmerp --auth-account-id 296 --model-name SalesOrder

# 4. 再去调用连接器或编写映射
lingtong-cli connector schema --connector kmerp --method erp.trade.list.query --auth-account-id 296
```

## 注意事项

- 本技能只读，不创建或修改模型。
- 动态模型与授权账号有关，换账号后字段结构可能不同。
- 如果不知道 method 名称，优先使用 `connector methods`。
- 如果需要刷新或创建动态模型，而专用命令未覆盖，可使用 `lingtong-service` 查询对应 `/meta/*` 接口。

## 相关技能

| 技能 | 说明 |
|------|------|
| `lingtong-connector` | 方法列表、字段 Schema、连接器调用 |
| `lingtong-scene` | 场景字段映射 |
| `lingtong-workflow` | 工作流节点字段配置 |
| `lingtong-service` | 长尾模型元数据 API |

## 维护说明

- Agent 执行细则以 [SKILL.md](SKILL.md) 为准。
- 新增模型子命令时同步更新主 README 和本 README。
