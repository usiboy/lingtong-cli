# 连接器 API 化示例

将连接器接口包装为独立 API：开始 → 连接器 → 结束。

## 场景

将快麦 ERP 的仓库查询接口包装为独立 API，支持按编码和名称查询。

## DSL

```json
{
  "environment": "formal",
  "name": "仓库查询API",
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
            "variableAttr": {"dataType": "text", "label": "仓库编码"}
          },
          {
            "variable": "name",
            "variableAttr": {"dataType": "text", "label": "仓库名称"}
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
        "domainModelId": 3,
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

## 完整流程

```bash
# 1. 创建工作流
lingtong-cli workflow create --name "仓库查询API" --dsl-file warehouse-api.json

# 2. 验证 DSL
lingtong-cli workflow validate --workflow-id <new_id>

# 3. 启用 Open API
lingtong-cli workflow api-enable --workflow-id <new_id>

# 4. 发布
lingtong-cli workflow publish --workflow-id <new_id> --version "v1.0.0"

# 5. 测试 API
lingtong-cli workflow api-test --app-tag <appTag> --params '{"code":"A","name":"广州"}'
```

## 参数映射

连接器节点的请求参数通过前端界面的"参数映射设置"配置：

- `code` → `$w_start_first.code`
- `name` → `$w_start_first.name`

DSL 中不需要包含 `fieldMapping` 字段。
