---
name: lingtong-scene
version: 2.0.0
description: "绫通场景管理：查询、列出、创建集成场景。当用户需要管理集成场景、查询场景详情、快速创建草稿场景时触发。关键词：scene、场景、scene list、scene create、scene info、integration。"
metadata:
  requires:
    bins: ["lingtong-cli"]
  cliHelp: "lingtong-cli scene --help"
---

# lingtong-scene

## 何时使用

使用本 skill：

- 用户要查看、列出集成场景
- 用户要查询场景详情
- 用户要快速创建一个草稿场景
- 用户要管理本地应用文件中的场景配置

不要使用本 skill：

- 完整的场景创建流程（连接器选择、账号配置、模型选择）不能只靠 `scene create`，需要组合 `lingtong-connector`、`lingtong-model`、`lingtong-workflow` 和 `lingtong-service`
- 数据视图、数据预处理、复杂字段映射配置优先使用现有专用命令或 `lingtong-service scene ...` 长尾 API
- 只是查询连接器信息或账户，转 `lingtong-connector`
- 只是管理工作流，转 `lingtong-workflow`

## 使用边界

- 场景查询使用 `lingtong-cli scene` 命令
- 复杂场景创建（需要选择连接器、账号、模型）需要组合连接器、模型、场景和工作流命令
- 本地应用文件中的场景管理使用 `lingtong-cli app scene` 命令
- API 响应数据以 `result` 字段包裹，不是 `data`
- `scene create` 命令仅创建草稿场景，完整配置需继续使用 `scene update`、`scene trigger`、`scene field-mapping` 或 `service scene ...`

## 快速路由

| 用户目标 | 优先命令 | 何时读 reference |
|---|---|---|
| 查看场景列表 | `scene list [--page <n>] [--page-size <n>] [--app-id <id>]` | - |
| 查看场景详情 | `scene info --scene-id <id>` | - |
| 快速创建草稿场景 | `scene create --name <name> [--description <desc>]` | 读 [limitations.md](references/limitations.md) 了解 CLI 创建的限制 |
| 完整场景创建流程 | 组合 `connector`、`model`、`scene update`、`workflow` 或 `service scene ...` | 读 [limitations.md](references/limitations.md) |
| 管理本地应用场景 | `app scene list/add/remove --file <file>` | - |
| 了解场景类型 | - | 读 [scene-types.md](references/scene-types.md) |
| 了解 API 响应结构 | - | 读 [api-response-fields.md](references/api-response-fields.md) |

## 场景心智模型

- 场景是绫通平台的**数据同步单元**，每个场景定义了一对连接器之间的数据流
- 场景由**源连接器**（source）和**目标连接器**（target）组成，通过字段映射实现数据转换
- 场景类型：1=正常场景，2=触发器场景，3=消息回调场景，4=推送数据场景
- 同步模式：1=双流模式（双向），2=直推模式（单向）
- 场景是工作流的容器，一个场景可包含多个工作流
- CLI `scene create` 仅创建草稿场景（只设置 name/description），完整配置需继续补齐连接器、账号、模型、字段映射和工作流

## 命令路由

### 平台场景管理

```bash
# 列出场景（支持分页和应用过滤）
lingtong-cli scene list [--page 1] [--page-size 20] [--app-id <id>]

# 查询场景详情
lingtong-cli scene info --scene-id <id>

# 快速创建草稿场景（仅 name/description，创建后为草稿状态）
lingtong-cli scene create --name <name> [--description <desc>]
```

**示例**：
```bash
# 列出所有场景
lingtong-cli scene list

# 分页查询
lingtong-cli scene list --page 1 --page-size 50

# 按应用过滤
lingtong-cli scene list --app-id 165 --page 1 --page-size 50

# 查询场景详情
lingtong-cli scene info --scene-id 1001

# 创建草稿场景
lingtong-cli scene create --name "订单同步" --description "从金蝶云同步订单到自有ERP"
```

### 本地应用场景管理

```bash
# 列出应用文件中的场景
lingtong-cli app scene list --file app.json

# 添加场景到应用文件
lingtong-cli app scene add --file app.json --name <name> --source <connector> --target <connector> [--output <file>]

# 从应用文件移除场景
lingtong-cli app scene remove --file app.json --name <name> [--output <file>]
```

**示例**：
```bash
# 列出应用中的所有场景
lingtong-cli app scene list --file app.json

# 添加快麦到金蝶的场景
lingtong-cli app scene add --file app.json --name "商品同步" --source kmerp --target kingDeeCloudStar

# 移除场景并输出到新文件
lingtong-cli app scene remove --file app.json --name "商品同步" --output updated.json
```

