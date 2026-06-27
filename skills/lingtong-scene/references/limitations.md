# CLI `scene create` 命令限制说明

## 概述

`lingtong-cli scene create` 命令是一个**简化版**的场景创建工具，仅支持设置场景名称和描述。它创建的场景处于**草稿状态**，需要后续配置才能使用。

## 命令能力

### 支持的操作

```bash
lingtong-cli scene create --name <name> [--description <desc>]
```

| 参数 | 说明 |
|------|------|
| `--name` | 场景名称（必填） |
| `--description` | 场景描述（可选） |

### 不支持的配置

以下配置**无法**通过 CLI 命令设置：

| 配置项 | 说明 |
|--------|------|
| `appId` | 集成应用 ID |
| `env` | 环境（test/prod） |
| `type` | 场景类型（1-4） |
| `sourceConnector` | 源连接器 |
| `targetConnector` | 目标连接器 |
| `sourceInterfaceModelId` | 源接口模型 |
| `targetInterfaceModelId` | 目标接口模型 |
| `ltSceneAccountDtoList` | 账号配置 |
| `syncModel` | 同步模式 |

## 完整场景创建流程

完整的场景创建需要 6 步 API 调用：

```
1. 选择集成应用 → 获取 appId
2. 选择源连接器 → 调用 /connector/list
3. 选择源账号 → 调用 /account/connector/list
4. 选择源接口 → 调用 /model/info/category/interface (interfaceModelType=query)
5. 选择目标连接器/账号/接口 → 同步骤 2-4
6. 创建场景 → 调用 /scene/model/save（包含所有配置）
7. 配置数据视图、预处理和字段映射 → 使用 `scene field-mapping` 或 `service scene ...` 长尾接口
```

## 推荐方案

对于需要完整配置的场景，不要只调用 `scene create`。推荐先用 `lingtong-connector` 和 `lingtong-model` 查询连接器、账号和模型，再用 `scene update`、`scene trigger`、`scene field-mapping` 或 `lingtong-cli service scene ...` 补齐长尾配置。

## 使用场景对比

| 场景 | 推荐方式 |
|------|---------|
| 快速创建占位场景 | `lingtong-cli scene create` |
| 创建可运行的完整场景 | 组合 `connector`、`model`、`scene update`、`workflow` 或 `service scene ...` |
| 批量创建场景 | `lingtong-cli service scene ...` 或应用导入流程 |
| 创建带触发器的场景 | `scene trigger save` 或 `service scene ...` |

## 相关文档

- [scene-types.md](./scene-types.md) - 场景类型详解
- [api-response-fields.md](./api-response-fields.md) - API 响应字段说明
