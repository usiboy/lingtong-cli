# 脚本编写指南

本文档说明绫通工作流脚本节点的编写规范。

## 语法限制

- **引擎**: Nashorn (JavaScript)
- **标准**: ECMAScript 5.1
- **不支持**: ES6+ 语法（箭头函数、let/const、模板字符串、解构等）

### 支持的语法

```javascript
// 变量声明
var name = "hello";
var count = 10;
var items = [1, 2, 3];

// 函数
function add(a, b) {
    return a + b;
}

// 循环
for (var i = 0; i < items.length; i++) {
    console.log(items[i]);
}

// 条件
if (count > 5) {
    // ...
} else {
    // ...
}
```

### 不支持的语法

```javascript
// ❌ 箭头函数
const add = (a, b) => a + b;

// ❌ let/const
let x = 10;
const y = 20;

// ❌ 模板字符串
const msg = `Hello ${name}`;

// ❌ 解构赋值
const {a, b} = obj;

// ❌ Promise/async/await
async function fetch() { ... }
```

## context API

脚本通过 `context` 对象访问输入变量。

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

## inputVariables 配置

脚本节点必须通过 `inputVariables` 声明输入变量：

```json
{
  "id": "w_script_transform",
  "type": "w_script",
  "data": {
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
      "script": "var items = context.get('rawData');\nreturn items;"
    }
  }
}
```

## AppInvoker

调用连接器接口：

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

### AppInvoker 参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `context` | object | 上下文对象 |
| `method` | string | 连接器方法名 |
| `authAccount` | string | 认证账户名称 |
| `requestBody` | object | 请求参数 |

### 必需设置

调用 AppInvoker 前必须设置 `_env`：

```javascript
context.put("_env", "test");  // 或 "prod"
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

### 错误处理

```javascript
try {
    var result = AppInvoker.invoke(context, "method", "account", {});
    return result;
} catch (e) {
    console.log("调用失败: " + e);
    throw "调用连接器失败: " + e;
}
```

## 注意事项

1. **使用 var 声明变量**，不要使用 let/const
2. **使用 context.get()** 获取输入变量
3. **使用 return** 返回结果
4. **设置 _env** 调用 AppInvoker 前
5. **ES5.1 语法**，不要使用 ES6+ 特性
