---
name: lingtong-workflow
version: 2.0.0
description: "绫通工作流全生命周期管理:设计、创建、验证、测试、发布、监控、文档生成。当用户需要编排业务流程、管理工作流DSL、执行工作流测试、生成文档时触发。关键词:workflow、工作流、DSL、模板、验证、测试、文档。"
---

# lingtong-workflow 技能

## 概述

本技能提供绫通平台工作流的**全生命周期管理**能力,涵盖:
- 工作流设计与DSL编排
- 模板管理与快速创建
- DSL验证与依赖分析
- 测试执行与报告
- 发布与版本管理
- 文档自动生成

## 工作流节点指南

绫通工作流支持 12 种节点类型:

### w_start - 开始节点
- **规则**: 每个工作流必须有且仅有一个
- **功能**: 定义工作流输入变量
- **配置**: outputVariables 数组

```json
{
  "id": "w_start_first",
  "type": "w_start",
  "data": {
    "outputVariables": [
      {
        "variable": "orderId",
        "variableAttr": {
          "dataType": "string",
          "label": "订单ID",
          "required": true
        }
      }
    ]
  }
}
```

### w_end - 结束节点
- **规则**: 至少一个,可多个(成功/失败路径)
- **功能**: 定义工作流输出
- **配置**: 从上游节点提取输出变量

### w_connector - 连接器节点
- **功能**: 调用外部系统 API
- **必需配置**: connector, interfaceModelId, authAccountId
- **分支**: 支持 true(成功)/false(失败) 双路径
- **参数映射**: 
  - 普通连接器节点 (w_connector): 通过前端界面"参数映射设置"配置，DSL 中不直接设置 fieldMapping
  - 管道节点 (w_modePipe): 支持 pipeConfig 和 queryParams 配置，用于批量数据同步
  - **重要**: 创建连接器节点时，只需配置基本连接器信息，参数映射通过前端界面设置

```json
{
  "id": "w_connector_sales",
  "type": "w_connector",
  "data": {
    "title": "销售出库单查询",
    "connector": "kmerp",
    "interfaceModelId": 73,
    "domainModelId": 3,
    "authAccountId": 296,
    "env": "test",
    "catId": "交易",
    "assertConfig": {"assertType": "throwException"},
    "outputVariables": [
      {
        "variable": "response",
        "variableAttr": {
          "dataType": "object",
          "title": "销售出库单查询返回",
          "value": "$.",
          "nodeId": "w_connector_sales"
        }
      }
    ]
  }
}
```

**参数映射配置说明**:
- 连接器接口的请求参数通过前端界面的"参数映射设置"对话框配置
- 每个参数可以映射到上游节点的输出变量或输入固定值
- DSL 中不需要包含 fieldMapping 或 queryParams 字段
- 前端会根据 interfaceModelId 自动加载接口参数列表供用户配置

### w_modePipe - 数据管道节点
- **功能**: 批量数据同步,自动处理分页和游标
- **模式**: count(计数), time(时间), cursor(游标)
- **输出**: 数组,通常配合 w_dataSplit 使用

### w_dataSplit - 数据拆分节点
- **功能**: 将数组遍历,每次输出单个对象
- **输入**: 上游数组变量
- **输出**: 单个对象,供下游逐条处理

### w_script - 代码执行节点
- **语言**: JavaScript ES5.1 (Nashorn引擎)
- **API**: context.get()/context.put()/console.log()
- **限制**: 不支持ES6+,不支持异步操作

### w_if - 条件分支节点
- **功能**: 根据条件表达式选择执行路径
- **分支**: true/false

### w_switch - 多路分支节点
- **功能**: 多条件选择
- **分支**: case_1, case_2, ..., default

### w_cycle - 循环节点
- **功能**: 重复执行子流程
- **控制**: 条件表达式控制循环终止

### w_joinPipeline - 汇合节点
- **功能**: 聚合多个分支的执行结果

### w_dataPush - 数据推送节点
- **功能**: 带重试机制的数据推送
- **配置**: 重试次数、补偿策略

### w_pushRecord - 推送记录节点
- **功能**: 查询推送历史记录

## DSL 编排规则

### 变量引用语法
```
格式: $nodeId.variableName

示例:
  $w_start_first.orderId              # 从开始节点获取 orderId
  $w_connector_1k65t.response.list    # 从连接器响应提取 list
  $w_modePipe_725vx.trades            # 从管道节点获取 trades 数组
```

### 边连接规则
- w_start: 只能有一条出边 (source 类型)
- w_connector: 可有两出边 (true/false)
- w_if: 可有多出边 (true/false)
- w_dataSplit: 一条出边 (true)
- w_end: 不能有出边

