# 测试指南

本文档说明如何测试绫通工作流。

## 单节点调试 /workflow/debug/node/do（强烈推荐排查脚本节点）

绫通控制台支持**只调试单个代码节点**，是定位「某节点输出不对/连接器查不到数据」的最快手段。
它是 SSE（`text/event-stream`）接口，用浏览器登录态的 `Server-Token` Cookie 鉴权（非 apk Token）：

```bash
curl -sN 'https://<host>/admin/ltappboot/workflow/debug/node/do' \
  -H 'accept: text/event-stream' -H 'content-type: application/json' \
  -b "Server-Token=<从浏览器会话获取>" \
  --data '{"workflowNumber":"W1022","nodeId":"w_script_query",
           "params":{"w_start_first":{"startTime":"2026-06-21 00:00:00","endTime":"2026-06-22 23:59:59","pageSize":100,"pageNo":1}}}'
```

要点：
- `params` 按 **nodeId → 该节点输出** 注入。调试下游节点时可直接喂上游输出，例如调试 transform：
  `"params":{"w_script_query":{"result":{...上游结果...}}}`（无需真正跑上游）。
- 单节点调试**不会自动运行上游节点**，下游依赖的输入要么自己注入，要么改调更上游的 nodeId。
- 最终结果在 `status:"complete"` 事件的 `outputVariable` 字段；脚本里 `return` 的对象即在此。
- 调试时可临时把脚本改成「返回诊断信息」（响应类型、keys、捕获的异常字符串）再 `workflow update` 推上去，定位后改回。
- **草稿即调**：debug 跑的是草稿（`workflow update` 后立即生效），无需 publish；而 `workflow api-test` 跑的是**已发布**版本。

## 测试命令

### 执行工作流

```bash
# 基础执行
lingtong-cli workflow execute --workflow-id 955

# 带参数执行
lingtong-cli workflow execute --workflow-id 955 --params '{"key":"value"}'

# 等待执行完成
lingtong-cli workflow execute --workflow-id 955 --wait
```

### 测试运行

```bash
# 简单测试
lingtong-cli workflow test run --workflow-id 955

# 带参数测试
lingtong-cli workflow test run --workflow-id 955 --params '{"code":"WH001"}'

# 从文件加载测试数据
lingtong-cli workflow test run --workflow-id 955 --test-data-file test.json

# 详细输出
lingtong-cli workflow test run --workflow-id 955 --params '{}' --verbose
```

### 查看执行日志

```bash
# 查询日志
lingtong-cli workflow logs --receipt-id abc123

# 分页查询
lingtong-cli workflow logs --receipt-id abc123 --page 1 --size 20
```

## 测试 Open API

```bash
# 启用 API
lingtong-cli workflow api-enable --workflow-id 956

# 测试 API
lingtong-cli workflow api-test --app-tag <tag> --params '{"key":"value"}'
```

## 测试流程

1. **验证 DSL**: `workflow validate --dsl-file workflow.json`
2. **创建工作流**: `workflow create --name "测试" --dsl-file workflow.json`
3. **执行测试**: `workflow test run --workflow-id <id>`
4. **查看日志**: `workflow logs --receipt-id <id>`
5. **发布**: `workflow publish --workflow-id <id>`

## 测试数据文件

测试数据文件格式：

```json
{
  "orderId": "ORDER001",
  "amount": 1000,
  "items": [
    {"id": 1, "name": "商品A"},
    {"id": 2, "name": "商品B"}
  ]
}
```

## 常见问题

| 问题 | 解决方案 |
|------|----------|
| 执行超时 | 检查连接器响应时间 |
| 参数错误 | 检查 Start 节点参数定义 |
| 脚本错误 | 检查 ES5.1 语法 |
| 连接器失败 | 检查 authAccount 名称 |
