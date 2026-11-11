# 连接器接口调用

单独调用连接器方法使用 `lingtong-cli connector invoke`。CLI 通过通用代理调用新接口，不创建或执行工作流。

## 调用方式

```bash
lingtong-cli connector invoke \
  --connector kmerp \
  --method erp.warehouse.list.query \
  --auth-account "广州力人服饰" \
  --env test \
  --body '{}' \
  --format pretty
```

`--connector`、`--method` 必填；`--auth-account` 为账户名称，内置连接器可省略，服务端校验授权。`--env` 默认 `prod`；`--body` 默认 `{}`，必须为非 null 的 JSON 对象，传业务参数即可，无需手动包裹 `requestBody`。

CLI 向 `/gw/ai/proxy` 提交 `method: POST`、`path: /gw/connector/invoke`，代理转发至控制器 `/connector/invoke`（对应外部接口 `/ltapp/connector/invoke`）。请求字段为 `connector`、`method`、`env`、`authName`、`requestBody`，后两者分别来自 `--auth-account` 和 `--body`。

## 查询参数与读取结果

先用 `connector account list` 确认账户，用 `connector methods` 查询方法，用 `connector schema` 获取请求字段。详细命令见 [连接器技能](../../lingtong-connector/SKILL.md)。

CLI 完整保留接口响应：

```json
{
  "success": true,
  "result": {
    "receiptStatus": "SUCCESS",
    "responseBody": {"list": [], "total": 0}
  }
}
```

外层成功不代表业务成功，需检查 `result.receiptStatus` 和业务状态；从 `result.responseBody` 读取业务数据，支持对象或数组。CLI 不自动重试，执行结果未知时先核实。

`--force` 仅保留为弃用的空操作。`connector cache list/clear` 仅管理旧本地缓存，不影响新调用，不删除平台历史工作流。

## 工作流脚本中的 AppInvoker（与 CLI 响应结构不同）

以快麦 ERP trade 类接口（`erp.trade.outstock.simple.query`、`erp.trade.list.query` 等）为例：

1. **请求参数传【扁平】，不要手动包 `requestBody`**。`AppInvoker.invoke` 的 body 运行时会自动包裹进
   `requestBody`；若自己再包一层会变成 `requestBody.requestBody`（双重包裹）→ ERP 收到空参数 →
   返回全 null（`TradeOutStockListQueryResponse(total=null,...)`）。工作流脚本节点应传扁平参数；CLI 的 `--body` 同样传业务参数，由 CLI 构造接口请求：
   ```js
   // ✅ 正确：扁平
   var body = {timeType:'created', startTime: s, endTime: e, pageSize: 100, pageNo: 1};
   var result = AppInvoker.invoke(context, 'erp.trade.outstock.simple.query', '广州力人服饰', body);
   // ❌ 错误：{requestBody:{...}} -> 双重包裹 -> 查不到数据
   ```
   用 `connector schema --connector <c> --method <m> --auth-account-id <id>` 看字段名（`requestFieldList`
   的 `fullPath` 形如 `requestBody.startTime`，确认要传哪些参数；但调用时传扁平即可，无需自己加 `requestBody` 前缀）。
2. **pageSize 不能太小**。kmerp 出库/订单接口对过小的 pageSize 会返回 null（实测 `pageSize=5` 失败、
   `20/50/100/200` 正常）。建议默认 100，并对 `<20` 的值兜底。
3. **成功响应直接是 `responseData` 对象**：字段在顶层（`{list, total, success, ...}`），读
   `result.list` / `result.total`（Nashorn 会按 JavaBean 规则把 `.list` 映射到 `getList()`），
   **不是** `result.result.list`。
4. **空/失败响应会抛异常**：无数据时连接器节点抛 `节点执行失败:XxxResponse(total=null,...)`，会中断整个工作流。
   查询脚本应 `try/catch`，仅对「空结果」(`total=null`) 降级为空数组、其它异常向上抛出：
   ```js
   var list = [], total = 0;
   try {
     var r = AppInvoker.invoke(context, method, account, body);
     if (r != null) { if (r.list != null) list = r.list; if (r.total != null) total = r.total; }
   } catch (e) { if (('' + e).indexOf('total=null') < 0) { throw e; } }
   return {list: list, total: total};
   ```

> 排错顺序：先用 `connector invoke` 单点验证接口能否返回数据（与工作流隔离），再排查工作流编排。
> 单节点级调试见 [testing-guide.md](testing-guide.md) 的「单节点调试 /workflow/debug/node/do」。

## 相关文档

- [脚本编写指南](script-guide.md)
- [节点详解](node-details.md)
