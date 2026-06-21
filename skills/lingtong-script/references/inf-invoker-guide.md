# InfInvoker 表格函数完整指南

InfInvoker 是绫通脚本中的内置对象，用于在脚本中操作绫通表格数据（基础资料）。

## 概述

**类型**: `LtInfInvoker`（Java 接口）

**实现**: `DefaultLtInfInvoker`

**使用场景**: 工作流脚本节点、场景字段函数、连接器字段函数（三种场景均可用）

**作用**: 操作绫通表格数据，包括查询、插入、更新、删除和 ES 聚合查询

## 前置条件

| 操作类型 | 前置条件 | 说明 |
|---------|---------|------|
| CRUD（search/insert/update/delete） | 无特殊要求 | 直接通过 MySQL 读写 |
| esAgg（ES 聚合查询） | **必须开启高性能模式** | 数据异步写入 ES 后才能聚合查询 |

**高性能模式**：在表格设置中开启（`openHighMode=1`），开启后表格数据会异步同步到 ES 索引。

**context 要求**：所有方法需要 `context` 中包含 `_tenantId`，在工作流和场景环境中会自动注入。

## ⚠️ 重要：Java 对象访问方式

InfInvoker 返回的是 **Java 对象**，在 Nashorn 脚本中必须使用 Java 方法访问属性，不能使用 JavaScript 的点号语法。

| 错误写法（JS 语法） | 正确写法（Java 方法） |
|---------------------|----------------------|
| `record.data` | `record.getData()` |
| `record.id` | `record.getId()` |
| `data['1']` | `data.get('1')` |
| `response.success` | `response.getSuccess()` |
| `response.dataList` | `response.getDataList()` |

**示例对比**：

```javascript
// ❌ 错误：使用 JS 属性访问
var record = response.getDataList()[0];
var data = record.data;           // 返回 undefined
var value = data['1'];             // 报错

// ✅ 正确：使用 Java 方法
var record = response.getDataList()[0];
var data = record.getData();      // 返回 Java Map
var value = data.get('1');        // 正确获取字段值
```

**注意**：`getData()` 返回的是 Java Map，字段 key 是数字 ID 的字符串形式（如 `'1'`、`'17'`）。

## 方法速查表

| 方法 | 签名 | 返回值 | 说明 |
|------|------|--------|------|
| `search` | `(context, basicDataId, params)` | `InfResponse` | 按字段条件查询 |
| `queryById` | `(context, basicDataId, id)` | `InfResponse` | 按主键 ID 查询 |
| `queryByIds` | `(context, basicDataId, ids)` | `InfResponse` | 批量 ID 查询 |
| `convertKey` | `(context, basicDataId, params)` | `InfResponse` | 字段名转换（数字ID→字段名） |
| `insert` | `(context, basicDataId, params)` | `int` | 插入记录 |
| `update` | `(context, basicDataId, id, params)` | `int` | 更新记录 |
| `delete` | `(context, basicDataId, id)` | `int` | 删除记录 |
| `esAgg` | `(context, params, body)` | `InfEsAggResponse` | ES 8.15 异步聚合查询 |

## 查询操作

### search — 按字段条件查询

```javascript
InfInvoker.search(context, basicDataId, params)
```

**参数**:
- `context` (object): 上下文对象
- `basicDataId` (number): 表格 ID
- `params` (object): 查询条件，key 为字段 ID（字符串），value 为匹配值

**返回**: `InfResponse`

**示例**:
```javascript
// 查询"快麦出入库记录"表中单据类型="测试"的记录
var basicDataId = 1568;
var params = {"1": "测试"};  // fieldId=1 对应"单据类型"字段

var response = InfInvoker.search(context, basicDataId, params);
if (response.getSuccess() == false) {
    throw "查询失败: " + response.getMessage();
}
if (response.getDataList() == null || response.getDataList().length == 0) {
    throw "未查询到数据";
}
// 返回第一条记录的 data（使用 Java 方法）
return response.getDataList()[0].getData();
```

**注意**: params 的 key 是字段的数字 ID（字符串形式），不是字段中文名。

### queryById — 按主键 ID 查询

```javascript
InfInvoker.queryById(context, basicDataId, id)
```

**参数**:
- `context` (object): 上下文对象
- `basicDataId` (number): 表格 ID
- `id` (number): 记录主键 ID

**返回**: `InfResponse`

**示例**:
```javascript
var basicDataId = 1568;
var id = 23043610;

var response = InfInvoker.queryById(context, basicDataId, id);
if (response.getSuccess() == false) {
    throw "查询失败: " + response.getMessage();
}
if (response.getDataList() == null || response.getDataList().length == 0) {
    throw "未查询到记录";
}
return response.getDataList()[0].getData();
```

### queryByIds — 批量 ID 查询

