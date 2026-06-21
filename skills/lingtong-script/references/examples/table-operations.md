# 表格操作示例

基于"快麦出入库记录"表（basicDataId=1568）的完整工作流脚本示例。

> **⚠️ 重要**: InfInvoker 返回的是 Java 对象，必须使用 Java 方法访问属性：
> - `record.getData()` 而不是 `record.data`
> - `data.get("1")` 而不是 `data["1"]`
> - `record.getId()` 而不是 `record.id`

## 表格信息

**快麦出入库记录**（basicDataId=1568, openHighMode=1）

| fieldId | 字段名 | 类型 | 说明 |
|---------|--------|------|------|
| 0 | 序号 | number | 主键 |
| 1 | 单据类型 | text | 如"测试"、"销售出库" |
| 2 | 出入库类型 | text | 如"出库"、"入库" |
| 3 | 系统订单号 | text | |
| 4 | 平台订单号 | text | |
| 5 | 商品名称/规格 | text | |
| 6 | 条形码 | text | |
| 17 | 数量 | number | |
| 18 | 仓库 | text | |
| 19 | 店铺 | text | |
| 21 | 操作人 | text | |
| 22 | 操作时间 | date | 时间戳（毫秒） |
| 24 | 操作后总库存 | number | |
| 25 | 重量 | number | |
| 26 | 体积 | number | |

**测试数据**:
- id=23043610: 序号=999, 单据类型=测试, 数量=1, 操作后总库存=79
- id=23043657: 序号=999999, 单据类型=测试单条, 数量=100, 操作后总库存=79

## 示例 1: 按单据类型查询出入库记录

**场景**: 查询指定单据类型的所有出入库记录

```javascript
// 查询"快麦出入库记录"表中单据类型="测试"的记录
var basicDataId = 1568;
var params = {"1": "测试"};  // fieldId=1 对应"单据类型"

var response = InfInvoker.search(context, basicDataId, params);
if (response.getSuccess() == false) {
    throw "查询失败: " + response.getMessage();
}

var dataList = response.getDataList();
if (dataList == null || dataList.length == 0) {
    return {"message": "未找到相关记录"};
}

// 提取关键字段（使用 Java 方法访问）
var result = [];
for (var i = 0; i < dataList.length; i++) {
    var record = dataList[i];
    var data = record.getData();  // 返回 Java Map
    result.push({
        "id": record.getId(),
        "序号": data.get("0"),
        "单据类型": data.get("1"),
        "出入库类型": data.get("2"),
        "数量": data.get("17"),
        "仓库": data.get("18"),
        "操作人": data.get("21")
    });
}

return {
    "total": result.length,
    "records": result
};
```

## 示例 2: 按仓库统计库存汇总（ES 聚合）

**前置条件**: 表格已开启高性能模式（openHighMode=1）

**场景**: 统计各仓库的出入库记录数和总数量

```javascript
// ES 聚合查询参数
var params = "batched_reduce_size=64&wait_for_completion_timeout=200ms";

// ES 查询 DSL：按仓库分组，统计记录数和总数量
var body = '{\n' +
    '  "aggs": {\n' +
    '    "by_warehouse": {\n' +
    '      "terms": {\n' +
    '        "field": "data18.keyword",\n' +
    '        "size": 50,\n' +
    '        "order": {"_count": "desc"}\n' +
    '      },\n' +
    '      "aggs": {\n' +
    '        "total_qty": {\n' +
    '          "sum": {"field": "data17"}\n' +
    '        },\n' +
    '        "avg_stock": {\n' +
    '          "avg": {"field": "data24"}\n' +
    '        }\n' +
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
    throw "ES 聚合失败: " + response.getMessage();
}

var aggData = response.getData();
if (aggData == null) {
    return {"message": "未查询到聚合数据"};
}

// 解析聚合结果
var buckets = aggData.aggregations.by_warehouse.buckets;
var result = [];
for (var i = 0; i < buckets.length; i++) {
    result.push({
        "仓库": buckets[i].key,
        "记录数": buckets[i].doc_count,
        "总数量": buckets[i].total_qty.value,
        "平均库存": buckets[i].avg_stock.value
    });
}

return {
    "warehouses": result.length,
    "details": result
};
```

## 示例 3: 批量查询并处理

**场景**: 根据 ID 列表批量查询记录，计算总数量

```javascript
var basicDataId = 1568;
var ids = [23043610, 23043657];

var response = InfInvoker.queryByIds(context, basicDataId, ids);
if (response.getSuccess() == false) {
    throw "查询失败: " + response.getMessage();
}

var dataList = response.getDataList();
if (dataList == null || dataList.length == 0) {
    return {"message": "未查询到记录"};
}

// 计算总数量和平均库存
var totalQty = 0;
var totalStock = 0;
var records = [];

for (var i = 0; i < dataList.length; i++) {
    var record = dataList[i];
    var data = record.getData();  // 返回 Java Map
    var qty = parseInt(data.get("17")) || 0;
    var stock = parseInt(data.get("24")) || 0;
    
    totalQty = totalQty + qty;
    totalStock = totalStock + stock;
    
    records.push({
        "id": record.getId(),
        "序号": data.get("0"),
        "单据类型": data.get("1"),
        "数量": qty,
        "操作后总库存": stock
    });
}

return {
    "count": records.length,
    "totalQty": totalQty,
    "avgStock": totalStock / records.length,
    "records": records
};
```

## 示例 4: 查询 + 更新组合

**场景**: 查询指定条件的记录，更新其备注字段

```javascript
var basicDataId = 1568;

// Step 1: 查询记录
var searchParams = {"1": "测试"};  // 单据类型="测试"
var searchResponse = InfInvoker.search(context, basicDataId, searchParams);
if (searchResponse.getSuccess() == false) {
    throw "查询失败: " + searchResponse.getMessage();
}

var dataList = searchResponse.getDataList();
if (dataList == null || dataList.length == 0) {
    return {"message": "未找到需要更新的记录"};
}

// Step 2: 逐条更新备注
var updateCount = 0;
for (var i = 0; i < dataList.length; i++) {
    var recordId = dataList[i].getId();  // 使用 Java 方法
    var updateParams = {
        "11": "已批量更新"  // 备注字段
    };
    
    var result = InfInvoker.update(context, basicDataId, recordId, updateParams);
    if (result != null && result > 0) {
        updateCount++;
    }
}

return {
    "searched": dataList.length,
    "updated": updateCount
};
```

## CLI 辅助命令

```bash
# 查看表格列表（确认 basicDataId）
lingtong-cli table list --app-id 165 --format pretty

# 查看表格 schema（获取字段 ID 映射）
lingtong-cli table schema query --basic-data-id 1568 --format pretty

# 查询表格数据（验证 InfInvoker 查询结果）
lingtong-cli table data query --basic-data-id 1568 --page-size 5 --format pretty

# 按条件筛选
lingtong-cli table data query --basic-data-id 1568 --filter '{"1":"测试"}' --format pretty
```
