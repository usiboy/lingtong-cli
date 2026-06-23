# 工作流节点详细配置

本文档详细说明绫通工作流 12 种节点类型的配置方法。

## 目录

- [w_start - 开始节点](#w_start)
- [w_end - 结束节点](#w_end)
- [w_connector - 连接器节点](#w_connector)
- [w_script - 脚本节点](#w_script)
- [w_if - 条件分支节点](#w_if)
- [w_switch - 选择分支节点](#w_switch)
- [w_cycle - 循环节点](#w_cycle)
- [w_modePipe - 数据管道节点](#w_modepipe)
- [w_dataSplit - 数据拆分节点](#w_datasplit)
- [w_joinPipeline - 聚合节点](#w_joinpipeline)
- [w_dataPush - 数据推送节点](#w_datapush)
- [w_pushRecord - 推送记录节点](#w_pushrecord)

---

## w_start

开始节点定义工作流的输入参数。

### 配置选项

```json
{
  "id": "w_start_first",
  "type": "w_start",
  "data": {
    "title": "开始",
    "outputVariables": [
      {
        "variable": "orderId",
        "variableAttr": {
          "dataType": "text",
          "label": "订单编号",
          "required": true
        }
      },
      {
        "variable": "amount",
        "variableAttr": {
          "dataType": "number",
          "label": "订单金额",
          "required": false,
          "defaultValue": 0
        }
      }
    ]
  }
}
```

### 规则

- 每个工作流必须有且仅有一个 `w_start` 节点
- `outputVariables` 定义工作流的输入参数
- 参数通过 Open API 调用时传入

### 数据类型

| dataType | 说明 | 示例 |
|----------|------|------|
| `text` | 字符串 | `"hello"` |
| `number` | 数字 | `123`, `45.6` |
| `json` | JSON 对象 | `{"key": "value"}` |
| `boolean` | 布尔值 | `true`, `false` |

---

## w_end

结束节点定义工作流的输出结果。

### 配置选项

```json
{
  "id": "w_end_result",
  "type": "w_end",
  "data": {
    "title": "结束",
    "pids": ["w_script_invoke"],
    "outputVariables": [
      {
        "variable": "result",
        "variableAttr": {
          "dataType": "json",
          "value": "$w_script_invoke.result",
          "nodeId": "w_end_result",
          "schemaObj": {
            "fullPath": "",
            "schemaNodeVariable": "result",
            "schemaNodeId": "w_script_invoke"
          }
        }
      }
    ]
  }
}
```

### 规则

- 至少有一个 `w_end` 节点
- `pids` 指定上游节点 ID
- `outputVariables` 从上游节点提取输出

---

## w_connector

连接器节点调用外部系统 API。

### 配置选项

```json
{
  "id": "w_connector_sales",
  "type": "w_connector",
  "data": {
    "title": "销售出库单查询",
    "pids": ["w_start_first"],
    "connector": "kmerp",
    "interfaceModelId": 73,
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
          "nodeId": "w_connector_sales"
        }
      }
    ]
  }
}
```

### 必需配置

| 字段 | 说明 |
|------|------|
| `connector` | 连接器类型（如 `kmerp`, `feishu`） |
| `interfaceModelId` | 接口模型 ID |
| `authAccountId` | 认证账户 ID |
| `env` | 环境（`test` / `prod`） |

### 参数映射

连接器节点的请求参数通过前端界面的"参数映射设置"配置，DSL 中不需要包含 `fieldMapping`。

### 分支

- 成功路径：`true`
- 失败路径：`false`（配合 `assertConfig` 使用）

---

## w_script

脚本节点执行 JavaScript 代码。

### 配置选项

```json
{
  "id": "w_script_transform",
  "type": "w_script",
  "data": {
    "title": "数据转换",
    "pids": ["w_connector_query"],
    "inputVariables": [
      {
        "variable": "rawData",
        "variableAttr": {
          "value": "$w_connector_query.response.list",
          "dataType": "json"
        }
      }
    ],
    "scriptConfig": {
      "language": "javascript",
      "script": "var items = context.get('rawData');\nvar result = [];\nfor (var i = 0; i < items.length; i++) {\n  result.push({id: items[i].id, name: items[i].name});\n}\nreturn result;"
    },
    "outputVariables": [
      {
        "variable": "result",
        "variableAttr": {"dataType": "json", "schema": ""}
      }
    ],
    "assertConfig": {"assertType": "throwException"}
  }
}
```

### 重要规则

- **assertConfig（必填）**: `data` 必须含 `assertConfig`（如 `{"assertType":"throwException"}`）；
  缺失时保存/发布不报错，但执行报 `断言配置不允许为null`
- **语法限制**: ES5.1（Nashorn 引擎），不支持 ES6+
- **inputVariables**: 必须声明，引用上游节点输出
- **context API**: `context.get("key")` 获取输入变量
- **返回值**: 使用 `return` 返回结果

详细脚本编写指南见 [script-guide.md](script-guide.md)

---

## w_if

条件分支节点实现二元判断。

### 配置选项

```json
{
  "id": "w_if_check",
  "type": "w_if",
  "data": {
    "title": "金额判断",
    "pids": ["w_start_first"],
    "ifConfig": {
      "conditions": [
        {
          "leftValue": "$w_start_first.amount",
          "operator": "gte",
          "rightValue": "1000",
          "dataType": "number"
        }
      ],
      "logicalOperator": "and"
    }
  }
}
```

### 操作符

| 操作符 | 说明 |
|--------|------|
| `eq` | 等于 |
| `ne` | 不等于 |
| `gt` | 大于 |
| `gte` | 大于等于 |
| `lt` | 小于 |
| `lte` | 小于等于 |
| `contains` | 包含 |

### 分支

- `true`: 条件成立
- `false`: 条件不成立

---

## w_switch

多路分支节点实现多条件匹配。

### 配置选项

```json
{
  "id": "w_switch_status",
  "type": "w_switch",
  "data": {
    "title": "状态路由",
    "pids": ["w_start_first"],
    "switchConfig": {
      "express": "$w_start_first.status",
      "caseList": [
        {"caseExpr": "pending", "outLinkKey": "handle_pending"},
        {"caseExpr": "confirmed", "outLinkKey": "handle_confirmed"},
        {"caseExpr": "cancelled", "outLinkKey": "handle_cancelled"}
      ]
    }
  }
}
```

---

## w_cycle

循环节点重复执行子流程。

### 配置选项

```json
{
  "id": "w_cycle_loop",
  "type": "w_cycle",
  "data": {
    "title": "遍历处理",
    "pids": ["w_dataSplit_item"],
    "cycleConfig": {
      "beginIndex": 0,
      "endCondition": "${index} < ${list.length}",
      "loopVariable": "${item}"
    }
  }
}
```

---

## w_modePipe

数据管道节点实现批量数据同步。

### 配置选项

```json
{
  "id": "w_modePipe_sync",
  "type": "w_modePipe",
  "data": {
    "title": "订单同步",
    "pids": ["w_start_first"],
    "connector": "kmerp",
    "interfaceModelId": 73,
    "domainModelId": 3,
    "authAccountId": 296,
    "env": "test",
    "pipeConfig": {
      "syncMode": "cursor",
      "cursorField": "modifiedTime"
    },
    "outputVariables": [
      {
        "variable": "orders",
        "variableAttr": {"dataType": "json"}
      }
    ]
  }
}
```

### 同步模式

| 模式 | 说明 |
|------|------|
| `count` | 计数模式 |
| `time` | 时间模式 |
| `cursor` | 游标模式 |

---

## w_dataSplit

数据拆分节点将数组遍历处理。

### 配置选项

```json
{
  "id": "w_dataSplit_item",
  "type": "w_dataSplit",
  "data": {
    "title": "拆分订单项",
    "pids": ["w_modePipe_sync"],
    "outputVariables": [
      {
        "variable": "item",
        "variableAttr": {
          "dataType": "json",
          "value": "$w_modePipe_sync.orders[*]"
        }
      }
    ]
  }
}
```

---

## w_joinPipeline

聚合节点合并多个分支结果。

### 配置选项

```json
{
  "id": "w_joinPipeline_merge",
  "type": "w_joinPipeline",
  "data": {
    "title": "合并结果",
    "pids": ["w_branch_a", "w_branch_b"],
    "outputVariables": [
      {
        "variable": "combined",
        "variableAttr": {"dataType": "json"}
      }
    ]
  }
}
```

---

## w_dataPush

数据推送节点带重试机制。

### 配置选项

```json
{
  "id": "w_dataPush_send",
  "type": "w_dataPush",
  "data": {
    "title": "推送到目标系统",
    "pids": ["w_script_transform"],
    "connector": "kingdee",
    "interfaceModelId": 456,
    "domainModelId": 789,
    "authAccountId": 123,
    "env": "prod",
    "retryConfig": {
      "maxRetry": 3,
      "retryInterval": 60
    }
  }
}
```

---

## w_pushRecord

推送记录节点查询推送历史。

### 配置选项

```json
{
  "id": "w_pushRecord_query",
  "type": "w_pushRecord",
  "data": {
    "title": "查询推送记录",
    "pids": ["w_dataPush_send"],
    "outputVariables": [
      {
        "variable": "records",
        "variableAttr": {"dataType": "json"}
      }
    ]
  }
}
```

---

## 节点通用配置

### pids

`pids` 数组指定上游节点 ID，用于建立节点间的依赖关系。

### assertConfig

错误处理配置：

```json
{
  "assertConfig": {
    "assertType": "throwException"  // 抛出异常
  }
}
```

### outputVariables

输出变量定义：

```json
{
  "variable": "result",
  "variableAttr": {
    "dataType": "json",
    "value": "$upstreamNode.output",
    "nodeId": "currentNodeId"
  }
}
```
