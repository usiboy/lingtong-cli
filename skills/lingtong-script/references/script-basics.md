# 脚本基础

本文档说明绫通脚本的语法规范、context API 和内置对象。

## 语法限制

- **引擎**: Nashorn (JavaScript)
- **标准**: ECMAScript 5.1
- **不支持**: ES6+ 语法

### 支持的语法

```javascript
// 变量声明（必须使用 var）
var name = "hello";
var count = 10;
var items = [1, 2, 3];

// 函数声明
function add(a, b) {
    return a + b;
}

// 循环
for (var i = 0; i < items.length; i++) {
    log.info(items[i]);
}

// 条件
if (count > 5) {
    // ...
} else {
    // ...
}

// 对象
var obj = {
    key: "value",
    nested: {a: 1}
};
```

### 不支持的语法

```javascript
// ❌ 箭头函数
var add = (a, b) => a + b;

// ❌ let/const
let x = 10;
const y = 20;

// ❌ 模板字符串
var msg = `Hello ${name}`;

// ❌ 解构赋值
var {a, b} = obj;

// ❌ Promise/async/await
async function fetch() { ... }

// ❌ 扩展运算符
var merged = {...obj1, ...obj2};

// ❌ 类语法
class MyClass { ... }
```

## context API

脚本通过 `context` 对象访问输入变量和上下文信息。

### 获取变量

```javascript
var orderId = context.get("orderId");
var rawData = context.get("rawData");
```

### 设置变量

```javascript
context.put("result", someValue);
context.put("processed", true);
```

### 检查变量

```javascript
if (context.containsKey("orderId")) {
    var id = context.get("orderId");
}
```

### 获取系统信息

```javascript
var tenantId = context.get("_tenantId");  // 租户 ID（InfInvoker 需要）
var env = context.get("_env");            // 环境（test/prod）
```

## 内置对象速查表

| 对象 | 工作流 | 场景函数 | 连接器函数 | 说明 |
|------|:---:|:---:|:---:|------|
| `context` | ✅ | ✅ | ✅ | 上下文对象 |
| `InfInvoker` | ✅ | ✅ | ✅ | 表格 CRUD + ES 聚合 |
| `AppInvoker` | ✅ | ✅ | ❌ | 连接器接口调用 |
| `PushStateInvoker` | ✅ | ✅ | ❌ | 推送状态查询 |
| `FlowDataInvoker` | ✅ | ✅ | ❌ | 工作流数据查询 |
| `log` | ✅ | ✅ | ✅ | 日志对象 |
| `console` | ✅ | ✅ | ❌ | 控制台输出 |

## AppInvoker

调用连接器接口。

```javascript
// 设置环境（必需）
context.put("_env", "test");

// 调用连接器
var result = AppInvoker.invoke(
    context,
    "erp.warehouse.list.query",  // 方法名
    "广州力人服饰",              // 账户名称
    {"pageSize": 10}             // 请求参数
);

return result;
```

**注意**: 调用 AppInvoker 前必须设置 `_env`。

## log 对象

输出日志信息。

```javascript
log.info("处理开始");
log.debug("变量值: " + someVar);
log.warn("警告信息");
log.error("错误信息");
```

## 返回值规则

- **工作流脚本节点**: 使用 `return` 返回值，作为节点输出
- **场景字段函数**: 使用 `return` 返回值，作为字段映射结果
- **连接器字段函数**: 使用 `return` 返回值，作为请求参数

## 错误处理

```javascript
try {
    var response = InfInvoker.search(context, 1568, {"1": "测试"});
    if (response.getSuccess() == false) {
        throw "查询失败: " + response.getMessage();
    }
    return response.getDataList();
} catch (e) {
    log.error("执行失败: " + e);
    throw "脚本执行失败: " + e;
}
```

## 常见模式

### 数据转换

```javascript
var items = context.get("rawData");
var result = [];

for (var i = 0; i < items.length; i++) {
    result.push({
        id: items[i].id,
        name: items[i].name,
        code: items[i].code
    });
}

return result;
```

### 条件过滤

```javascript
var items = context.get("items");
var filtered = [];

for (var i = 0; i < items.length; i++) {
    if (items[i].status === "active") {
        filtered.push(items[i]);
    }
}

return filtered;
```

### 数据聚合

```javascript
var items = context.get("items");
var total = 0;

for (var i = 0; i < items.length; i++) {
    total = total + items[i].amount;
}

return {total: total, count: items.length};
```

## 注意事项

1. **使用 var 声明变量**，不要使用 let/const
2. **使用 context.get()** 获取输入变量
3. **使用 return** 返回结果
4. **设置 _env** 调用 AppInvoker 前
5. **ES5.1 语法**，不要使用 ES6+ 特性
6. **字符串拼接** 使用 `+`，不要使用模板字符串