### 数据流规则
- 数据从上游流向下游,不可逆向
- 循环引用会导致执行错误
- 数组节点配合 w_dataSplit 实现遍历处理

### 最佳实践
1. 始终配置 assertConfig 错误处理
2. 关键节点配置补偿机制
3. 使用有意义的节点 ID (不要使用自动生成的随机ID)
4. 为变量添加清晰的 label 和 dataType
5. 工作流超过20个节点时考虑拆分为子工作流

## 工作流模板

### 可用模板

| 模板名 | 节点数 | 用途 | 适用场景 |
|--------|--------|------|----------|
| simple | 2 | 开始→结束 | 测试、学习 |
| connector | 3 | 开始→连接器→结束 | 简单API调用 |
| order_sync | 5 | 订单同步流程 | ERP订单同步到其他系统 |
| approval | 4 | 审批流程 | 条件审批工作流 |
| data_pipeline | 5 | 数据管道流程 | 批量数据处理与转换 |
| api_wrapper | 4 | API包装流程 | 将连接器接口暴露为API |

### 使用模板

```bash
# 列出所有模板
lingtong-cli workflow template list

# 查看模板详情
lingtong-cli workflow template show --name order_sync

# 使用模板生成DSL文件
lingtong-cli workflow template use --name order_sync --output workflow.json

# 使用模板直接创建工作流
lingtong-cli workflow create --name "订单同步" --template order_sync
```

## 完整命令参考

### 基础命令

```bash
# 列出工作流
lingtong-cli workflow list --app-id 165

# 查询工作流详情
lingtong-cli workflow info --workflow-id 955

# 执行工作流
lingtong-cli workflow execute --workflow-id 955

# 查询执行日志
lingtong-cli workflow logs --receipt-id abc123
```

### 模板管理

```bash
# 模板列表
lingtong-cli workflow template list

# 模板详情
lingtong-cli workflow template show --name api_wrapper

# 生成DSL文件
lingtong-cli workflow template use --name approval --output approval.json
```

### 创建工作流

```bash
# 使用模板创建
lingtong-cli workflow create --name "审批流程" --template approval

# 使用DSL文件创建
lingtong-cli workflow create --name "自定义流程" --dsl-file workflow.json

# 使用DSL字符串创建
lingtong-cli workflow create --name "简单流程" --dsl-string '{"nodes":[...],"edges":[...]}'
```

### 验证DSL

```bash
# 验证服务器上的工作流
lingtong-cli workflow validate --workflow-id 955

# 验证本地DSL文件
lingtong-cli workflow validate --dsl-file workflow.json

# 严格模式(包含警告)
lingtong-cli workflow validate --dsl-file workflow.json --strict
```

**验证规则**:
- 必须包含 nodes 和 edges 字段
- 节点类型必须是12种合法类型之一
- 必须有且仅有一个 w_start 节点
- 至少有一个 w_end 节点
- 边的 source/target 必须引用存在的节点
- connector/modePipe 节点必须配置 connector 字段

### 依赖分析

```bash
# 列出工作流依赖
lingtong-cli workflow dependency list --workflow-id 955

# 从本地DSL分析
lingtong-cli workflow dependency list --dsl-file workflow.json
```

**输出内容**:
- 连接器依赖 (connector, interfaceModelId, authAccountId)
- 脚本依赖 (nodeId, language)
- 依赖计数统计

### 测试执行

```bash
# 简单测试
lingtong-cli workflow test run --workflow-id 955 --params '{"key":"value"}'

# 从文件加载测试数据
lingtong-cli workflow test run --workflow-id 955 --test-data-file test.json

# 详细输出
lingtong-cli workflow test run --workflow-id 955 --params '{"code":"WH001"}' --verbose
```

### 文档生成

```bash
# 生成Markdown文档
lingtong-cli workflow doc generate --workflow-id 955 --output-file doc.md

# 包含API文档
lingtong-cli workflow doc generate --workflow-id 956 --include-api --output-file api-doc.md

# 包含调用示例
lingtong-cli workflow doc generate --workflow-id 955 --include-examples --output-file full-doc.md

# 从本地DSL生成
lingtong-cli workflow doc generate --dsl-file workflow.json --output-file doc.md
```

### 发布与版本管理

```bash
# 发布工作流
lingtong-cli workflow publish --workflow-id 955 --version "v1.0.0" --memo "初始版本"

# 查看版本历史
lingtong-cli workflow versions --workflow-id 955

# 回滚到指定版本
lingtong-cli workflow version rollback --workflow-id 955 --version "v1.0.0"

# 预览回滚(不执行)
lingtong-cli workflow version rollback --workflow-id 955 --version "v1.0.0" --dry-run
```

### API管理