### 快捷命令

```bash
# 快捷列出场景（支持透传所有参数）
lingtong-cli +scene-list [--page <n>] [--page-size <n>] [--app-id <id>]
```

## 常见恢复

| 错误 / 现象 | 恢复动作 |
|---|---|
| `no host configured` | 运行 `lingtong-cli config init --host <url>` |
| 场景创建返回但无 sceneId | 兼容提取 `result.sceneId ?? result.id` |
| 场景详情查询失败 | 确认使用 `--scene-id` 参数（不是 `--id`） |
| API 返回权限错误 | 检查 Token 是否有效：`lingtong-cli auth status` |
| `scene create` 后场景不完整 | 这是预期行为，继续使用 `scene update`、`scene trigger`、`scene field-mapping` 或 `service scene ...` 补齐配置 |
| 本地文件场景解析失败 | 检查 JSON 格式是否正确 |

## 数据模型

### 场景列表响应 (`/scene/list`)

```json
{
  "success": true,
  "code": 10000,
  "result": {
    "data": [
      {
        "id": 1001,
        "name": "kmerp商品同步",
        "type": 1,
        "status": 1,
        "open": 1,
        "env": "test",
        "sourceConnector": "kmerp",
        "targetConnector": "chanjet"
      }
    ],
    "total": 1
  }
}
```

### 场景详情响应 (`/scene/detail/get`)

```json
{
  "success": true,
  "code": 10000,
  "result": {
    "id": 1001,
    "tenantId": "1828697714616041472",
    "appId": 15,
    "sceneNumber": "S1001",
    "name": "kmerp商品同步",
    "type": 1,
    "status": 1,
    "open": 1,
    "env": "test",
    "sourceConnector": "kmerp",
    "sourceCatId": "商品",
    "sourceDomainModelId": 81,
    "sourceInterfaceModelId": 988,
    "targetConnector": "chanjet",
    "targetCatId": "T+基础档案-存货",
    "targetDomainModelId": 1355,
    "targetInterfaceModelId": 1387,
    "syncModel": 1,
    "persistent": 1,
    "ltSceneAccountDtoList": [
      {"accountType": "source", "authAccountId": 30},
      {"accountType": "target", "authAccountId": 305}
    ]
  }
}
```

### 场景创建响应 (`/scene/model/save`)

```json
{
  "success": true,
  "code": 10000,
  "msg": "操作成功",
  "result": {
    "sceneId": 1001
  },
  "clueId": "719128196028928301"
}
```

## 最佳实践

1. **先列出场景**，了解现有场景结构和 ID
2. **使用 `scene info`** 查看完整配置再修改
3. **快速创建用 `scene create`**，但要知道它只创建草稿
4. **完整配置需要组合多个命令**，先查连接器和模型，再补齐场景配置、字段映射和工作流
5. **本地开发用 `app scene`** 命令管理应用 JSON 文件
6. **场景 ID 提取**：兼容 `result.sceneId` 和 `result.id`（优先 `sceneId`）

## 生命周期命令(增强)

除 list/create/info 外,场景命令组提供完整生命周期操作:

```bash
lingtong-cli scene update --data '{"id":123,"name":"renamed"}'   # 更新
lingtong-cli scene copy --scene-id 123                            # 复制
lingtong-cli scene open --scene-id 123                            # 开启/关闭(切换)
lingtong-cli scene publish --scene-id 123                         # 发布版本
lingtong-cli scene publish --data '{"sceneId":123,"remark":"v2"}' # 发布(完整快照)
lingtong-cli scene version list --scene-id 123                    # 版本历史
lingtong-cli scene trigger get  --scene-id 123                    # 触发器条件
lingtong-cli scene trigger save --data '{"sceneId":123}'          # 保存触发器条件
lingtong-cli scene field-mapping list    --scene-id 123           # 字段映射列表
lingtong-cli scene field-mapping execute --data '{"sceneId":123}' # 字段映射调试
lingtong-cli scene delete --scene-id 123 --yes                    # 删除(破坏性,需 --yes)
```

**规则**:`delete` 默认拒绝,必须加 `--yes`,否则以**退出码 10**(需要确认)失败。更长尾的场景接口(数据合并/预处理/暂存模型等)用自动生成层 [[lingtong-service]]:`lingtong-cli service scene ...`。

## 保留 Reference

- [limitations.md](references/limitations.md) - CLI `scene create` 命令的限制说明
- [scene-types.md](references/scene-types.md) - 场景类型和同步模式详解
- [api-response-fields.md](references/api-response-fields.md) - 完整 API 响应字段说明