```javascript
InfInvoker.queryByIds(context, basicDataId, ids)
```

**参数**:
- `context` (object): 上下文对象
- `basicDataId` (number): 表格 ID
- `ids` (array): 记录主键 ID 数组

**返回**: `InfResponse`

**示例**:
```javascript
var basicDataId = 1568;
var ids = [23043610, 23043657];

var response = InfInvoker.queryByIds(context, basicDataId, ids);
if (response.getSuccess() == false) {
    throw "查询失败: " + response.getMessage();
}
return response.getDataList();
```

**最佳实践**: 使用 `queryByIds` 代替循环调用 `queryById`，减少数据库访问次数。

### convertKey — 字段名转换

```javascript
InfInvoker.convertKey(context, basicDataId, params)
```

**参数**:
- `context` (object): 上下文对象
- `basicDataId` (number): 表格 ID
- `params` (object): 包含 `primaryValue` 和 `data` 的对象

**返回**: `InfResponse`

**说明**: 将表格内的数据按表格的字段名称转换为可读可解析的 JSON 数据。表格数据内部使用数字 ID 作为 key，convertKey 将其转换为字段名称。

**示例**:
```javascript
var basicDataId = 1568;
var param = {
    "primaryValue": "999",
    "data": {
        "0": "999",      // 序号
        "1": "测试",     // 单据类型
        "17": "1"        // 数量
    }
};

var response = InfInvoker.convertKey(context, basicDataId, param);
if (response.getSuccess() == false) {
    throw "转换失败: " + response.getMessage();
}
// 返回转换后的数据，key 变为字段名
return response.getData().getData();
// 结果: {"序号": "999", "单据类型": "测试", "数量": "1", ...}
```

## 写入操作

### insert — 插入记录

```javascript
InfInvoker.insert(context, basicDataId, params)
```

**参数**:
- `context` (object): 上下文对象
- `basicDataId` (number): 表格 ID
- `params` (object): 插入的数据，key 为字段 ID（字符串）

**返回**: `int` — 成功返回 1，失败返回 0 或 null

**示例**:
```javascript
var basicDataId = 1568;
var params = {
    "0": "1001",       // 序号
    "1": "销售出库",    // 单据类型
    "2": "出库",       // 出入库类型
    "17": "50",        // 数量
    "18": "主仓",      // 仓库
    "21": "张三"       // 操作人
};

var result = InfInvoker.insert(context, basicDataId, params);
if (result == null || result == 0) {
    throw "插入失败";
}
return "插入成功";
```

### update — 更新记录

```javascript
InfInvoker.update(context, basicDataId, id, params)
```

**参数**:
- `context` (object): 上下文对象
- `basicDataId` (number): 表格 ID
- `id` (number): 记录主键 ID
- `params` (object): 更新的数据，key 为字段 ID（字符串）

**返回**: `int` — 成功返回 1，失败返回 0 或 null

**示例**:
```javascript
var basicDataId = 1568;
var id = 23043610;
var params = {
    "17": "100",    // 更新数量为 100
    "11": "已审核"   // 更新备注
};

var result = InfInvoker.update(context, basicDataId, id, params);
if (result == null || result == 0) {
    throw "更新失败";
}
return "更新成功";
```

### delete — 删除记录

```javascript
InfInvoker.delete(context, basicDataId, id)
```

**参数**:
- `context` (object): 上下文对象
- `basicDataId` (number): 表格 ID
- `id` (number): 记录主键 ID

**返回**: `int` — 成功返回 1，失败返回 0 或 null

**示例**:
```javascript
var basicDataId = 1568;
var id = 23043610;

var result = InfInvoker.delete(context, basicDataId, id);
if (result == null || result == 0) {
    throw "删除失败";
}
return "删除成功";
```

## ES 聚合查询

### esAgg — ES 8.15 异步聚合查询

```javascript
InfInvoker.esAgg(context, params, body)
```

**前置条件**: 表格必须开启高性能模式（`openHighMode=1`）

**参数**:
- `context` (object): 上下文对象（需包含 `_tenantId`）
- `params` (string): URL 查询参数字符串
- `body` (string): ES 查询 DSL（JSON 字符串）

**返回**: `InfEsAggResponse`

**说明**: 直接操作 ES 索引（`lt-basic-data-{tenantId}`），不通过 basicDataId。可以使用 ES 8.15 的全部聚合能力。

### esAgg 使用方法