```bash
# 启用API访问
lingtong-cli workflow api-enable --workflow-id 956

# 禁用API访问
lingtong-cli workflow api-disable --workflow-id 956

# 测试开放API
lingtong-cli workflow api-test --app-tag b2nUqsNASykMfq --params '{"code":"WH001"}'
```

### 更新与删除

```bash
# 更新工作流DSL
lingtong-cli workflow update --workflow-id 955 --dsl-file updated.json

# 删除工作流
lingtong-cli workflow delete --workflow-id 955 --confirm
```

## 典型工作流示例

### 示例1: 连接器API化

将快麦ERP的仓库查询接口包装为独立API:

```json
{
  "nodes": [
    {
      "id": "w_start_first",
      "type": "w_start",
      "data": {
        "outputVariables": [
          {"variable": "code", "variableAttr": {"dataType": "string"}},
          {"variable": "name", "variableAttr": {"dataType": "string"}}
        ]
      }
    },
    {
      "id": "w_connector_1k65t",
      "type": "w_connector",
      "data": {
        "connector": "kmerp",
        "interfaceModelId": 119,
        "authAccountId": 296,
        "fieldMapping": [
          {"sourceField": "$w_start_first.code", "targetFieldId": 4693},
          {"sourceField": "$w_start_first.name", "targetFieldId": 4694}
        ],
        "outputVariables": [
          {"variable": "response", "variableAttr": {"value": "$."}}
        ]
      }
    },
    {
      "id": "w_end_success",
      "type": "w_end",
      "data": {
        "outputVariables": [
          {"variable": "warehouses", "variableAttr": {
            "dataType": "array",
            "value": "$w_connector_1k65t.response.list[*]"
          }}
        ]
      }
    }
  ],
  "edges": [
    {"source": "w_start_first", "target": "w_connector_1k65t"},
    {"source": "w_connector_1k65t", "target": "w_end_success"}
  ]
}
```

**创建并发布**:
```bash
# 1. 创建工作流
lingtong-cli workflow create --name "仓库查询API" --template api_wrapper

# 2. 验证DSL
lingtong-cli workflow validate --workflow-id <new_id>

# 3. 启用API
lingtong-cli workflow api-enable --workflow-id <new_id>

# 4. 发布
lingtong-cli workflow publish --workflow-id <new_id> --version "v1.0.0"

# 5. 测试API
lingtong-cli workflow api-test --app-tag <appTag> --params '{"code":"WH001","name":"测试"}'
```

### 示例2: 订单同步流程

从快麦ERP同步订单到其他系统:

```bash
# 1. 使用订单同步模板创建
lingtong-cli workflow create --name "订单同步" --template order_sync

# 2. 分析依赖
lingtong-cli workflow dependency list --workflow-id <new_id>

# 3. 测试执行
lingtong-cli workflow test run --workflow-id <new_id> --params '{"startDate":"2026-04-01"}'

# 4. 生成文档
lingtong-cli workflow doc generate --workflow-id <new_id> --output-file order-sync-doc.md

# 5. 发布
lingtong-cli workflow publish --workflow-id <new_id> --version "v1.0.0" --memo "订单同步v1"
```

## 常见问题排查

| 问题 | 可能原因 | 解决方案 |
|------|----------|----------|
| 验证失败: 缺少开始节点 | DSL中没有w_start | 添加w_start节点 |
| 验证失败: 节点类型无效 | 使用了不支持的类型 | 检查12种合法节点类型 |
| 测试执行超时 | 连接器响应慢 | 检查连接器配置 |
| 变量引用错误 | $nodeId.variable格式错误 | 检查节点ID和变量名 |
| API返回"未发布" | 工作流未发布 | 先执行publish命令 |
| 回滚失败 | 版本不存在 | 使用versions查看可用版本 |

## 最佳实践

### 设计阶段
1. 使用模板快速开始
2. 先用validate验证DSL正确性
3. 用dependency检查依赖是否满足

### 测试阶段
1. 使用test run进行功能测试
2. 测试各种参数组合
3. 验证错误处理路径

### 发布阶段
1. 使用语义化版本号 (v1.0.0, v1.1.0, v2.0.0)
2. 添加清晰的发布说明(memo)
3. 重要变更前先测试

### 运维阶段
1. 使用versions查看版本历史
2. 生产问题优先rollback而非热修复
3. 定期生成文档保持更新

## 参考文档

详细的节点配置和DSL规范请参考:
- `references/node-details.md` - 12种节点详细配置
- `references/dsl-specification.md` - DSL完整规范
- `references/api-examples.md` - API调用示例
- `references/error-codes.md` - 错误码说明
- `references/performance-tuning.md` - 性能优化指南
- `references/testing-guide.md` - 测试指南
