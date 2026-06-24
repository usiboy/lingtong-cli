# 销售出库单商品明细同步

从快麦ERP抽取销售出库单，按商品明细维度展开后写入绫通表格。

## 场景概述

| 项目 | 说明 |
|------|------|
| 数据源 | 快麦ERP `erp.trade.outstock.simple.query` 接口 |
| 数据处理 | 按订单商品明细维度展开（1个订单 → N个SKU行） |
| 目标 | 绫通高性能表格（17个字段） |
| 同步方式 | 管道节点自动分页 + 游标增量同步 |

## 工作流结构

```
w_start → w_modePipe → w_script(transform) → w_script(write) → w_script(stats) → w_end
```

## 前置准备

### 1. 连接器信息

| 参数 | 值 | 获取方式 |
|------|-----|----------|
| connector | kmerp | 连接器名称 |
| interfaceModelId | 73 | `model/info/category/interface` API |
| domainModelId | 3 | `childMetaInfos[0].id`（Trade 领域模型） |
| authAccountId | 296 | 广州力人服饰账户 |
| env | test | 测试环境 |

### 2. 目标表格

- **表格 ID**: 1636
- **模式**: 高性能模式
- **主键**: `detailId`（子订单行唯一ID，格式：`lineId_k`）
- **字段**（17个）：

| 字段 key | 标题 | 类型 | 说明 |
|----------|------|------|------|
| detailId | 明细ID | text | 主键，行级唯一 |
| sid | 系统订单号 | text | |
| tid | 平台订单号 | text | |
| created | 下单时间 | text | |
| buyerNick | 买家昵称 | text | |
| receiverMobile | 收件人手机 | text | |
| sysTitle | 商品标题 | text | |
| sysOuterId | 商家编码 | text | |
| sysSkuPropertiesName | SKU规格 | text | |
| num | 数量 | number | |
| price | 单价 | number | |
| cost | 成本 | number | |
| payAmount | 实付金额 | number | |
| postFee | 邮费 | number | |
| discountFee | 优惠 | number | |
| destName | 订单归属 | text | |
| sellerFlag | 旗帜 | text | |
| syncTime | 同步时间 | text | |

## 节点配置详解

### 开始节点（w_start）

```json
{
  "id": "w_start_first",
  "type": "w_start",
  "data": {
    "title": "开始",
    "outputVariables": []
  }
}
```

**说明**：无输入参数，管道节点通过内部游标机制自动管理数据拉取位置。

### 管道节点（w_modePipe）

```json
{
  "id": "w_modePipe_51yu5",
  "type": "w_modePipe",
  "data": {
    "title": "查询销售出库单",
    "pids": ["w_start_first"],
    "connector": "kmerp",
    "interfaceModelId": 73,
    "domainModelId": 3,
    "authAccountId": 296,
    "env": "test",
    "queryParams": [
      {
        "fieldId": 2952,
        "key": "timeType",
        "value": "created",
        "dataType": "string"
      }
    ],
    "pipeConfig": {
      "config": {
        "modeType": "count",
        "props": [
          {"paramType": "startModified", "isAdapted": true, "paramName": "startTime"},
          {"paramType": "endModified", "isAdapted": true, "paramName": "endTime"},
          {"paramType": "pageNo", "isAdapted": true, "paramName": "pageNo"},
          {"paramType": "pageSize", "isAdapted": true, "paramName": "pageSize"}
        ]
      }
    },
    "outputVariables": [
      {
        "variable": "list",
        "variableAttr": {
          "dataType": "json",
          "value": "$.list",
          "refValueType": "self"
        }
      },
      {
        "variable": "total",
        "variableAttr": {
          "dataType": "number",
          "value": "$.total",
          "refValueType": "self"
        }
      }
    ]
  }
}
```

**关键配置说明**：

| 配置项 | 值 | 说明 |
|--------|-----|------|
| modeType | count | 按 total 计算最大页码 |
| queryParams | timeType=created | 仅配置非游标字段 |
| outputVariables | $.list, $.total | value 必须以 `$.` 开头 |
| 出边 sourceHandle | "true" | WHILE 类型必须 |

**工作原理**：
1. 首次进入 → 初始化游标（startTime/endTime/pageNo/pageSize）
2. 调用连接器拉取一页数据
3. 通过 JsonPath 提取 `$.list` 和 `$.total` 输出
4. 递增 pageNo，检查是否到达最后一页
5. 保存游标到数据库
6. 循环直到 pageNo >= maxPage

### 转换脚本（w_script）

