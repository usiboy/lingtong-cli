# 数据管道示例

批量数据同步：开始 → 数据管道 → 数据拆分 → 脚本处理 → 结束。

## 场景

从快麦 ERP 同步订单数据，逐条处理后输出。

## 工作流结构

```
w_start → w_modePipe → w_dataSplit → w_script → w_end
```

## 关键节点配置

### 数据管道节点

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

### 数据拆分节点

```json
{
  "id": "w_dataSplit_item",
  "type": "w_dataSplit",
  "data": {
    "title": "拆分订单",
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

### 脚本处理节点

```json
{
  "id": "w_script_process",
  "type": "w_script",
  "data": {
    "title": "处理订单",
    "pids": ["w_dataSplit_item"],
    "inputVariables": [
      {
        "variable": "order",
        "variableAttr": {
          "value": "$w_dataSplit_item.item",
          "dataType": "json"
        }
      }
    ],
    "scriptConfig": {
      "script": "var order = context.get('order');\nreturn {\n  orderId: order.id,\n  status: order.status,\n  amount: order.totalAmount\n};"
    },
    "outputVariables": [
      {
        "variable": "result",
        "variableAttr": {"dataType": "json"}
      }
    ]
  }
}
```

## 完整流程

```bash
# 1. 创建工作流
lingtong-cli workflow create --name "订单同步" --template data_pipeline

# 2. 分析依赖
lingtong-cli workflow dependency list --workflow-id <id>

# 3. 测试执行
lingtong-cli workflow test run --workflow-id <id> --params '{"startDate":"2026-04-01"}'

# 4. 发布
lingtong-cli workflow publish --workflow-id <id> --version "v1.0.0"
```
