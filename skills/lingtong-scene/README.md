# lingtong-scene 技能

`lingtong-scene` 用于平台集成场景管理，覆盖场景查询、草稿创建、生命周期操作、触发条件和字段映射调试。应用 JSON 文件内的场景维护由 `lingtong-cli app scene` 提供，属于 `lingtong-cli-app` 的范围。

## 何时使用

- 查询平台场景列表或详情
- 快速创建只有名称和描述的草稿场景
- 复制、更新、开关、发布或删除场景
- 查看场景版本历史
- 查询或保存触发条件
- 调试场景字段映射

## 快速开始

```bash
lingtong-cli scene list --page 1 --page-size 20 --app-id 165
lingtong-cli scene info --scene-id 123
lingtong-cli scene create --name 订单同步 --description 同步订单
```

## 命令清单

| 目标 | 命令 |
|------|------|
| 列出场景 | `lingtong-cli scene list --page 1 --page-size 20 --app-id <id>` |
| 查询详情 | `lingtong-cli scene info --scene-id <id>` |
| 创建草稿 | `lingtong-cli scene create --name <name> --description <desc>` |
| 更新场景 | `lingtong-cli scene update --data '<json>'` |
| 复制场景 | `lingtong-cli scene copy --scene-id <id>` |
| 开关场景 | `lingtong-cli scene open --scene-id <id>` |
| 发布场景 | `lingtong-cli scene publish --scene-id <id>` |
| 版本列表 | `lingtong-cli scene version list --scene-id <id>` |
| 查询触发条件 | `lingtong-cli scene trigger get --scene-id <id>` |
| 保存触发条件 | `lingtong-cli scene trigger save --data '<json>'` |
| 字段映射列表 | `lingtong-cli scene field-mapping list --scene-id <id>` |
| 字段映射调试 | `lingtong-cli scene field-mapping execute --data '<json>'` |
| 删除场景 | `lingtong-cli scene delete --scene-id <id> --yes` |

## 边界说明

- `scene create` 只创建最小草稿场景，不会完成连接器账号、模型选择、字段映射等完整配置。
- 应用配置文件内的场景增删查使用 `lingtong-cli app scene ...`。
- 数据预处理、复杂字段映射脚本可结合 `lingtong-script`。
- 长尾场景 API 使用 `lingtong-cli service scene ...`。

## 常见流程

```bash
# 查看详情后复制并发布
lingtong-cli scene info --scene-id 123
lingtong-cli scene copy --scene-id 123
lingtong-cli scene publish --scene-id 124

# 调试字段映射
lingtong-cli scene field-mapping list --scene-id 123
lingtong-cli scene field-mapping execute --data '{"sceneId":123,"record":{}}'
```

## Reference

| 文档 | 说明 |
|------|------|
| [references/limitations.md](references/limitations.md) | `scene create` 的限制 |
| [references/scene-types.md](references/scene-types.md) | 场景类型和同步模式 |
| [references/api-response-fields.md](references/api-response-fields.md) | API 响应字段说明 |

## 相关技能

| 技能 | 说明 |
|------|------|
| `lingtong-cli-app` | 应用文件中的场景配置 |
| `lingtong-connector` | 连接器和授权账户 |
| `lingtong-model` | 模型元数据 |
| `lingtong-workflow` | 场景内工作流编排 |
| `lingtong-script` | 字段映射脚本 |

## 维护说明

- Agent 执行细则以 [SKILL.md](SKILL.md) 为准。
- 新增场景子命令时同步更新主 README、本 README 和 reference。