```json
{
  "id": "w_script_transform",
  "type": "w_script",
  "data": {
    "title": "转换为商品明细",
    "pids": ["w_modePipe_51yu5"],
    "inputVariables": [
      {
        "variable": "queryResult",
        "variableAttr": {
          "dataType": "json",
          "value": "$w_modePipe_51yu5.list"
        }
      }
    ],
    "scriptConfig": {
      "language": "javascript",
      "script": "var list = context.get('queryResult') || [];\nvar rows = [];\nvar st = new Date().toISOString().replace('T', ' ').substring(0, 19);\nfunction fmtTime(v) { if (v == null || v === '') return ''; if (typeof v === 'number') { return new Date(v).toISOString().replace('T',' ').substring(0,19); } return '' + v; }\nfunction num(v) { var n = parseFloat(v); return isNaN(n) ? 0 : n; }\nfunction pushRow(o, od, d, detailId) {\n    rows.push({\n        detailId: '' + detailId,\n        sid: o.sid || '', tid: o.tid || '', created: fmtTime(o.created),\n        buyerNick: o.buyerNick || '', receiverMobile: o.receiverMobile || '',\n        sysTitle: d.sysTitle || '', sysOuterId: d.sysOuterId || '', sysSkuPropertiesName: d.sysSkuPropertiesName || '',\n        num: num(d.num), price: num(d.price), cost: num(d.cost),\n        payAmount: num(o.payAmount), postFee: num(o.postFee), discountFee: num(o.discountFee),\n        destName: o.destName || '', sellerFlag: o.sellerFlag || '', syncTime: st\n    });\n}\nfor (var i = 0; i < list.length; i++) {\n    var o = list[i];\n    var orders = o.orders || [];\n    for (var j = 0; j < orders.length; j++) {\n        var od = orders[j];\n        var lineId = od.id || od.oid;\n        var suits = od.suits || [];\n        if (suits.length > 0) { for (var k = 0; k < suits.length; k++) { pushRow(o, od, suits[k], lineId + '_' + k); } }\n        else { pushRow(o, od, od, lineId); }\n    }\n}\nreturn {detailRows: rows, orderCount: list.length, detailCount: rows.length, total: 0, pageSize: 0, pageNo: 0};"
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

**数据展开逻辑**：
```
订单 (list[])
  └── 子订单 (orders[])
        ├── 有SKU明细 (suits[]) → 每个SKU生成一行
        └── 无SKU明细 → 子订单本身生成一行
```

### 写入脚本（w_script）

```json
{
  "id": "w_script_write",
  "type": "w_script",
  "data": {
    "title": "写入绫通表格",
    "pids": ["w_script_transform"],
    "inputVariables": [
      {
        "variable": "transformResult",
        "variableAttr": {
          "dataType": "json",
          "value": "$w_script_transform.result"
        }
      }
    ],
    "scriptConfig": {
      "language": "javascript",
      "script": "var basicDataId = 1636;\nvar t = context.get('transformResult') || {};\nvar rows = t.detailRows || [];\nvar ok = 0, fail = 0, firstError = null;\nfor (var i = 0; i < rows.length; i++) {\n    try {\n        var r = InfInvoker.insert(context, basicDataId, rows[i]);\n        if (r != null && r > 0) { ok++; } else { fail++; if (firstError == null) { firstError = 'insert 返回 ' + r; } }\n    } catch (e) { fail++; if (firstError == null) { firstError = '' + e; } }\n}\nreturn {successCount: ok, failCount: fail, totalCount: rows.length, firstError: firstError, orderCount: t.orderCount, detailCount: t.detailCount, pageNo: t.pageNo, pageSize: t.pageSize, total: t.total};"
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

### 统计脚本（w_script）

```json
{
  "id": "w_script_stats",
  "type": "w_script",
  "data": {
    "title": "统计同步结果",
    "pids": ["w_script_write"],
    "inputVariables": [
      {
        "variable": "writeResult",
        "variableAttr": {
          "dataType": "json",
          "value": "$w_script_write.result"
        }
      }
    ],
    "scriptConfig": {
      "language": "javascript",
      "script": "var w = context.get('writeResult');\nvar s = {syncTime: new Date().toISOString().replace('T',' ').substring(0,19), success: true, message: '同步完成', statistics: {orderCount: w.orderCount, detailCount: w.detailCount, successCount: w.successCount, failCount: w.failCount, pageNo: w.pageNo, pageSize: w.pageSize, total: w.total}};\nif(w.failCount > 0){s.message = '同步完成，但有'+w.failCount+'条失败';s.success=false;}\nreturn s;"
    },
    "outputVariables": [
      {
        "variable": "statistics",
        "variableAttr": {"dataType": "json", "label": "同步统计"}
      }
    ]
  }
}
```

### 结束节点（w_end）

```json
{
  "id": "w_end_result",
  "type": "w_end",
  "data": {
    "title": "结束",
    "pids": ["w_script_stats"],
    "outputVariables": [
      {
        "variable": "result",
        "variableAttr": {
          "dataType": "json",
          "value": "$w_script_stats.statistics",
          "nodeId": "w_end_result"
        }
      }
    ]
  }
}
```

## 边连接

```json
[
  {"source": "w_start_first", "target": "w_modePipe_51yu5", "sourceHandle": "source"},
  {"source": "w_modePipe_51yu5", "target": "w_script_transform", "sourceHandle": "true"},
  {"source": "w_script_transform", "target": "w_script_write", "sourceHandle": "source"},
  {"source": "w_script_write", "target": "w_script_stats", "sourceHandle": "source"},
  {"source": "w_script_stats", "target": "w_end_result", "sourceHandle": "source"}
]
```

**注意**：管道节点出边 `sourceHandle` 必须为 `"true"`（WHILE 类型特殊要求）。

## 执行与验证

### 发布

```bash
lingtong-cli workflow publish --workflow-id 1017 --version "v2.0.3" --memo "管道节点+数据加工+表格写入"
```

### 测试

```bash
lingtong-cli workflow api-test --app-tag <tag> --params '{}'
```

### 查看游标

```bash
lingtong-cli service workflow design-getcursor --workflowId 1017 --nodeId w_modePipe_51yu5
```

### 查看执行日志

```bash
lingtong-cli service workflow log-querylist --flowId 1017 --appId 165 --pageSize 10 --pageNum 1
```

## 注意事项

1. **管道节点是 WHILE 类型**，每页迭代一次，下游节点逐页处理
2. **高性能表主键需行级唯一**（子订单ID），不可用订单ID
3. **InfInvoker.insert 使用字段语义 key**（sid/num），非数字 id
4. **游标首次执行自动初始化**，后续从上次位置继续增量同步
5. **输出变量 value 必须以 `$.` 开头**，否则会被解析为 context 类型
6. **WHILE 节点出边 sourceHandle 必须为 `"true"`**，否则流程引擎无法解析