```javascript
// 1. 构建 params（URL 参数）
var params = "batched_reduce_size=64&wait_for_completion_timeout=200ms";

// 2. 构建 body（ES 查询 DSL）
var body = '{\n' +
    '  "aggs": {\n' +
    '    "warehouse_stats": {\n' +
    '      "terms": {\n' +
    '        "field": "data18.keyword",\n' +
    '        "size": 10\n' +
    '      }\n' +
    '    }\n' +
    '  },\n' +
    '  "size": 0,\n' +
    '  "query": {\n' +
    '    "bool": {\n' +
    '      "must": [],\n' +
    '      "filter": []\n' +
    '    }\n' +
    '  }\n' +
    '}';

var response = InfInvoker.esAgg(context, params, body);
if (response.getSuccess() == false) {
    throw "esAgg 失败: " + response.getMessage();
}
return response.getData();
```

### ES 字段命名规则

ES 索引中的字段名格式为 `data{fieldId}`：
- 普通字段: `data1`, `data2`, `data17` 等
- keyword 类型: 加 `.keyword` 后缀，如 `data18.keyword`
- 数值类型: 直接使用 `data17`

### 常见聚合模式

#### Terms 聚合 — 按字段分组统计

```javascript
// 按仓库统计出入库记录数
var body = '{\n' +
    '  "aggs": {\n' +
    '    "by_warehouse": {\n' +
    '      "terms": {\n' +
    '        "field": "data18.keyword",\n' +
    '        "size": 20,\n' +
    '        "order": {"_count": "desc"}\n' +
    '      }\n' +
    '    }\n' +
    '  },\n' +
    '  "size": 0\n' +
    '}';
```

#### Sum 聚合 — 数值求和

```javascript
// 统计各仓库的总数量
var body = '{\n' +
    '  "aggs": {\n' +
    '    "by_warehouse": {\n' +
    '      "terms": {"field": "data18.keyword", "size": 20},\n' +
    '      "aggs": {\n' +
    '        "total_qty": {"sum": {"field": "data17"}}\n' +
    '      }\n' +
    '    }\n' +
    '  },\n' +
    '  "size": 0\n' +
    '}';
```

#### Date Histogram 聚合 — 按时间分组

```javascript
// 按天统计操作记录
var body = '{\n' +
    '  "aggs": {\n' +
    '    "by_date": {\n' +
    '      "date_histogram": {\n' +
    '        "field": "data22",\n' +
    '        "calendar_interval": "day"\n' +
    '      }\n' +
    '    }\n' +
    '  },\n' +
    '  "size": 0\n' +
    '}';
```

## 响应模型

### InfResponse

InfResponse 是 **Java 对象**，必须使用 Java 方法访问属性：

```javascript
// Java 对象结构
{
    success: true/false,     // 使用 getSuccess() 访问
    message: "错误信息",     // 使用 getMessage() 访问
    dataList: [              // 使用 getDataList() 访问
        {
            id: 23043610,    // 使用 getId() 访问
            data: {          // 使用 getData() 访问，返回 Java Map
                "0": "999",  // 使用 data.get("0") 访问字段
                "1": "测试",
                "17": "1"
            }
        }
    ]
}
```

**方法**:
- `getSuccess()` — 获取是否成功
- `getMessage()` — 获取失败原因
- `getDataList()` — 获取数据列表
- 列表中每项：`getId()` 获取主键，`getData()` 获取数据 Map
- 数据 Map：`data.get("字段ID")` 获取字段值

### InfEsAggResponse

```javascript
{
    success: true/false,
    message: "错误信息",
    data: {                  // ES 聚合结果
        "aggregations": {
            "warehouse_stats": {
                "buckets": [...]
            }
        }
    }
}
```

**方法**:
- `getSuccess()` — 获取是否成功
- `getMessage()` — 获取失败原因
- `getData()` — 获取聚合数据

## 常见陷阱与最佳实践

### 陷阱

| 陷阱 | 说明 | 解决方案 |
|------|------|----------|
| 使用 JS 属性访问 | `record.data` 返回 undefined | 使用 Java 方法：`record.getData()` |
| 字段名用中文 | `{"单据类型": "测试"}` 会报错 | 使用字段 ID：`{"1": "测试"}` |
| esAgg 未开启高性能 | 返回空结果或报错 | 在表格设置中开启高性能模式 |
| 循环中逐条查询 | 性能差 | 使用 `queryByIds` 批量查询 |
| ES5.1 语法错误 | 使用 let/const/箭头函数报错 | 使用 var 和 function |
| esAgg body 格式 | JSON 字符串拼接易出错 | 使用 `\n` 换行，确保 JSON 合法 |

### 最佳实践

1. **始终检查响应状态**: `if (response.getSuccess() == false) throw ...`
2. **空值检查**: `if (response.getDataList() == null || response.getDataList().length == 0)`
3. **批量查询**: 使用 `queryByIds` 代替循环 `queryById`
4. **字段 ID 查询**: 使用 `lingtong-cli table schema query --basic-data-id <id>` 获取字段 ID
5. **ES 聚合调试**: 在 Kibana 中构建查询，复制 body 到脚本中
