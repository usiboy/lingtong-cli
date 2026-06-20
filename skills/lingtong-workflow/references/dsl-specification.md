# DSL 规范

本文档定义绫通工作流 DSL 的完整规范。

## 顶层结构

```json
{
  "environment": "formal",
  "name": "工作流名称",
  "viewport": {"x": 52, "y": 11, "zoom": 1},
  "workId": 1002,
  "nodes": [...],
  "edges": [...]
}
```

| 字段 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `environment` | string | 是 | 运行环境：`formal` |
| `name` | string | 是 | 工作流名称 |
| `viewport` | object | 是 | 画布视口配置 |
| `workId` | number | 更新时 | 工作流 ID |
| `nodes` | array | 是 | 节点列表 |
| `edges` | array | 是 | 边列表 |

## 节点结构

```json
{
  "id": "w_start_first",
  "type": "w_start",
  "dragging": false,
  "width": 256,
  "height": 92,
  "position": {"x": 100, "y": 100},
  "positionAbsolute": {"x": 100, "y": 100},
  "selected": false,
  "data": {
    "title": "开始",
    "desc": "",
    "disabled": false,
    "outputVariables": [...]
  }
}
```

### 节点 UI 属性

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | string | 节点唯一标识 |
| `type` | string | 节点类型 |
| `dragging` | boolean | 是否拖拽中 |
| `width` | number | 节点宽度 |
| `height` | number | 节点高度 |
| `position` | object | 节点位置 |
| `positionAbsolute` | object | 绝对位置 |
| `selected` | boolean | 是否选中 |

### 节点 data 属性

| 字段 | 类型 | 说明 |
|------|------|------|
| `title` | string | 节点标题 |
| `desc` | string | 节点描述 |
| `disabled` | boolean | 是否禁用 |
| `pids` | array | 上游节点 ID |
| `outputVariables` | array | 输出变量定义 |

## 边结构

```json
{
  "id": "e1",
  "source": "w_start_first",
  "target": "w_script_invoke",
  "sourceHandle": "source",
  "targetHandle": "target",
  "type": "rounded-corner"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | string | 边唯一标识 |
| `source` | string | 源节点 ID |
| `target` | string | 目标节点 ID |
| `sourceHandle` | string | 源连接点：`source` |
| `targetHandle` | string | 目标连接点：`target` |
| `type` | string | 边类型：`rounded-corner` |

## 变量引用语法

```
$nodeId.variableName              # 基础引用
$w_start_first.orderId            # 从开始节点获取
$w_connector_1k65t.response.list  # 嵌套属性访问
$w_modePipe_725vx.orders[*]       # 数组遍历
```

## 完整 DSL 示例

```json
{
  "environment": "formal",
  "name": "仓库查询工作流",
  "viewport": {"x": 52, "y": 11, "zoom": 1},
  "nodes": [
    {
      "id": "w_start_first",
      "type": "w_start",
      "dragging": false,
      "width": 256,
      "height": 92,
      "position": {"x": 100, "y": 100},
      "positionAbsolute": {"x": 100, "y": 100},
      "selected": false,
      "data": {
        "title": "开始",
        "desc": "",
        "disabled": false,
        "outputVariables": [
          {
            "variable": "code",
            "variableAttr": {
              "dataType": "text",
              "label": "仓库编码"
            }
          }
        ]
      }
    },
    {
      "id": "w_connector_query",
      "type": "w_connector",
      "dragging": false,
      "width": 256,
      "height": 58,
      "position": {"x": 416, "y": 100},
      "positionAbsolute": {"x": 416, "y": 100},
      "selected": false,
      "data": {
        "title": "查询仓库",
        "desc": "",
        "disabled": false,
        "pids": ["w_start_first"],
        "connector": "kmerp",
        "interfaceModelId": 119,
        "authAccountId": 296,
        "env": "test",
        "assertConfig": {"assertType": "throwException"},
        "outputVariables": [
          {
            "variable": "response",
            "variableAttr": {
              "dataType": "json",
              "value": "$.",
              "nodeId": "w_connector_query"
            }
          }
        ]
      }
    },
    {
      "id": "w_end_result",
      "type": "w_end",
      "dragging": false,
      "width": 256,
      "height": 92,
      "position": {"x": 732, "y": 100},
      "positionAbsolute": {"x": 732, "y": 100},
      "selected": false,
      "data": {
        "title": "结束",
        "desc": "",
        "disabled": false,
        "pids": ["w_connector_query"],
        "outputVariables": [
          {
            "variable": "warehouses",
            "variableAttr": {
              "dataType": "json",
              "value": "$w_connector_query.response.list",
              "nodeId": "w_end_result",
              "schemaObj": {
                "fullPath": "",
                "schemaNodeVariable": "response",
                "schemaNodeId": "w_connector_query"
              }
            }
          }
        ]
      }
    }
  ],
  "edges": [
    {
      "id": "e1",
      "source": "w_start_first",
      "target": "w_connector_query",
      "sourceHandle": "source",
      "targetHandle": "target",
      "type": "rounded-corner"
    },
    {
      "id": "e2",
      "source": "w_connector_query",
      "target": "w_end_result",
      "sourceHandle": "source",
      "targetHandle": "target",
      "type": "rounded-corner"
    }
  ]
}
```

## 验证规则

- 必须包含 `nodes` 和 `edges` 字段
- 必须有且仅有一个 `w_start` 节点
- 至少有一个 `w_end` 节点
- 边的 `source`/`target` 必须引用存在的节点 ID
- `w_connector`/`w_modePipe` 节点必须配置 `connector` 字段
