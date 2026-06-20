# 简单工作流示例

最基础的工作流：开始 → 结束。

## DSL

```json
{
  "environment": "formal",
  "name": "简单工作流",
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
            "variable": "message",
            "variableAttr": {
              "dataType": "text",
              "label": "消息",
              "required": true
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
      "position": {"x": 416, "y": 100},
      "positionAbsolute": {"x": 416, "y": 100},
      "selected": false,
      "data": {
        "title": "结束",
        "desc": "",
        "disabled": false,
        "pids": ["w_start_first"],
        "outputVariables": [
          {
            "variable": "result",
            "variableAttr": {
              "dataType": "text",
              "value": "$w_start_first.message",
              "nodeId": "w_end_result"
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
      "target": "w_end_result",
      "sourceHandle": "source",
      "targetHandle": "target",
      "type": "rounded-corner"
    }
  ]
}
```

## 创建命令

```bash
# 使用模板创建
lingtong-cli workflow create --name "简单工作流" --template simple

# 使用 DSL 文件创建
lingtong-cli workflow create --name "简单工作流" --dsl-file simple.json
```

## 测试

```bash
# 执行工作流
lingtong-cli workflow execute --workflow-id <id> --params '{"message":"Hello World"}'
```
