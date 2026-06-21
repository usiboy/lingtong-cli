# API 响应字段完整说明

## 概述

本文档详细说明绫通场景相关 API 的响应字段结构。所有 API 响应通过 `/gw/ai/proxy` 代理接口返回。

## 通用响应结构

```json
{
  "success": true,
  "code": 10000,
  "msg": "操作成功",
  "result": {...},
  "clueId": "719128196028928301"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| success | Boolean | 请求是否成功 |
| code | Number | 响应码，10000 表示成功 |
| msg | String | 响应消息 |
| result | Object/Array | 响应数据主体 |
| clueId | String | 请求追踪 ID |

**注意**：响应数据包裹在 `result` 字段中，不是 `data`。

---

## 场景列表 (`/scene/list`)

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pageNum | Number | 是 | 页码，从 1 开始 |
| pageSize | Number | 是 | 每页数量 |
| appId | Number | 否 | 应用 ID 过滤 |

### 响应结构

```json
{
  "success": true,
  "result": {
    "data": [
      {
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
        "targetConnector": "chanjet",
        "syncModel": 1,
        "persistent": 1,
        "createTime": "2026-01-15 10:30:00",
        "updateTime": "2026-01-15 10:30:00"
      }
    ],
    "total": 1
  }
}
```

### 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| id | Number | 场景 ID |
| tenantId | String | 租户 ID |
| appId | Number | 应用 ID |
| sceneNumber | String | 场景编号 |
| name | String | 场景名称 |
| type | Number | 场景类型：1=正常，2=触发器，3=消息回调，4=推送数据 |
| status | Number | 状态：-1=禁用，0=草稿，1=已发布 |
| open | Number | 开放状态：0=关闭，1=开放 |
| env | String | 环境：test/formal |
| sourceConnector | String | 源连接器标识 |
| targetConnector | String | 目标连接器标识 |
| syncModel | Number | 同步模式：1=双流，2=直推 |
| persistent | Number | 是否持久化：0=否，1=是 |
| createTime | String | 创建时间 |
| updateTime | String | 更新时间 |

---

## 场景详情 (`/scene/detail/get`)

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sceneId | Number | 是 | 场景 ID |

### 响应结构

```json
{
  "success": true,
  "result": {
    "id": 1001,
    "tenantId": "1828697714616041472",
    "appId": 15,
    "sceneNumber": "S1001",
    "name": "kmerp商品同步",
    "description": "从快麦ERP同步商品到畅捷通T+",
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
    "isAssoPullDataScene": 0,
    "joinMasterSceneId": null,
    "sourceSceneId": null,
    "dynamicInterfaceOnlyCreate": 0,
    "switchOutMasterSceneId": 0,
    "ltSceneAccountDtoList": [
      {
        "accountType": "source",
        "authAccountId": 30
      },
      {
        "accountType": "target",
        "authAccountId": 305
      }
    ],
    "createTime": "2026-01-15 10:30:00",
    "updateTime": "2026-01-15 10:30:00"
  }
}
```

### 额外字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| description | String | 场景描述 |
| sourceCatId | String | 源类目 ID |
| sourceDomainModelId | Number | 源领域模型 ID |
| sourceInterfaceModelId | Number | 源接口模型 ID |
| targetCatId | String | 目标类目 ID |
| targetDomainModelId | Number | 目标领域模型 ID |
| targetInterfaceModelId | Number | 目标接口模型 ID |
| isAssoPullDataScene | Number | 是否关联拉取场景：0=否，1=是 |
| joinMasterSceneId | Number/null | 关联主场景 ID |
| sourceSceneId | Number/null | 源场景 ID |
| dynamicInterfaceOnlyCreate | Number | 动态接口模式：0=只插入，1=插入或更新，2=只更新 |
| switchOutMasterSceneId | Number | 切换掉的主场景 ID |
| ltSceneAccountDtoList | Array | 账号配置列表 |

### 账号配置字段

| 字段 | 类型 | 说明 |
|------|------|------|
| accountType | String | 账号类型：source（源）/ target（目标） |
| authAccountId | Number | 认证账号 ID |

---

## 场景创建 (`/scene/model/save`)

### 请求体

```json
{
  "appId": 15,
  "name": "kmerp商品同步到畅捷通",
  "description": "从快麦ERP同步商品到畅捷通T+",
  "env": "test",
  "type": 1,
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
  "isAssoPullDataScene": 0,
  "dynamicInterfaceOnlyCreate": 0,
  "switchOutMasterSceneId": 0,
  "ltSceneAccountDtoList": [
    {"accountType": "source", "authAccountId": 30},
    {"accountType": "target", "authAccountId": 305}
  ]
}
```

### 响应结构

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

### 响应字段

| 字段 | 类型 | 说明 |
|------|------|------|
| result.sceneId | Number | 创建的场景 ID |

**注意**：部分接口可能返回 `result.id` 而非 `result.sceneId`，建议兼容处理：
```javascript
const sceneId = result.sceneId ?? result.id;
```

---

## 错误响应

### 常见错误码

| 错误码 | 说明 |
|--------|------|
| 10000 | 成功 |
| 40001 | 参数错误 |
| 40003 | 无权限 |
| 40004 | 资源不存在 |
| 50000 | 服务器内部错误 |

### 错误响应示例

```json
{
  "success": false,
  "code": 40004,
  "msg": "场景不存在",
  "result": null
}
```

---

## 相关文档

- [limitations.md](./limitations.md) - CLI 命令限制说明
- [scene-types.md](./scene-types.md) - 场景类型详解
