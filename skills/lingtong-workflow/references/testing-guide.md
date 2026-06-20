# 测试指南

本文档说明如何测试绫通工作流。

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
